# Testing

[Documentation](README.md) | [Development commands](development.md) | [Roadmap](roadmap.md)

Use ordinary tests for application and transport logic, and a real VyOS lab for firmware behavior.
An in-process SSH server can exercise connection handling but cannot establish how VyOS loads, commits, or restores configuration.
Tests that modify device state run only on disposable Docker routers owned by the test process.
The external-endpoint mode is read-only.

## Required technologies

| Component | Purpose and configuration |
| --- | --- |
| Go and GNU Make | Run the test harness and shared local/CI commands; Go is selected by `go.mod`. |
| `golang.org/x/crypto/ssh` | SSH transport with explicit authentication and trusted host identity; dependency versions are in `go.mod` and `go.sum`. |
| Docker with Linux containers | Run ordinary Linux tests and disposable VyOS routers; router provisioning requires a local engine. |
| Native VyOS OCI filesystem | Provide `/sbin/init`, systemd, SSH, configuration tools, commit history, and native rollback timers. |
| ISO conversion tools | Installed in `Dockerfile.lab`, rather than on the host; the official converter prepares the router filesystem. |
| Node.js 18 or later | Run registry-cleanup tests without accessing GitHub. |
| actionlint and ShellCheck | Validate workflows and their shell scripts; see [tool installation](development.md#commands). |
| GitHub Actions and GHCR | Run the same Make targets on hosted Linux runners and distribute the prepared router image. |

Docker Desktop must use its Linux engine; on Windows this requires working WSL and virtualization support.
Ordinary native Go tests do not require Docker or a router.
Go's race detector also requires a working C compiler on platforms where it uses cgo.

## Ordinary tests

```text
make test
make test-race
make test-docker
```

The Go tests use an in-process SSH server for controlled transport responses and failures.
They do not emulate VyOS configuration semantics.
`make test-docker` runs the same ordinary suite in Linux using the toolchain in `Dockerfile.test`.
The Docker build context excludes local environment files and credentials.

`make check` runs build, static checks, module consistency, and ordinary tests without connecting to a device.
`make registry-tests` uses a mocked GitHub client; it performs no package deletion or other remote operation.

## External lab endpoint

Copy `.env.integration.example` to `.env.integration` in the repository root and configure a dedicated lab endpoint.
Run the read-only check with:

```text
make smoke
```

| Variable | Configuration |
| --- | --- |
| `UBQ_TEST_HOST` | Required hostname or IP address. |
| `UBQ_TEST_PORT` | SSH port, default 22. |
| `UBQ_TEST_USER` | Required SSH username. |
| `UBQ_TEST_SSH_KEY` | Private key file; mutually exclusive with password authentication. |
| `UBQ_TEST_KEY_PASSPHRASE` | Optional private key passphrase. |
| `UBQ_TEST_PASSWORD` | Password; mutually exclusive with a private key. |
| `UBQ_TEST_KNOWN_HOSTS` | Required trusted OpenSSH known-hosts file. |
| `UBQ_TEST_EXPECT_HOSTNAME` | Required expected active hostname. |
| `UBQ_TEST_TIMEOUT` | Readiness deadline, default 2m and maximum 3m. |

Process environment values override dotenv values, including explicitly empty values.
The dotenv file is optional when all settings are supplied through the environment.
Credential paths are relative to the dotenv file's directory unless absolute.

Obtain the host key through a trusted console or provisioning mechanism; the harness does not accept unknown keys automatically.
Missing or invalid settings fail an explicitly requested run.
SSH operations have deadlines, and errors do not print credentials or full configuration contents.

## Disposable Docker router

Configure `UBQ_VYOS_IMAGE` with a complete immutable local image ID (`sha256:...`) in `.env.integration` or the process environment.
Docker mode needs only this setting and ignores external endpoint credentials.
Prepare the image as described under [the lab baseline](#verified-lab-baseline), then run the device integration suite:

```text
make integration-docker
```

`make smoke-docker` is available for a shorter read-only Docker check.
Docker test targets use an already prepared image; they do not download firmware or pull mutable image tags.

The [Docker harness](../integration/docker_test.go) creates a uniquely named IPv6-enabled network and a privileged router container running `/sbin/init`.
It mounts `/lib/modules` read-only and uses temporary filesystems for `/run` and `/run/lock`.
SSH is published only on host loopback with a dynamically assigned port.
A remote Docker engine cannot provide this local endpoint; provision it separately and use the external read-only mode instead.

Each router receives the [boot fixture](../integration/lab/config.boot.tmpl) before startup and a newly generated temporary SSH identity.
The fixture enables key-only SSH, native commit history, and reload-based confirmed-commit recovery.
Docker and VyOS use the same initial hostname so native commands can resolve the local host.
The harness reads the router's Ed25519 public host key through Docker and restricts SSH negotiation to that trusted key type.
After a container restart it obtains the current port binding again.

Readiness is established before native operations begin.
Each test owns its container and network and removes them on completion or ordinary failure.
A killed test process can leave resources behind; identify its containers and networks by the `ubq.test=integration` label.

## Device integration tests

Run all device integration tests, or select groups with `TAGS`:

```text
make integration-docker
make integration-docker TAGS=deploy
make integration-docker TAGS="deploy,smoke"
```

Omitting `TAGS` runs all integration tests.
`TAGS` accepts comma- or space-separated group names and runs groups matching any supplied name.
Independent groups run in parallel using separate routers and networks.
Stateful deployment operations remain sequential within their router, and dependent work stops when a prerequisite fails.
The application command and GitHub App are not involved in this test layer.

Complete native configuration fixtures are checked in under [integration/lab](../integration/lab/).
The harness substitutes only the temporary SSH public key and the firmware version footer produced by native `save`.
That footer lets the firmware interpret the file using its current configuration version.
Device behavior is evaluated through live active configuration, saved configuration, fresh SSH connections, and native timer state.
VyOS validation and rollback are executed on the router rather than reimplemented or mocked.
Recovery waits for real timer expiry; allow up to five minutes for a full run, with a ten-minute overall limit.

The [native driver](../integration/lab/deploy.vbash) keeps loading, confirmed committing, confirmation, and saving separate so their effects can be observed independently.
Native exit status alone is insufficient to establish commit success on the recorded firmware; see [device behavior and recovery](features/device-deployment.md#native-vyos-deployment-integration-tests).
The test source defines scenario coverage, and the [roadmap](roadmap.md) records implementation progress.

## Verified lab baseline

[baseline.env](../integration/lab/baseline.env) is the source of the pinned firmware inputs:

| Setting | Meaning |
| --- | --- |
| `VYOS_VERSION` | Exact official rolling release tag for the generic AMD64 ISO. |
| `VYOS_ISO_SHA256` | Published ISO checksum, as 64 hex characters. |
| `VYOS_CONVERTER_REVISION` | Full commit SHA of the official `vyos-build` conversion helper. |

VyOS documents ISO conversion in its [Docker installation instructions](https://docs.vyos.io/en/rolling/installation/virtual/docker.html).
The filesystem must contain real VyOS services and configuration tools; an arbitrary Linux SSH image cannot establish firmware compatibility.
Prepare and inspect the pinned image with:

```text
make lab-image
docker image inspect ubq-vyos:lab --format "{{.Id}}"
```

Set `UBQ_VYOS_IMAGE` to that inspected ID and run `make integration-docker`.
The `ubq-vyos:lab` tag is a convenience name; tests use its immutable local ID.
`make lab-image` downloads and verifies the ISO, runs the pinned converter inside the builder container, and imports the resulting filesystem.
The builder installs the necessary Linux tools, so the host needs Docker and Make rather than its own conversion toolchain.
Verified conversion output, checksums, and input metadata are cached in `.cache/lab`.
Changed baseline inputs or failed archive verification trigger conversion again.
The cache and `.env.integration` are ignored by Git.

### Updating the VyOS baseline

1. Select an [official release](https://github.com/vyos/vyos-nightly-build/releases) and its `vyos-<release-tag>-generic-amd64.iso` asset.
2. Update `VYOS_VERSION` and the asset's published `VYOS_ISO_SHA256` in `integration/lab/baseline.env`, without the `sha256:` checksum prefix.
3. Keep the converter revision unless an update is needed; use a full commit SHA from [vyos-build](https://github.com/vyos/vyos-build) when changing it.
4. Run `make lab-image`, inspect the imported image ID, update `UBQ_VYOS_IMAGE`, and run `make integration-docker`.

The changed inputs produce a new GHCR lookup hash automatically.
No cache key or registry tag needs a manual edit.

## GitHub Actions

[Go CI](../.github/workflows/go.yml) uses the same Make commands as local development for ordinary tests and static checks.
Go is selected from `go.mod`, tool versions are pinned in the workflow, and `setup-go` caches modules and build outputs.
The router job calls the reusable [VyOS lab workflow](../.github/workflows/vyos-lab.yml) on `ubuntu-24.04`.
It provisions its own routers and invokes `make integration-docker` once, using no production credentials or maintainer-hosted runner.
The workflows can also be invoked manually.
Build and router jobs have 15-minute and 20-minute limits respectively; failed lab runs retain diagnostic logs for seven days.

### Registry image reuse

The lab image is published under `ghcr.io/<owner>/<repository>/vyos-lab`.
The lookup tag `inputs-<hash>` derives from the firmware baseline, builder Dockerfile, conversion script, and Makefile.
Fixtures and injected configuration are excluded so configuration changes reuse the prepared image.
CI resolves an available tag to its registry digest, pulls by that digest, and passes the inspected local image ID to the harness.
A registry digest and a local image ID identify different objects.
When no accessible image exists, CI prepares the pinned baseline locally.

Successful runs with registry write access can publish after the integration suite passes.
Publishing uses the built-in `GITHUB_TOKEN` with `packages: write`; the source-repository label links the package to the repository.
`LAB_IMAGE_SOURCE` overrides that label when preparing images for a fork.
Fork pull requests have no publishing permission and fall back to local preparation if the image is inaccessible.
GHCR packages initially default to private; [make the package public](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility) to allow anonymous contributor pulls.

For local reuse, copy the digest reference from the workflow summary:

```text
docker pull ghcr.io/OWNER/REPOSITORY/vyos-lab@sha256:DIGEST
docker image inspect ghcr.io/OWNER/REPOSITORY/vyos-lab@sha256:DIGEST --format "{{.Id}}"
```

Set `UBQ_VYOS_IMAGE` to the inspected ID and run `make integration-docker`.
Images contain no test credentials or injected configuration.

### Image retention

The [cleanup workflow](../.github/workflows/vyos-cleanup.yml) expires non-main images daily after seven days and older main images weekly after 30 days.
Non-main age is measured from creation; main age uses the last metadata update, or creation when no update time exists.
Successful main publication adds `main-<hash>` and moves `main-current` to the verified image; `main-current` is always retained.
Manual cleanup applies both policies, and only versions carrying the workflow's managed tags are eligible.

The cleanup token needs administrator access to the linked package.
Publication and cleanup share a concurrency group, which serializes lab workflow runs and prevents deletion during main promotion.
Cleanup rechecks current tags before deletion.
