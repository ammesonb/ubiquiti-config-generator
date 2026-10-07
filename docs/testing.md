# Testing and development

[Documentation](README.md) | [Roadmap](roadmap.md)

The Go test suite has two levels: ordinary tests that need no device, and opt-in smoke tests against a dedicated VyOS lab.
The smoke test authenticates SSH, reads active configuration, and verifies the expected hostname.
It never loads or commits configuration.
Deployment, confirmation, and rollback tests will be added with the deployment mechanism.

## Ordinary tests

```text
make test
make test-race
make test-docker
```

`make test` exercises dotenv configuration, credential selection, host identity verification, command failures, configuration checks, and connection deadlines using a local in-process SSH server.
That server tests the harness; it does not emulate VyOS or establish firmware compatibility.
The race target also requires a working C compiler on platforms where Go's race detector uses cgo.

`make test-docker` builds `Dockerfile.test` with the pinned Go toolchain and runs the same ordinary suite in Linux.
It copies only Go source, the lab fixture, and module metadata; local environment files and credentials are excluded from the build context.
It requires no Buildkite agent, host-specific mounts, or CI credentials.
`make check` includes ordinary tests and statically checks the integration code without connecting to a router.

## External lab endpoint

Copy `.env.integration.example` to `.env.integration` in the repository root and provide the dedicated router's host, SSH port, username, authentication, trusted known-hosts file, and expected hostname.
Then run:

```text
make smoke
```

The dotenv file is optional when all settings are supplied in the process environment.
Process values take precedence, including explicitly empty values.
Credential file paths are relative to the dotenv file's directory unless absolute.
The settings are:

| Variable | Meaning |
| --- | --- |
| `UBQ_TEST_HOST` | Required endpoint hostname or IP address |
| `UBQ_TEST_PORT` | SSH port, default 22 |
| `UBQ_TEST_USER` | Required SSH username |
| `UBQ_TEST_SSH_KEY` | Private key file; use instead of a password |
| `UBQ_TEST_KEY_PASSPHRASE` | Optional private key passphrase |
| `UBQ_TEST_PASSWORD` | Password; use instead of a private key |
| `UBQ_TEST_KNOWN_HOSTS` | Required trusted OpenSSH known-hosts file |
| `UBQ_TEST_EXPECT_HOSTNAME` | Required expected active hostname |
| `UBQ_TEST_TIMEOUT` | Readiness deadline, default 2m and maximum 3m |

Obtain the host key through a trusted console or the lab provisioning mechanism.
The test does not automatically accept unknown SSH identities.
Missing settings fail an explicit smoke run rather than silently skipping it.
Independent configuration errors are reported together, including required settings, port and timeout validation, credential selection, and credential-file errors.
Failures report stages without printing the router's full configuration or credentials.

## Disposable Docker router

`make smoke-docker` uses a locally imported VyOS OCI image and the fixture in `integration/lab/config.boot.tmpl`.
The host must have Go and a local Docker engine using Linux containers.
This mode publishes SSH only on host loopback; a remote Docker engine should instead be provisioned separately and tested through the external-endpoint mode.

