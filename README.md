# Nanosite

Nanosite is an independent Go command that renders repository Markdown into a static site.
It has no dependency on the project's implementation, its root module, or a
neighboring checkout. Repository-specific policy lives in a JSON configuration.

## Build and preview

Requires Go 1.27.1 or later. From the Nanosite repository root:

```sh
sh site/build.sh
sh site/build.sh -check
sh site/build.sh -serve localhost:8080
```

The build downloads the Go dependencies on first use, then renders local inputs.
Preview builds once; rebuild after editing source or configuration. To deploy
under a URL prefix, use `-base-path /project/`; preview serves that same prefix.
All generated page, navigation and asset links are relative, including search.
Generated pages can also be opened through `file://`.

The reusable command can build an unrelated repository:

```sh
go build -o /tmp/nanosite .
/tmp/nanosite -repo /path/to/project -config /path/to/site.json -out /tmp/project-site
```

`-config` and `-out` are resolved against the command's working directory.
Include patterns, custom template paths and `staticDir` are relative to `-repo`.
`-check` performs parsing, routing, rendering and link validation without writing
output. `-drafts` includes draft pages locally and adds `noindex` metadata.

## Configuration

Start from `site/config.json`. Unknown keys fail the build to catch typos.

| Key | Behavior |
|:--|:--|
| `title`, `description`, `lang` | Site identity, fallback page description, HTML language |
| `repository`, `branch` | Optional repository URL and branch for source links; blank disables them |
| `siteURL` | Optional HTTP(S) origin such as `https://docs.example.com`; enables canonical URLs, sitemap and robots file |
| `basePath` | Deployment prefix, default `/`; use `/project/` for subpath hosting |
| `include` | Repository-relative Markdown globs |
| `exclude` | Globs excluded from pages and copied assets; wins over include |
| `assets` | Explicit repository assets copied at their repository-relative path |
| `prettyURLs` | `topic.md` becomes `topic/index.html` when true, `topic.html` when false |
| `drafts` | Include draft pages; default false |
| `search`, `toc` | Global search and per-page TOC defaults |
| `sections` | Ordered navigation groups: `title`, `match` globs, optional `collapsed` |
| `defaults` | Ordered `match` globs and `values` metadata; later matches override earlier values |
| `redirects` | Old route to published route mapping, e.g. `/old.html` to `/guide/` |
| `template` | Optional repository-relative Go HTML template, replacing the embedded layout |
| `staticDir` | Optional public directory copied directly into the output root |
| `styles`, `scripts` | Additional local output-relative assets loaded after the bundled CSS and JS |

Globs support `*`, `?`, character classes, and `**` as a whole path segment.
`**` matches zero or more directories. Braces, negated pattern lists and
platform-specific separators are not supported. Use separate patterns instead.
The generator does not follow selected input symlinks.

Default inputs are `README.md`, `docs/**/*.md`, and `spec/**/*.md`. Authoring
`TEMPLATE.md` and `000-template.md` files, underscore-prefixed Markdown files,
and underscore-prefixed directories are excluded by the template configuration.
The engine does not hardcode any documentation section or project name.

`README.md` and `index.md` map to their directory's `index.html`. Both in the
same directory cause a collision error. Overlapping routes, file/directory
conflicts, assets, aliases and reserved generator files are checked together.

## YAML front matter

Front matter is optional. A page without it uses matching configuration defaults,
its first H1 as the title, and the default route. Existing Markdown remains
readable in GitHub without requiring front matter.

```yaml
---
title: Streaming guide
description: Process a large input incrementally.
navTitle: Streaming
section: Guides
weight: 20
layout: doc
permalink: /guide/streaming/
aliases:
  - /docs/streaming.html
publish: true
draft: false
nav: true
search: true
toc: true
---
```

| Field | Meaning |
|:--|:--|
| `title`, `description` | Page title and description metadata; title does not replace the Markdown H1 |
| `navTitle` | Short navigation label; falls back to title |
| `section` | Override the matched section; unknown names form additional groups |
| `weight` | Integer sort priority, lower first; absent defaults to zero |
| `layout` | `home` or `doc`; home emphasizes the introductory heading |
| `permalink` | Site-relative route ending in `/` or `.html` |
| `aliases` | Previous routes that redirect directly to this page |
| `publish` | False omits the page even in draft previews |
| `draft` | True omits the page unless drafts are enabled |
| `nav` | False hides it from navigation and adjacent-page links |
| `search` | False excludes it from the local search index |
| `toc` | Override the global TOC default |

