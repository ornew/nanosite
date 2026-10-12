# Maintain the Nanosite site

The Nanosite repository builds its own documentation from the root Go module.
From the repository root, `sh site/build.sh` changes into `site/` and runs the
root checkout with `-repo .. -config config.json -out dist`. It does not download
another Nanosite revision and does not support `NANOSITE_SOURCE` or
`NANOSITE_BIN` overrides.

## Build and inspect

Use Go 1.27.1 or later. From the repository root:

```sh
sh site/build.sh -check
sh site/build.sh
sh site/build.sh -serve localhost:8080
```

The output is `site/dist/`, which is ignored by Git. `-check` validates without
writing; a successful build replaces generated pages and assets after checking
the inputs. A failed build preserves the prior output. The output directory must
be new, empty, or marked as owned by Nanosite. Preview builds once; rebuild and
restart it after editing content or configuration.

The `site/config.json` file controls the repository's published Markdown,
navigation, search, theme, and source links. The [configuration guide](../guide/configuration.md)
explains reader-facing settings; the [reference](../reference/configuration.md)
lists all standard fields and flags. Do not hand-edit `site/dist/`.

## Source checkout versus a pinned consumer build

The wrapper in this repository builds the current source checkout, which is
useful when changing the generator or its site. A consuming project should pin a
full Nanosite commit instead of relying on a neighboring checkout. From that
project's root, a fixed-revision build can use:

```sh
go run github.com/ornew/nanosite@FULL_COMMIT_SHA -repo . -config site/config.json -out site/dist -check
go run github.com/ornew/nanosite@FULL_COMMIT_SHA -repo . -config site/config.json -out site/dist
```

Replace `FULL_COMMIT_SHA` with the selected revision. Keep the generator and
installed skill at that same revision. Run
`go run github.com/ornew/nanosite@FULL_COMMIT_SHA skill -out .agents/skills/nanosite`
to export the skill from the pinned command. Some consuming-project scaffolds
provide a helper to synchronize the generator pin and exported skill; follow
that project's wrapper and configuration rather than adding a second
source-checkout override.

## Publication and customization

The repository config excludes authoring templates and underscore-prefixed
drafts. Front matter can set page title, navigation label, section, weight,
permalink, aliases, publication, draft, search and TOC behavior. `siteURL` is
left blank until the canonical origin is known; `basePath` separately configures
subpath hosting. Local builds do not verify DNS, remote URLs or deployment.

For content and theme changes, inspect navigation and search on desktop and
mobile, check light/dark/system colors, and confirm local links and fragments.
If browser assets change, run the optional Node.js browser-focused checks along
with the Go test suite. The root `layout.html` and `static/` directories are the
source for the embedded default layout and client assets.
