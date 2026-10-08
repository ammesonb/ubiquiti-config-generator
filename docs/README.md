# Project documentation

Ubiquiti Config Generator manages network configuration as version-controlled source.
It turns native files and reusable definitions of hosts, networks, and services into device configurations, supports review and validation through GitHub, and deploys changes with visible status and recovery.

This repository implements a GitHub App.
Separate repositories on which the app is installed hold the actual configuration files and device settings.
See [repository boundaries](features/github-integration.md#github-app-and-repository-boundaries).

- [Features](features/README.md): capabilities, expected behavior, constraints, and open design decisions.
- [Roadmap](roadmap.md): current capability status, delivery milestones, and their dependencies.
- [Development](development.md): tool installation, repository layout, and Make commands.
- [Testing](testing.md): testing methodology, required technologies, and local/CI lab configuration.

Keep behavior and design discussion in the relevant feature article.
Update roadmap status as capabilities become usable, linking implementation or test evidence.
Split articles further when a subject needs enough detail to warrant its own page.
