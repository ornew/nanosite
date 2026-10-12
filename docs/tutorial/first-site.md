# Build your first site

This tutorial creates a small Markdown site, checks its links and routes, and
previews the generated files. Nanosite reads the repository without changing its
Markdown sources.

## Create a site repository

Nanosite requires Go 1.27.1 or later. Create a separate directory for the
content you want to publish:

```sh
mkdir -p /tmp/first-site/docs /tmp/first-site/site
cd /tmp/first-site
```

Create `README.md`:

```markdown
# Field Notes

Short guides for the team.

[Start with the first guide](docs/first-guide.md)
```

Create `docs/first-guide.md`:

```markdown
---
title: First guide
description: A short example page.
---

# First guide

This page is written in Markdown and will become a static HTML page.
```

Create `site/config.json`:

```json
{
  "title": "Field Notes",
  "description": "Short guides for the team.",
  "lang": "en",
  "include": ["README.md", "docs/**/*.md"],
  "exclude": ["**/TEMPLATE.md"],
  "sections": [
    { "title": "Overview", "match": ["README.md"] },
    { "title": "Guides", "match": ["docs/**"] }
  ]
}
```

Unknown configuration keys fail the build, so misspelled options are caught
early. See the [configuration guide](../guide/configuration.md) for publication
rules and the [configuration reference](../reference/configuration.md) for all
fields.

## Pin and run Nanosite

Choose a full Nanosite commit SHA and use that same revision for every command.
For a published build, replace `FULL_COMMIT_SHA` below with the revision you
intend to pin. The command fetches that exact Go module version:

```sh
go run github.com/ornew/nanosite@FULL_COMMIT_SHA -repo . -config site/config.json -out site/dist -check
go run github.com/ornew/nanosite@FULL_COMMIT_SHA -repo . -config site/config.json -out site/dist
go run github.com/ornew/nanosite@FULL_COMMIT_SHA -repo . -config site/config.json -out site/dist -serve localhost:8080
```

The first command validates the inputs without writing output. The second
creates `site/dist/`; the third rebuilds and serves it at
`http://localhost:8080`. Preview builds once, so restart it after editing source
files. Generated output must be new, empty, or marked as output owned by
Nanosite.

To run from a Nanosite source checkout, use the checkout path instead of a
published module version:

```sh
go -C /path/to/nanosite run . -repo /tmp/first-site -config /tmp/first-site/site/config.json -out /tmp/first-site/site/dist -check
go -C /path/to/nanosite run . -repo /tmp/first-site -config /tmp/first-site/site/config.json -out /tmp/first-site/site/dist
go -C /path/to/nanosite run . -repo /tmp/first-site -config /tmp/first-site/site/config.json -out /tmp/first-site/site/dist -serve localhost:8080
```

## What to explore next

- Use the [configuration guide](../guide/configuration.md) to control drafts,
  routes, navigation, search and assets.
- Follow [the record-index recipe](../cookbook/record-index.md) to generate a
  list from numbered Markdown records.
- Follow [the custom-theme recipe](../cookbook/custom-theme.md) to add CSS or
  replace the HTML layout.
