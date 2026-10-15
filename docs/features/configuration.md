# Configuration management

[Feature index](README.md) | [Roadmap](../roadmap.md)

This article describes intended behavior.
See the [roadmap progress](../roadmap.md#progress-at-a-glance) for implementation progress.

Configuration management answers: what should each device be configured to do?
Its output is a complete native configuration for each target.
The same output passes through validation, preview, and deployment, whether it originates from a single file or many shared modules.

## Parsing and device schemas

Support native configuration files and the device definitions needed to interpret them.
Configuration values and schema definitions are separate inputs: a boot file contains desired settings, while node or XML definitions describe available settings and their constraints.

The compatibility goal is to accommodate different Ubiquiti devices and firmware without maintaining a separate hard-coded configuration model for each one.
Support legacy `node.def` trees and XML definitions where applicable.
Schema support alone does not establish device compatibility; loading and committing configurations also differ by firmware.

Use definitions matching the target firmware.
Paths must be configurable or discoverable.
The classic CLI definition tree commonly lives under `/opt/vyatta/share/vyatta-cfg/templates`; current upstream VyOS uses `/usr/share/vyos/templates` for service-rendering templates, which serve a different purpose.
XML command definitions do not imply XML deployment files.
See the [upstream definitions](https://github.com/vyos/vyos-1x) and [template renderer](https://github.com/vyos/vyos-1x/blob/rolling/python/vyos/template.py).

## Static validation

Catch errors that can be determined cheaply from configuration and schema data: unknown paths, incorrect node/value structure, invalid IP addresses or CIDRs, numeric range violations, disallowed values, and strings that violate declared patterns.

Validation has an explicit boundary.
Device-local scripts and runtime-dependent checks remain on the device; load and commit errors must report their results.
A successful static check means that the supported constraints passed, not that a deployment is guaranteed to succeed.
Regex constraints require compatible semantics rather than assuming all device patterns can be evaluated unchanged by Go.

Definitions and parsed constraints may be cached.
A cache must distinguish firmware and definition changes, and allow refresh after upgrades.
The acquisition and cache format are open choices.

## Logical entities

Allow users to describe a host or network once and derive related native settings from it.
For example, a host can own its address, port forwards, hairpin rules, and firewall access attributes.
Changing that host updates the settings derived from those attributes; deleting it removes those settings.

Entities reduce repeated facts and keep related configuration together.
They compile into native configuration, so deployment does not need a separate execution path for every entity type.
Rules that belong to the abstraction itself, such as conflicting host assignments, can be checked during generation without reproducing the device's validation engine.

## Modularization and reuse

Compose device configuration from shared modules, device-specific input, and logical entities.
One shared definition should be usable by multiple devices without copying its contents into independent configurations.

Composition needs deterministic ordering and explicit rules for conflicts, overrides, and deletion.
Removing a module or entity must remove its contribution from the next complete configuration.
Native settings and generated settings also need an ownership rule when both address the same path.
These rules must be settled before shared modules can be deployed reliably.

The result is an effective configuration for each device.
This becomes the basis for validation, previews, and deciding which devices need deployment.

## Open decisions

- Which firmware definitions should be supported first, and how will they be acquired and refreshed?
- Which constraint and regex semantics can be evaluated faithfully?
- What are module precedence and conflict rules, including ownership of native versus generated settings?

## Related capabilities

[GitHub checks](github-integration.md#checks-and-review) publish validation results.
[Device deployment](device-deployment.md) applies the resulting native configuration and reports device-local validation errors.

## Implementation references

[libvyosconfig and ConfigTree](https://github.com/vyos/vyos-1x/blob/rolling/python/vyos/configtree.py) are candidates for native configuration manipulation, not selected dependencies.
[VyOS validation structure](https://docs.vyos.io/en/1.5/contributing/development.html) describes device-side validation responsibilities.
Pin upstream versions when implementing against them.
