# Development

Repository maintenance and contribution instructions belong here. User-facing
installation and feature usage belong in tutorials and guides.

Start with [adopting the template](adopting-template.md) when creating a project.

## Repository map

| Path | Purpose |
|:--|:--|
| `spec/` | Normative public contracts |
| `docs/design/` | Numbered design decisions |
| `docs/optimizations/` | Measured optimization records |
| `site/` | Configuration and wrapper for Nanosite's own documentation |
| `.agents/skills/` | Task-specific agent guidance |

Add implementation, tests, examples and generated artifacts after choosing the
project language. Document their source of truth and generation commands.

## Setup and verification

Use Go 1.27.1 or later. Run `gofmt` on changed Go files, then `go test ./...`
and `go vet ./...` from the root. Build the command with `go build .`. The root
Go module contains the generator; `layout.html` and `static/` are embedded assets.

For the documentation site, see [site maintenance](site.md).

## Contribution workflow

Reproduce the problem, agree on public behavior, implement the smallest complete
change, update documentation and run affected checks. Keep proposed behavior
separate from implemented behavior. Explain the remaining limitations.

Follow the [commit message conventions](commit-message.md). Add release,
compatibility, generated-output and security-reporting procedures when the
project has them; do not invent a support policy in a template.
