# Device deployment

[Feature index](README.md) | [Roadmap](../roadmap.md)

This article describes intended behavior.
See the [roadmap progress](../roadmap.md#progress-at-a-glance) for implementation progress.

Deployment answers: how does the desired configuration become active, and how do we establish the outcome?
It operates on complete native configurations independently of how they were authored.

## Preview and dry run

Provide a pre-deploy comparison of live and desired configuration, including additions, modifications, and deletions.
The preview must describe the artifact that will actually be applied.

A device-side dry load is an optional additional capability.
Where supported, load a candidate without applying it, capture errors and the candidate diff, and discard it.
Before enabling this feature, establish its locking, session isolation, and cleanup behavior.
Candidate loading may not execute all commit-time validation and must not be represented as a guarantee of successful deployment.

## Applying configuration

Transfer and load the complete configuration using the target's native mechanism.
Verify that omitted settings are removed; an additive sequence of commands does not meet the desired-state contract.
Whole-file loading also avoids the overhead of issuing separately sequenced commands for every node.

Upload, load, and commit failures must be attributable to the failing stage.
A failed prerequisite must prevent later operations, such as committing after a failed load.
Device-specific commands and error handling should be isolated where firmware behavior differs, without requiring a universal device framework for the first target.

## Status checks and recovery

Report progress while deployment is running, keeping status requests responsive independently of device operations.
Distinguish queued or running work, changes awaiting confirmation, confirmed success, failure, rollback, and an unknown outcome when connectivity prevents verification.

Check device results and retrieve active state where needed to establish what was applied.
A connection loss or successful command submission alone is insufficient evidence of the final state.
The exact post-deployment checks are an open design choice; they must not silently substitute for the configured confirmation policy.

Changes should roll back if not confirmed within a configured deadline.
Prefer native confirmed commit where available, and verify confirmation, persistence, and expiry behavior for each firmware.
Saving persistent configuration must respect this policy.
Confirmation ownership and recovery of pending deployments after application restart remain open decisions.

## Multiple devices and smart deployment

Track desired configuration and deployment state independently for every target.
Determine the affected targets from shared inputs, then compare effective configurations to avoid unnecessary deployments.
A changed source file need not change every device's generated configuration.

Unchanged desired content does not prove the live device is unchanged.
Drift handling needs an explicit policy.
Multi-device execution also requires per-device exclusion of conflicting deployments and clear reporting of partial success.
Cross-device ordering and concurrency can be introduced when concrete dependencies require them.

## Native VyOS deployment integration tests

The [deployment integration tests](../testing.md#device-integration-tests) exercises whole-file replacement, native rejection, explicit confirmation and saving, timeout rollback, and recovery across a container restart.
It establishes behavior on the recorded firmware through a test-only driver; the application deployment command remains to be implemented.
Native commit failure is not necessarily atomic, and the confirmed-commit wrapper's exit status does not reliably indicate validation success on this firmware.
On the recorded firmware, `commit-confirm` can return zero after reporting a failed commit and leave its rollback timer armed.
The driver checks remaining candidate changes through native `cli-shell-api sessionChanged` and preserves timeout recovery after failure.
Confirmation and saving remain separate operations so a rejected or partially applied candidate is not made persistent.

## Open decisions

- Which model and firmware should be supported first, and what native commands implement loading, committing, and recovery?
- Is candidate loading isolated, safe to discard, and useful on each target?
- Who confirms a deployment, when is configuration saved, and how are pending changes reconciled after disconnect or restart?
- Which status checks establish success, and how should smart deployment handle live drift?

## Related capabilities

[Configuration management](configuration.md) supplies complete native artifacts.
[GitHub integration](github-integration.md) connects deployment results to source revisions and review.
[The testing methodology](../testing.md#device-integration-tests) describes the lab environment and native device testing approach.

## Implementation references

[Go-VyOS](https://docs.vyos.io/en/rolling/automation/vyos-govyos.html) is a candidate client for the VyOS HTTP API.
It is not a selected dependency or a substitute for firmware compatibility testing.
