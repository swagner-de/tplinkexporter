# Contributing

Thanks for your interest in improving the TP-Link Easy Smart switch exporter.

## Development setup

Prerequisites:

- Go (matching the version in `go.mod`)
- `make`

Common workflows:

```bash
make test        # Run unit tests with coverage
make lint        # Run the configured linter
make lint-fix    # Auto-fix lint issues where possible
make build       # Build the exporter binary into bin/exporter
make run         # Run the exporter locally
```

## Commit conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/)
so that [release-please](https://github.com/googleapis/release-please) can
generate changelogs and cut releases automatically. Use the standard types,
for example:

- `feat: ...` for new features
- `fix: ...` for bug fixes
- `fix(deps): ...` / `chore(deps): ...` for dependency bumps
- `docs: ...`, `ci: ...`, `build: ...`, `chore: ...` for everything else

## Pull requests

1. Fork the repository and create a topic branch.
2. Ensure `make lint` and `make test` pass.
3. Write a clear description of what changed and why.

## Reporting bugs

Open an issue with steps to reproduce, the exporter version, and relevant logs.
Please redact any secrets or switch credentials.
