# Project roadmap

Last updated: 2026-10-06

Build a configuration management system that lets users define network intent once, review and validate changes through GitHub, and deploy complete configurations to Ubiquiti devices with visible status and recovery.
Shared configurations and logical entities should reduce duplication across both settings and devices.

The product is an installed [GitHub App](features/github-integration.md#github-app-and-repository-boundaries); device configurations come from separate repositories.
The first deployment command exercises the mechanism that the app will invoke.
The [feature articles](features/README.md) describe intended behavior and open design choices.

## Progress at a glance

Each milestone has one required-features checklist below, which records its implementation progress.
Checked items are implemented and verified to the scope stated; unchecked items still need work.
A milestone is complete when all its required features are checked.
Work proceeds through milestones in order, starting with the first unchecked requirement.

This roadmap tracks the current Go implementation.
The existing Python implementation, examples, and tests remain available as reference.
The `go_conversion` branch contains earlier parsers, abstractions, and GitHub integration code for selective reuse; those capabilities are not implemented here.

## Delivery milestones

### Foundation: portable development and CI

Establish a repeatable development and test workflow that contributors can run locally and in CI without the maintainer's infrastructure.
Keep the Python implementation and reference material available for selective reuse.
The lab must establish readiness and cleanup before deployment tests depend on it.
Configuration loading and recovery belong to milestone 1.
Release packaging follows a usable deployment implementation; app hosting remains separate from CI.

#### Required features

- [x] Go entry point and module metadata: [command](../cmd/ubiquiti-config-generator/main.go) and [go.mod](../go.mod).
- [x] Portable build, format, lint, vet, and module checks: [Makefile](../Makefile) and [development commands](development.md#commands).
- [x] Ordinary SSH harness tests covering deadlines, authentication settings, host identity, and command failures: [test helpers and tests](../internal/testlab/).
- [x] Ordinary tests verified locally, with the race detector, and in Linux Docker: [test container](../Dockerfile.test) and [test commands](testing.md#ordinary-tests).
- [x] External lab endpoint configuration implemented and unit-tested: [endpoint settings](testing.md#external-lab-endpoint).
- [x] Disposable VyOS smoke test verified against the recorded firmware, including SSH, active hostname, and cleanup after success and failure: [integration tests](../integration/) and [verified baseline](testing.md#verified-lab-baseline).
- [ ] GitHub Actions runs the build, static checks, and ordinary tests using the same commands: [workflow implemented](../.github/workflows/go.yml), awaiting hosted verification.
- [ ] A GitHub Actions lab job provisions the recorded firmware and passes the router smoke test without maintainer infrastructure, reusing verified images from GHCR: [reusable workflow implemented](../.github/workflows/vyos-lab.yml), awaiting first publication and hosted verification.

### 1 Deploy a native configuration

Feature: [Applying configuration](features/device-deployment.md#applying-configuration).

Make a checked-in, complete native configuration deployable to one supported test device.
This connects desired configuration to actual device state before adding generation or schema dependencies.
A manually invoked Go command provides local iteration without a push or GitHub event.

#### Required features

- [ ] Transfer a complete native configuration file to the target device.
- [ ] Load and commit through native device commands, replacing existing configuration including deleted settings.
- [ ] Report upload, load, and commit failures and stop subsequent operations.
- [ ] Support confirmation and timeout-based rollback, with an explicit save policy.
- [ ] Verify replacement, deletion, failure handling, confirmation, and rollback on the lab target.

### 2 Make deployments reviewable and observable

Features: [Preview and dry run](features/device-deployment.md#preview-and-dry-run), [status checks and recovery](features/device-deployment.md#status-checks-and-recovery), and [GitHub deployment history](features/github-integration.md#deployment-triggers-and-history).

Meaningful previews, progress, and error reporting depend on real device operations, so this follows the basic deployment path.
Live device state provides the comparison baseline; repository history alone cannot describe drift.

#### Required features

- [ ] Preview additions, changes, and deletions against the target's live configuration before applying them.
- [ ] Expose execution progress while device operations run, keeping status queries responsive.
- [ ] Report whether a deployment was applied, confirmed, failed, or rolled back.
- [ ] Connect deployment outcomes to the device, deployed artifact, and source revision when available through GitHub history.

### 3 Validate proposed configurations

Features: [Static validation](features/configuration.md#static-validation) and [GitHub checks](features/github-integration.md#checks-and-review).

Schema acquisition, format compatibility, and unsupported constraints make validation a separate capability from deployment.
Cheap checks catch common errors before deployment; device-local checks remain part of native load and commit error reporting.
Evaluate caching and optional candidate dry loads here, without making them prerequisites for static validation.

#### Required features

- [ ] Read target-device definitions in the supported node and XML formats.
- [ ] Check valid paths and structure, address formats, numeric ranges, allowed values, and supported string patterns.
- [ ] Report actionable failures through GitHub checks before deployment.
- [ ] State validation coverage and deferred device checks explicitly.
- [ ] Verify representative valid and invalid inputs against the supported checks.

### 4 Generate configuration from logical entities

Feature: [Logical entities](features/configuration.md#logical-entities).

Group related settings around hosts and networks to reduce duplicated facts.
Keep generation connected to the established native deployment path so the authoring format can evolve independently.
Ownership rules are needed where generated and manually authored settings meet.

#### Required features

- [ ] Define host and network entities with related settings such as addresses, port forwards, and firewall access.
- [ ] Generate native configuration from those entities.
- [ ] Define ownership and conflict handling for generated and manually authored settings.
- [ ] Verify that entity changes update all related settings and entity deletion removes them from the next deployment.

### 5 Reuse configuration across devices

Features: [Modularization and reuse](features/configuration.md#modularization-and-reuse) and [multiple devices and smart deployment](features/device-deployment.md#multiple-devices-and-smart-deployment).

Shared modules require deterministic composition, explicit overrides, and handling of partial rollout outcomes.
Shared repository history cannot stand in for the state of every device.
Drift policy must be defined before unchanged desired input can reliably skip deployment.

#### Required features

- [ ] Compose shared modules and device-specific inputs into effective configuration for each device.
- [ ] Define deterministic merge, override, and ownership rules.
- [ ] Select affected targets based on effective configuration changes and the chosen drift policy.
- [ ] Track execution state and deployment history independently for each device, including partial failures.
- [ ] Verify reuse without copying, correct target selection, and independent outcomes across multiple devices.

## Maintaining progress

Update a requirement's checkbox when its stated behavior is implemented and verified, linking the relevant code or acceptance evidence.
Record implemented but unverified work beside its unchecked requirement.
Keep detailed behavior and unresolved technical decisions in the feature articles.
The checklists are the progress record; avoid separate delivered and remaining lists.

Tests grow with each milestone's capabilities.
See [testing and development](testing.md) for local and CI commands and acceptance scenarios.
