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

On Windows, the converter can run in a disposable Linux container without installing its Linux dependencies on the host.
After `make test-docker`, put the verified ISO and converter named `iso-to-oci` in `.cache/lab`, then run this from the repository root in PowerShell:

```powershell
$labDir = (Resolve-Path .cache/lab).Path
docker run --rm --mount "type=bind,source=$labDir,target=/lab" --workdir /lab ubq-go-tests:local bash -c 'apt-get update -qq && apt-get install -y -qq xorriso squashfs-tools jq xz-utils >/dev/null && bash /lab/iso-to-oci /lab/vyos-2026.09.30-1921-rolling-generic-amd64.iso'
docker import --platform=linux/amd64 .cache/lab/vyos-2026.09.30-1921-rolling-oci-amd64.tar.xz ubq-vyos:2026.09.30-1921-rolling
docker image inspect ubq-vyos:2026.09.30-1921-rolling --format "{{.Id}}"
```

The mount contains only lab preparation artifacts; `.cache` and `.env.integration` remain ignored by Git.
Set `UBQ_VYOS_IMAGE` to the inspected ID and run `make smoke-docker`.

## CI follow-up

GitHub Actions will invoke the same locally verified test commands.
A CI lab job should reuse a prepared, pinned router image and retain diagnostic output when a smoke test fails.
No Go CI workflow is added at this stage.
The existing Python tests and historical CI configuration remain separate.

## Future deployment scenarios

- Deploy configuration A, then B with additions, changes, and deletions; verify active state matches B.
- Fail upload or load and verify no subsequent commit occurs.
- Reject invalid configuration and verify the reported error and resulting active state.
- Confirm a deployment and verify it persists according to policy.
- Let confirmation expire and verify restoration of the previous configuration.
- Keep progress queries responsive during a deployment.
- Modify a shared module and verify affected-device selection and independent outcomes.
