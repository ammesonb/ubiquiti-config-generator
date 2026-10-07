const { test } = require('node:test');
const assert = require('node:assert/strict');
const cleanup = require('./cleanup-vyos-images.cjs');
const hash = 'a'.repeat(64);
const now = Date.parse('2026-10-06T00:00:00Z');
const policy = {
  otherBefore: Date.parse('2026-09-29T00:00:00Z'),
  mainBefore: null,
};
const image = (id, tags, created_at = '2026-09-01T00:00:00Z') => ({
  id, created_at, metadata: { container: { tags } },
});
const context = ownerType => ({
  repo: { owner: 'owner', repo: 'Project' },
  payload: { repository: { owner: { type: ownerType } } },
});
const core = { info() {} };

function pages(...data) {
  return { async *iterator() { for (const items of data) yield { data: items }; } };
}

test('getImages yields every image across pages', async () => {
  const github = { paginate: pages([{ id: 1 }], [{ id: 2 }, { id: 3 }]) };
  const images = [];
  for await (const item of cleanup.getImages(github, '/route', {})) images.push(item.id);
  assert.deepEqual(images, [1, 2, 3]);
});

test('other-image expiry only selects old managed images', () => {
  assert.equal(cleanup.shouldDelete(image(1, [`inputs-${hash}`]), policy), true);
  assert.equal(cleanup.shouldDelete(image(2, [`inputs-${hash}`, `main-${hash}`]), policy), false);
  assert.equal(cleanup.shouldDeleteOther(image(3, [`inputs-${hash}`], '2026-10-01T00:00:00Z'), policy), false);
  assert.equal(cleanup.shouldDeleteOther(image(4, [`inputs-${hash}`], 'invalid'), policy), false);
  assert.equal(cleanup.shouldDeleteOther(image(5, ['unrelated-tag']), policy), false);
  assert.equal(cleanup.shouldDeleteOther(image(6, []), policy), false);
});

test('weekly main expiry preserves the current image and recent promotions', () => {
  const weekly = { ...policy, mainBefore: Date.parse('2026-09-06T00:00:00Z') };
  const old = image(1, [`inputs-${hash}`, `main-${hash}`]);
  assert.equal(cleanup.shouldDelete(old, weekly), true);
  assert.equal(cleanup.shouldDelete(image(2, [`main-${hash}`]), weekly), true);
  assert.equal(cleanup.shouldDelete(image(3, [`main-${hash}`, 'main-current']), weekly), false);
  assert.equal(cleanup.shouldDeleteMain({ ...old, updated_at: '2026-10-01T00:00:00Z' }, weekly), false);
  assert.equal(cleanup.shouldDeleteMain({ ...old, updated_at: 'invalid' }, weekly), false);
});

for (const ownerType of ['User', 'Organization']) {
  test(`cleanup handles ${ownerType} ownership and rechecks promotion`, async () => {
    const expired = image(42, [`inputs-${hash}`]);
    const promoted = image(43, [`inputs-${hash}`]);
    const main = image(44, [`main-${hash}`]);
    const calls = [];
    const github = {
      paginate: { async *iterator(route, params) {
        assert.ok(route.startsWith(ownerType === 'User' ? 'GET /users/' : 'GET /orgs/'));
        assert.equal(params.package_name, 'project/vyos-lab');
        yield { data: [expired, promoted, main] };
      } },
      request: async (route, params) => {
        calls.push([route, params.package_version_id]);
        return { data: params.package_version_id === 43 ? image(43, [`inputs-${hash}`, 'main-current']) : expired };
      },
    };
    await cleanup({ github, context: context(ownerType), core, now });
    assert.deepEqual(calls.filter(([route]) => route.startsWith('DELETE')).map(([, id]) => id), [42]);
  });
}

test('daily and weekly orchestration apply different main retention rules', async () => {
  for (const pruneMain of [false, true]) {
    const old = image(1, [`main-${hash}`], '2026-08-01T00:00:00Z');
    const current = image(2, [`main-${hash}`, 'main-current'], '2026-08-01T00:00:00Z');
    const deleted = [];
    const github = {
      paginate: pages([old, current]),
      request: async (route, params) => {
        if (route.startsWith('DELETE')) deleted.push(params.package_version_id);
        return { data: old };
      },
    };
    await cleanup({ github, context: context('User'), core, now, pruneMain });
    assert.deepEqual(deleted, pruneMain ? [1] : []);
  }
});

test('unexpected API failures are reported', async () => {
  const failure = Object.assign(new Error('permission denied'), { status: 403 });
  const github = { paginate: { async *iterator() { throw failure; } } };
  await assert.rejects(cleanup({ github, context: context('User'), core, now }), /permission denied/);
});