VyOS documents conversion of an ISO into an OCI filesystem using its `iso-to-oci` helper.
Obtain a specific firmware ISO and verify its published checksum, then follow the [official conversion and import instructions](https://docs.vyos.io/en/rolling/installation/virtual/docker.html).
Record the firmware and the source revision of the conversion helper when preparing an image.
The resulting image must contain `/sbin/init` and the VyOS configuration tools; an arbitrary Linux SSH container is not a substitute.

After importing, find its immutable local image ID:

```text
docker image inspect YOUR_IMPORTED_IMAGE --format "{{.Id}}"
```

Put that complete `sha256:...` value in `UBQ_VYOS_IMAGE` in `.env.integration` or the process environment, then run:

```text
make smoke-docker
```

This mode needs only `UBQ_VYOS_IMAGE`; it does not use external-endpoint credentials or settings.
It does not pull unpinned images or automatically download firmware.
The test creates a uniquely named IPv6-enabled network and a privileged router container, generates a temporary SSH identity, and copies the fixture before boot.
The fixture enables key-only SSH and sets `ubq-test-router` as the hostname, retaining Docker's management interface configuration.
The image must be prepared for container operation as described by VyOS.

Readiness checks wait for the router's SSH host key and then for a readable active configuration.
The host key is retrieved through the Docker engine before the SSH connection is trusted.
The same read-only smoke assertion runs for Docker and external endpoints.
On completion or ordinary failure, the test removes its own container and network; startup logs are included on failure.
An interrupted or killed test process may require manual cleanup of resources bearing the `ubq.test=smoke` label.

## Verified lab baseline

The ordinary test container and disposable-router smoke test pass on Windows 10 with Docker Desktop using Linux containers.
The router test verifies SSH authentication, host identity, the active fixture hostname, and container and network cleanup.
It does not yet establish deployment, confirmation, or rollback behavior.

| Input | Tested value |
| --- | --- |
| Firmware | `2026.09.30-1921-rolling`, generic AMD64 |
| ISO SHA256 | `df10cf1ae7d69c50078a402fb0a404f73091ea567692fe5f955714492ea925c0` |
| `vyos-build` converter revision | `414b5ae6d9f568d218880b1a3c4293f7ef193df7` |

Use the ISO from the [official release](https://github.com/vyos/vyos-nightly-build/releases/tag/2026.09.30-1921-rolling) and the [converter at the tested revision](https://github.com/vyos/vyos-build/blob/414b5ae6d9f568d218880b1a3c4293f7ef193df7/scripts/iso-to-oci).
Verify the ISO checksum before conversion.
Image IDs are local import results; obtain yours with the inspection command above.
The harness requests the Ed25519 host-key algorithm whose public key it retrieves through Docker, since VyOS also offers other host-key types.

The pinned firmware and converter inputs are declared in [baseline.env](../integration/lab/baseline.env).
Prepare the same image used by CI from PowerShell or a Linux shell:

```text
make lab-image
docker image inspect ubq-vyos:lab --format "{{.Id}}"
```

Set `UBQ_VYOS_IMAGE` in `.env.integration` to the inspected ID, then run `make smoke-docker`.
`make lab-image` downloads the pinned ISO when needed, verifies its SHA256, and converts it with the official helper inside a disposable Linux container.
It requires Docker only; Linux conversion tools are installed in the builder image rather than on the host.
It stores the converted filesystem, checksum, and baseline metadata in `.cache/lab` and reuses them when verification succeeds.
A changed baseline or failed archive verification rebuilds the filesystem.
The imported `ubq-vyos:lab` tag is a convenience name; smoke tests always use its immutable image ID.
Lab preparation artifacts and `.env.integration` remain ignored by Git.

### Updating the VyOS baseline

1. Choose a release from the [official VyOS builds](https://github.com/vyos/vyos-nightly-build/releases) and its `vyos-<release-tag>-generic-amd64.iso` asset.
2. In [baseline.env](../integration/lab/baseline.env), set `VYOS_VERSION` to that exact release tag and `VYOS_ISO_SHA256` to the ISO's published SHA256, using 64 hex characters without the `sha256:` prefix.
3. Keep `VYOS_CONVERTER_REVISION` unless the new firmware needs a converter update; if changing it, use a full commit SHA from [vyos-build](https://github.com/vyos/vyos-build).
4. Rebuild and inspect the new local image:

```text
make lab-image
docker image inspect ubq-vyos:lab --format "{{.Id}}"
```

Set `UBQ_VYOS_IMAGE` in `.env.integration` to that new image ID, then run `make smoke-docker`.
After it passes, update the tested-value table and release links above.
The baseline change automatically produces a new GHCR lookup hash; there is no cache or registry tag to change manually.

## GitHub Actions

[Go CI](../.github/workflows/go.yml) runs on pull requests, pushes to any branch, and manual invocation.
The build job runs `make registry-tests` and `make workflow-lint`, then `make check`, `make test-race`, and `make test-docker`.
Go is selected from `go.mod`; golangci-lint and actionlint are installed at pinned versions.
Go modules and build outputs are cached by `setup-go`.
The existing Python tests and historical CI configuration remain separate.

The router job calls the reusable [VyOS lab workflow](../.github/workflows/vyos-lab.yml), which can also be invoked manually.
It uses standard `ubuntu-24.04` runners and creates its own disposable router without production credentials.
The build and router jobs have 15-minute and 20-minute time limits respectively.
Failed router preparation or smoke runs retain diagnostic logs for seven days.

### Registry image reuse

The lab workflow uses `ghcr.io/<owner>/<repository>/vyos-lab` in the repository owner's namespace.
Its lookup tag is `inputs-<hash>`, derived from the firmware baseline, builder Dockerfile, conversion script, and Makefile.
Test fixtures and injected configuration are excluded from that hash so different configurations reuse the same image.

When the tagged image is available, the workflow resolves its registry digest and uses that digest to pull the image for the run.
It supplies the resulting local image ID to `make smoke-docker`.
This skips conversion, builder creation, and filesystem import on subsequent runs.
The registry digest and local image ID identify different objects; the harness still accepts only the local ID.

When the image is unavailable, the workflow runs `make lab-image` and the smoke test locally.
Every successful run with registry write access can publish the verified image using the built-in `GITHUB_TOKEN` with `packages: write`.
This includes branch pushes, pull requests within the repository, and manual lab runs with publishing enabled.
Fork pull requests cannot publish because their tokens are read-only; they reuse accessible images or build locally.
When an image already exists, publication reuses its content instead of rebuilding it.
A successful `main` run applies `main-<hash>` and `main-current` tags to the same image version, marking it as the latest verified main baseline even if it originated on a branch.

The image's source label connects it to the workflow repository.
No personal access token or configured repository secret is required.

New GHCR packages are private by default.
After the first publication, make the package public through its settings to permit anonymous pulls by contributors and fork pull requests, following [GitHub's visibility instructions](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility).
A fork pull request without access to the image falls back to local preparation.
The workflow summary records the published or reused digest.

### Image retention

The [cleanup workflow](../.github/workflows/vyos-cleanup.yml) runs daily for non-main images and weekly for older main images.
Daily runs expire managed non-main versions seven days after creation; reuse alone does not reset that period.
Weekly runs also expire older main versions 30 days after their last metadata update, falling back to creation time when no update time is available.
The image marked `main-current` is always retained, regardless of age.
Manual cleanup runs apply both policies.
Only versions with this workflow's tag scheme are eligible for deletion.

Cleanup reads all pages through `getImages`, then routes each image through `isMain` to `shouldDeleteMain` or `shouldDeleteOther`.
It rechecks the live tags immediately before deletion.
Publication and cleanup share a concurrency group so cleanup cannot delete an image during main promotion.
This serializes lab workflow runs within the repository.

Cleanup uses the repository's built-in token and requires administrator access to its linked GHCR package.
`make registry-tests` verifies the deletion rules and main-promotion protection without accessing GitHub.
Actual GHCR deletion permissions remain part of hosted verification.

To reuse a published image locally, copy its full digest reference from the workflow summary:

```text
docker pull ghcr.io/OWNER/REPOSITORY/vyos-lab@sha256:DIGEST
docker image inspect ghcr.io/OWNER/REPOSITORY/vyos-lab@sha256:DIGEST --format "{{.Id}}"
```

Set `UBQ_VYOS_IMAGE` in `.env.integration` to the inspected local ID and run `make smoke-docker`.
The base image contains no test credentials or router fixture; those are supplied separately for each smoke run.

Standard hosted runners are free for public repositories; private repositories use the owner's plan allowance, as described in [GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions).
GHCR container-image storage and bandwidth are currently free, as described in [GitHub Packages billing](https://docs.github.com/en/billing/concepts/product-billing/github-packages).
The workflow syntax and underlying commands can be verified locally before pushing.
A successful hosted run and first GHCR publication are still required to establish registry permissions and GitHub runner compatibility.

## Future deployment scenarios

- Deploy configuration A, then B with additions, changes, and deletions; verify active state matches B.
- Fail upload or load and verify no subsequent commit occurs.
- Reject invalid configuration and verify the reported error and resulting active state.
- Confirm a deployment and verify it persists according to policy.
- Let confirmation expire and verify restoration of the previous configuration.
- Keep progress queries responsive during a deployment.
- Modify a shared module and verify affected-device selection and independent outcomes.
