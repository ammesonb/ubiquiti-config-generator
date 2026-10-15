const DAY = 24 * 60 * 60 * 1000;
const tagsOf = image => image.metadata?.container?.tags ?? [];
const mainDate = image => Date.parse(image.updated_at ?? image.created_at);

function isMain(image) {
  return tagsOf(image).some(tag => tag === 'main-current' || /^main-[0-9a-f]{64}$/.test(tag));
}

function shouldDeleteMain(image, policy) {
  if (policy.mainBefore === null) return false;
  if (tagsOf(image).includes('main-current')) return false;
  const updated = mainDate(image);
  return Number.isFinite(updated) && updated < policy.mainBefore;
}

function shouldDeleteOther(image, policy) {
  const managed = tagsOf(image).some(tag => /^inputs-[0-9a-f]{64}$/.test(tag));
  const created = Date.parse(image.created_at);
  return managed && Number.isFinite(created) && created < policy.otherBefore;
}

function shouldDelete(image, policy) {
  return isMain(image) ? shouldDeleteMain(image, policy) : shouldDeleteOther(image, policy);
}

async function* getImages(github, route, params) {
  for await (const page of github.paginate.iterator(`GET ${route}`, { ...params, per_page: 100, state: 'active' })) {
    yield* page.data;
  }
}

async function cleanup({ github, context, core, now = Date.now(), pruneMain = false }) {
  const owner = context.repo.owner;
  const packageName = `${context.repo.repo.toLowerCase()}/vyos-lab`;
  const namespace = context.payload.repository.owner.type === 'Organization' ? 'orgs' : 'users';
  const route = `/${namespace}/{owner}/packages/container/{package_name}/versions`;
  const params = { owner, package_name: packageName };
  const images = [];
  try {
    for await (const image of getImages(github, route, params)) images.push(image);
  } catch (error) {
    if (error.status === 404) {
      core.info('No accessible VyOS package to clean up.');
      return;
    }
    throw error;
  }
  const policy = {
    otherBefore: now - 7 * DAY,
    mainBefore: pruneMain ? now - 30 * DAY : null,
  };
  for (const image of images) {
    if (!shouldDelete(image, policy)) continue;
    const endpoint = `${route}/{package_version_id}`;
    const args = { ...params, package_version_id: image.id };
    // Recheck the live tags before deletion, including main promotion.
    const { data: current } = await github.request(`GET ${endpoint}`, args);
    if (!shouldDelete(current, policy)) continue;
    await github.request(`DELETE ${endpoint}`, args);
    core.info(`Deleted expired image version ${image.id}.`);
  }
}

module.exports = cleanup;
Object.assign(module.exports, { getImages, isMain, shouldDelete, shouldDeleteMain, shouldDeleteOther });
