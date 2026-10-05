# Project roadmap

Last updated: 2026-10-05

Build a configuration management system that lets users define network intent once, review and validate changes through GitHub, and deploy complete configurations to Ubiquiti devices with visible status and recovery.
Shared configurations and logical entities should reduce duplication across both settings and devices.

The [feature articles](features/README.md) describe configuration management, GitHub integration, and device deployment.
This roadmap records capability status and the milestones that connect those areas into usable workflows.

## Capability status

| Area | Present in the Go implementation | Next capability |
| --- | --- | --- |
| Development foundation | Dependency-free Go command, Make targets, formatting and lint configuration, and documentation. | Docker-based testing and portable CI. |
| Configuration management | Feature specifications. | Static-file deployment first; parsing, validation, and generation follow. |
| GitHub integration | App and repository boundaries documented. | Installation authentication, event handling, checks, and deployment reporting. |
| Device deployment | Feature specifications. | Native whole-file loading with verified errors and recovery. |
| Testing | Test strategy documented. | Unit and device integration tests, a lab fixture, and disposable CI provisioning. |

The existing Python implementation, examples, and tests remain in this repository.
The table tracks the new Go implementation only.
The `go_conversion` branch contains earlier parsers, abstractions, and GitHub integration code for selective reuse.
Those capabilities are not implemented in this scaffold.
Update this table as features land and link their implementation or acceptance evidence.

## Delivery milestones

### Foundation Portable development and CI

Establish a minimal implementation workspace with the project documentation, a Go entry point, and Docker-based test commands.
Keep the Python implementation and reference material in place, and reuse components as milestones require them.

Provide repeatable local unit tests and a device integration harness.
The harness must accept an external test endpoint as well as a Docker-provisioned lab, so local testing and CI use the same scenarios.
Verify the selected router image's load and recovery behavior before relying on it.

Use GitHub Actions to invoke the same portable commands for build and test.
Remove dependence on Buildkite agent binaries, host paths, and credentials from the new test setup.
Release packaging follows a usable deployment implementation; app hosting remains separate from CI.

**Done when:** a contributor can run the documented tests locally, and a standard CI job can run them without the maintainer's infrastructure.
Device tests must verify lab readiness and clean up after failure.

### 1 Deploy a native configuration

Feature: [Applying configuration](features/device-deployment.md#applying-configuration).

Make a checked-in, complete native configuration deployable to one supported test device.
This establishes the connection between desired configuration and actual device state before adding generation or schema dependencies.

The first workflow needs file transfer, native loading and committing, useful error reporting, and a verified recovery mechanism.
A manually invoked Go command is the proposed entry point.
Local edits must be testable without pushing a commit or relying on GitHub.

**Done when:** configuration A can be replaced with B, including deleted settings; active-state checks verify the result; failures stop subsequent operations; and confirmation and recovery behavior are demonstrated on the lab target.

### 2 Make deployments reviewable and observable

Features: [Preview and dry run](features/device-deployment.md#preview-and-dry-run), [status checks and recovery](features/device-deployment.md#status-checks-and-recovery), and [GitHub deployment history](features/github-integration.md#deployment-triggers-and-history).

Add a preview against live configuration and status reporting throughout execution.
Connect deployment outcomes to GitHub so a user can identify what was attempted, on which device, and whether it was confirmed or rolled back.

This follows the basic deployment path because meaningful progress and errors depend on real device operations.
Preview must use live state rather than assume the last repository revision describes the router.

**Done when:** a user can review additions, changes, and deletions before applying them, follow progress while work is running, and inspect the outcome associated with the deployed artifact and source revision when available.

### 3 Validate proposed configurations

Features: [Static validation](features/configuration.md#static-validation) and [GitHub checks](features/github-integration.md#checks-and-review).

Add cheap checks derived from the target device's definitions and publish them through GitHub checks.
Cover valid paths and structure, address formats, numeric ranges, allowed values, and supported string patterns.
Leave device-local validation to load and commit error reporting.

Schema acquisition, format compatibility, and unsupported constraints make this a separate capability from deployment.
Evaluate caching as part of this work.
An optional candidate dry load can provide another check where firmware supports it; its feasibility must not block static validation.

**Done when:** representative invalid inputs produce actionable check failures before deployment, supported checks report their coverage, and deferred device checks are not mistaken for successful validation.

### 4 Generate configuration from logical entities

Feature: [Logical entities](features/configuration.md#logical-entities).

Let users define hosts and networks with their related settings, such as a host's address, port forwards, and firewall access.
Generate native configuration from those entities and deploy it through the established path.

This reduces duplicated facts while keeping deployment independent of the authoring format.
Generation must account for deletions and conflicts with manually authored native settings.

**Done when:** changing an entity updates all of its generated settings consistently, and deleting it removes those settings from the next deployment.

### 5 Reuse configuration across devices

Features: [Modularization and reuse](features/configuration.md#modularization-and-reuse) and [multiple devices and smart deployment](features/device-deployment.md#multiple-devices-and-smart-deployment).

Compose shared modules and device-specific inputs into an effective configuration for each device.
Select targets based on their effective changes and report deployment state and history independently for each target.

This requires deterministic composition, explicit override and ownership rules, and handling of partial rollout outcomes.
Shared repository history cannot stand in for the state of every device.
Drift policy must be defined before unchanged desired input can reliably skip deployment.

**Done when:** multiple devices reuse configuration without copying it, affected targets are selected correctly, unnecessary deployments can be skipped under the chosen drift policy, and successes and failures remain visible per device.

## Testing throughout delivery

Device integration testing begins with the first milestone; parsing, validation, and composition coverage grows with those capabilities.
See [testing and development](testing.md) for the shared local and CI suite and acceptance scenarios.
Automated lab provisioning need not block testing against an available endpoint.

## Immediate implementation work

1. Review the Go scaffold, developer commands, and documentation.
2. Establish Docker-based testing and select a pinned router lab image, retaining support for an externally supplied endpoint.
3. Add portable GitHub Actions build and test workflows.
4. Implement and test static-file deployment, including native load, deletion, failure, and confirmation semantics.

The product is an installed [GitHub App](features/github-integration.md#github-app-and-repository-boundaries); configuration comes from a separate repository.
The first deployment command exercises the deployment mechanism that the app will invoke.

Keep detailed behavior and unresolved technical decisions in the [feature articles](features/README.md); update this roadmap when capability status or delivery order changes.
