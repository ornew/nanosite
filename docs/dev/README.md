# Development

Nanosite is a Go documentation generator. Its command and reusable `sitegen`,
`render`, and `publish` packages live in one Go module at the repository root. The `site/`
directory configures the documentation site for this source checkout; it is not
a separate Go module.

## Repository map

| Path | Purpose |
|:--|:--|
| `main.go`, `skills.go` | CLI, version reporting and embedded skill export |
| `sitegen/` | Standard configuration, Markdown pipeline, build API and Go extensions |
| `render/` | Routing, navigation and record-index APIs |
| `publish/` | Safe output staging and replacement |
| `sitegen/layout.html`, `sitegen/static/` | Embedded default site layout and browser assets |
| `.agents/skills/nanosite/` | Skill source embedded in the command by `skills.go` |
| `site/config.json`, `site/build.sh` | Configuration and wrapper for this checkout's own site |
| `docs/` | Tutorials, guides, recipes, reference, design and development docs |

## Setup and checks

Use Go 1.27.1 or later. Format changed Go files with `gofmt`; the first build
may download dependencies. From the repository root:

```sh
go test ./...
go vet ./...
go build .
```

The tests cover the CLI, rendering and publication packages, and the embedded
skill. `go test ./...` also runs the browser-focused search URL check when Node.js
is installed and skips it when Node.js is absent. Do not treat these checks as
evidence of a deployed site.

## Documentation and site changes

Write user instructions in tutorials, guides, recipes and reference pages.
Keep development setup here and site-maintenance details in [site maintenance](site.md).
The Nanosite skill source is `.agents/skills/nanosite/`; the build embeds that
directory into the CLI. Keep its configuration reference aligned with the
standard CLI and `sitegen.BuildWith`, which share the same config and publication
pipeline.

For Markdown, config, or theme changes, validate and rebuild the local site:

```sh
sh site/build.sh -check
sh site/build.sh
```

Follow the [commit-message conventions](commit-message.md). Keep local checks,
CI, site deployment and external link availability as separate evidence.
