# Testing and development

[Documentation](README.md) | [Roadmap](roadmap.md)

This document describes the test strategy and planned integration environment.
See the [capability status](roadmap.md#capability-status) for implementation progress.

Use ordinary tests for parsing, supported validation, composition, and failure handling.
Use integration tests against a dedicated router instance for native loading, active-state verification, and recovery.
Configuration-replacing tests must not target production devices.

## Local and CI environments

Keep test commands independent of the CI provider.
Docker-based tooling should run locally and on GitHub Actions without Buildkite agent mounts, maintainer-specific host paths, or CI credentials.
The Go test environment and the router lab are separate components; tests can also connect to an externally supplied router endpoint.

The integration suite accepts a test endpoint and credentials.
Local developers can supply these through an ignored dotenv file and run against a persistent lab VM or container, including with uncommitted code.
Commit an example configuration; the exact variable names and dotenv loader are implementation choices.

GitHub Actions provisions a disposable lab, supplies the same endpoint settings, and runs the same suite.
Provisioning includes readiness checks, diagnostic collection, and cleanup after failures.
Pin the lab image and avoid rebuilding the router OS on every test run.
Starting with a manually supplied endpoint keeps automated provisioning from blocking deployment development.

VyOS is a candidate lab target.
Containers require suitable Linux privileges and kernel facilities; VMs provide their own kernel and support reboot and persistence tests.
Prove the selected image works on the selected runner before depending on it.
VyOS tests establish behavior for that firmware, not compatibility with every Ubiquiti device.

## Acceptance scenarios

- Deploy configuration A, then B with additions, changes, and deletions; verify active state matches B.
- Fail upload or load and verify no subsequent commit occurs.
- Reject invalid configuration and verify the reported error and resulting active state.
- Confirm a deployment and verify it persists according to policy.
- Let confirmation expire and verify restoration of the previous configuration.
- Keep progress queries responsive during a deployment.
- Later, modify a shared module and verify affected-device selection and independent outcomes.

## Open decisions

- Which pinned lab image and local runtime should be used?
- Which CI runner can support the image's required privileges and kernel facilities?
- What endpoint settings and dotenv loading command should the integration suite expose?

## Provisioning references

[Docker installation](https://docs.vyos.io/en/latest/installation/virtual/docker.html), [OCI image support](https://blog.vyos.io/vyos-project-september-2026-update), [VM installation](https://docs.vyos.io/en/rolling/installation/virtual/libvirt.html), and [GitHub runner capabilities](https://docs.github.com/en/actions/reference/runners/github-hosted-runners) inform lab provisioning.
Pin upstream versions when implementing against them.
