# GitHub integration

[Feature index](README.md) | [Roadmap](../roadmap.md)

This article describes intended behavior.
See the [roadmap progress](../roadmap.md#progress-at-a-glance) for implementation progress.

GitHub is the review and history surface for configuration changes.
A configuration repository holds desired state; checks describe whether a proposed revision is acceptable, and deployments record what happened when that revision was applied.
Repository state and actual device state remain distinct.

## GitHub App and repository boundaries

This repository contains the application implementation, tests, packaging, and documentation.
The application runs as a GitHub App installed on separate configuration repositories.
Those repositories contain the native configuration files and device configuration that the app checks and deploys.

Handle events in the context of the app installation and source repository.
Fetch the requested revision using installation-scoped access, and publish checks and deployment records back to that configuration repository.
Device credentials and app credentials belong in runtime secret configuration, not committed configuration files.

GitHub Actions in this implementation repository builds and tests the application and can produce release artifacts.
It does not replace the installed app or require configuration repositories to copy the implementation's CI workflows.
The deployed app needs network access to its target devices; hosted CI runners only need access to disposable test devices.

## Checks and review

For a proposed revision, resolve the affected devices, build their effective configurations, and publish validation results through GitHub checks.
Identify the affected configuration or device in errors so the author can correct the source.

Review should show the proposed configuration's difference from live device state.
A diff between repository revisions alone cannot reveal manual changes on a device.
A preview is also time-dependent: if live state changes after review, the deployment preview needs to be refreshed or otherwise reconciled.

Static checks should be useful without access to a device.
Device-dependent previews and optional dry loads must report unavailable or unsupported results separately from successful checks.

## Deployment triggers and history

The intended automated flow deploys an accepted revision after a merge or push to the configured branch.
Deployment records connect a source revision, target device, configuration artifact, progress, and final outcome.
Manual local deployment remains useful for development and troubleshooting.

History must distinguish an attempted deployment from a confirmed configuration.
For multiple devices, each target has its own result: one successful device cannot mark an entire rollout successful.
The GitHub representation for per-device records and aggregate outcomes remains to be selected.

## Open decisions

- How should GitHub represent per-device deployment records and aggregate rollout outcomes?
- How should a deployment respond when live state has changed since the reviewed preview?

## Related capabilities

[Configuration management](configuration.md) produces and statically validates each device's effective configuration.
[Device deployment](device-deployment.md) supplies live previews, progress, and actual outcomes.
[Testing](../testing.md) describes GitHub Actions as a test environment, separately from the product's GitHub integration.