Precedence is engine defaults, matching config defaults in order, then page
front matter. Explicit false and zero override earlier values. Unknown keys,
incorrect types, duplicate keys and unclosed front matter fail builds.
Excluded files stay excluded even if their metadata says `publish: true`.

Navigation sorts by weight, then places README introductions first among pages
of equal weight, then sorts by source path. Previous/next links stay within the
same section and skip pages hidden from navigation. Heading IDs preserve Unicode,
remove punctuation, replace whitespace with hyphens and suffix duplicates with
`-1`, `-2`, and so on. The TOC includes H2 through H4.

Hiding a page from navigation or search does not make it private: its HTML is
still published, and it appears in the sitemap when configured. Use
`publish: false`, draft exclusion, or file exclusion to omit its contents.

## Links, assets and redirects

Markdown links resolve against the source Markdown directory, even when the
target has a custom permalink. Directory links resolve to a published README or
index page. Local page and image targets and published heading fragments must
exist. Links to excluded or unpublished pages fail the build; remove the link
or use the repository's explicit external URL when linking authoring sources.
External URLs are preserved without remote availability checks.

Include local images and downloads in `assets`, or put a playground, downloads
and other static resources under `staticDir`. For example:

```json
{
  "staticDir": "site/public",
  "styles": ["assets/project.css"],
  "scripts": ["assets/project.js"]
}
```

This copies `site/public/assets/project.css` to `assets/project.css`. Default
assets are reserved; customize with additional CSS, JS or a complete template.
Any executable browser assets are served locally. Customize Netlify's CSP when
the project deliberately needs additional capabilities such as WebAssembly.

Aliases produce standalone HTML redirect pages and a Netlify `_redirects` file.
They are omitted from navigation, search and sitemap. Redirect chains, external
redirects and missing destinations are rejected. Redirect pages send users to
the target route; they do not preserve arbitrary incoming query/fragment state.

Raw HTML in Markdown is disabled. Fenced code, tables, task lists, strikethrough
and normal Markdown are supported through Goldmark GFM. Language-specific syntax
highlighting, math, Mermaid rendering, API reference generation, automatic design
indexes and WebAssembly builds are extension points rather than bundled project
features. Write custom pages with a template or explicitly copied static HTML.
Custom HTML/template links and static HTML heading fragments are not validated
by the Markdown link checker; verify those custom assets separately.

## Output lifecycle

Input documents are read only. All output is planned, checked and rendered before
replacement. A failed build preserves the previous usable output; a successful
rebuild removes stale pages and assets.

The output directory must be new, empty, or contain the exact `.nanosite`
marker written by this generator. The generator rejects unrelated nonempty output,
repository roots/ancestors and symlinked output paths. Do not keep handwritten
files in generated output. Changes should be serialized against a given output
directory; concurrent builds to the same destination are unsupported.

## Custom templates and repository layout

`layout.html` is the default Go `html/template` layout. A custom template receives
`View`: `.Config`, `.Page`, trusted rendered `.Body`, `.Groups`, `.TOC`,
`.Previous`, `.Next`, relative asset URLs, optional `.Source` and `.Canonical`,
plus `.URL` for relative output paths (`{{call .URL "index.html"}}`). Keep page
content in `.Body`; do not treat metadata as trusted HTML. Custom templates and
public assets are trusted project code.

The Go command, module, tests, embedded `layout.html` and `static/` live at the
repository root. `site/config.json` configures Nanosite's own documentation site;
`site/build.sh` builds it using the root command. The compiled binary embeds its
fallback assets and runs independently of this checkout.

Nanosite began as the site tooling of `ornew/template-basic`. This repository
also uses that OSS documentation template. It has not yet been released: build
from the checkout for now. Once a tag or commit has been published, adopting
projects can install that pinned revision and keep only their site configuration
and build wrapper.

## Verification

```sh
go test ./...
go vet ./...
```

Tests exercise publication controls, metadata defaults, routing, source-relative
links, fragments, assets, navigation, search data, custom templates, output
ownership and preservation on failure. Also build the actual repository Markdown
and inspect the site after changing layout or browser behavior.
