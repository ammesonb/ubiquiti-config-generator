# Development

## Toolchain

Use Go 1.27.1 (the version declared in `go.mod`), GNU Make, and golangci-lint v2.13.2.
Install the matching prebuilt linter release for your OS from the [golangci-lint releases](https://github.com/golangci/golangci-lint/releases/tag/v2.13.2).
Install Go from [the Go downloads page](https://go.dev/dl/).
Keep the documented versions aligned when upgrading tools.

The Make recipes invoke tools directly and do not require Bash, Docker, or a CI agent.
On Windows, run GNU Make from PowerShell.
`GO` and `GOLANGCI_LINT` can override executable names or paths, for example `make check GOLANGCI_LINT=./.tools/golangci-lint.exe`.

## Commands

| Command | Behavior |
| --- | --- |
| `make` or `make help` | List development commands. |
| `make build` | Build the command into ignored `bin/`. |
| `make run` | Run the command, which currently displays scaffold usage. |
| `make fmt` | Rewrite Go formatting and organize imports. |
| `make lint` | Check formatting, imports, errors, and static analysis, including integration-tagged code. |
| `make vet` | Run Go's built-in analysis, including integration-tagged code. |
| `make tidy` | Update module metadata after dependency changes. |
| `make tidy-check` | Fail if module metadata needs changes, showing a diff without applying it. |
| `make test` | Run ordinary Go tests without a router. |
| `make test-race` | Run ordinary tests with the race detector. |
| `make test-docker` | Build and run ordinary tests in a Go container. |
| `make smoke` | Read-only smoke test using an explicit lab endpoint. |
| `make smoke-docker` | Provision a disposable local VyOS container and smoke-test it. |
| `make check` | Run build, vet, lint, ordinary tests, and module consistency checks. |

Formatting and lint rules are defined in `.golangci.yml`.
Formatting violations cause lint failure; `make fmt` is the explicit repair command.
`.editorconfig` and `.gitattributes` keep indentation and line endings consistent across platforms.

The test harness uses pinned SSH, dotenv, and environment-tag parsing modules recorded in `go.mod` and `go.sum`.
Endpoint environment names, defaults, and required settings are declared on struct fields with `env` tags.
Router smoke tests are explicit opt-in commands and are not part of `make check`.
See [testing](testing.md) for endpoint settings, Docker image preparation, and verification status.

## Layout

- `cmd/ubiquiti-config-generator/`: Go executable entry point.
- `docs/`: product features, delivery roadmap, and development guidance.
- `bin/`: local build output, ignored by Git.
- `internal/testlab/`: test-only SSH and endpoint helpers with ordinary tests.
- `integration/`: opt-in router smoke test and disposable lab fixture.
- `ubiquiti_config_generator/`: existing Python application.
- `sample_router_config/`: existing configuration examples.
- `tests/`: existing Python tests and fixtures.

Add Go implementation packages when a feature needs them.
Keep existing Python files and tooling in place until a focused migration replaces them.
The root Make targets operate on Go packages; Python tooling such as `setup.py`, `reqs.txt`, and the existing pre-commit and CI configuration remains separate.
The GitHub App will read device configurations from separate installed repositories; this scaffold does not contain production configurations or credentials.

## Documentation style

Write Markdown prose with one sentence per source line, keeping blank lines between paragraphs.
Keep tables and fenced code blocks in their required Markdown structure.
