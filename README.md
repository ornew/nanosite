# Nanosite

Nanosite is an independent Go command that renders repository Markdown into a static site.
It has no dependency on the project's implementation, its root module, or a
neighboring checkout. Repository-specific policy lives in a JSON configuration.

[Tutorial](docs/tutorial/first-site.md) ·
[Configuration guide](docs/guide/configuration.md) ·
[Go extension guide](docs/guide/extensions.md) ·
[Cookbook](docs/cookbook/README.md) ·
[Configuration reference](docs/reference/configuration.md)

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
| `searchHeadings` | Add heading fragments to search results; default false |
| `pager` | Previous/next scope: `section` (default), `group`, `site`, or `none` |
| `sections` | Nested navigation groups with globally unique titles, `match`, `children`, optional `index` and `first` |
| `recordIndexes` | Virtual Markdown indexes for numbered design or optimization records |
| `markdown` | Raw HTML, alerts, heading links/IDs and alignment behavior |
| `tocDepth`, `searchIndex`, `directoryLinks`, `repositoryLinks` | Go generator settings for TOC/search output and source links |
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

For example, record indexes publish a generated list beside design records, and
nested sections organize guides under a common navigation group:

```json
{
  "recordIndexes": [
    {
      "source": "docs/design/README.md",
      "directory": "docs/design",
      "title": "Design Records",
      "introduction": "Accepted and proposed design decisions.",
      "columns": [
        { "title": "Record", "field": "Title" },
        { "title": "Status", "field": "Status" },
        { "title": "Date", "field": "Date" }
      ],
      "required": ["Status", "Date"],
      "templateText": "See the record files for the full rationale."
    }
  ],
  "sections": [
    {
      "title": "Documentation",
      "match": ["README.md"],
      "children": [
        {
          "title": "Guides",
          "match": ["docs/guide/**"],
          "index": "docs/guide/README.md",
          "first": ["docs/guide/getting-started.md"]
        },
        { "title": "Design", "match": ["docs/design/**"] }
      ]
    }
  ],
  "pager": "group",
  "markdown": {
    "rawHTML": false,
    "alerts": true,
    "headingLinks": true,
    "headingIDs": "unicode",
    "alignAttributes": true
  },
  "searchHeadings": true
}
```

Each `recordIndexes` entry creates a virtual Markdown source at `source`; its
front matter and matching defaults are handled like those of a file, and its
route, links and fragments resolve normally. The source file is never created
or overwritten. Records are selected using the same publication rules as pages:
`include`/`exclude`, `publish`, and draft settings apply. Numbered records are
sorted by their NNN prefix; `000` templates are excluded from the list. The
`directory` defaults to the directory containing `source`. `title` and
`columns` are required; `required` lists metadata fields required on each record.
Set `template` to a template basename in that directory. When that named
template is selected for publication, `templateText` is appended as a Markdown
footer. Without `template`, `templateText` is always appended. The template
file's contents are not included.

Navigation sections can contain nested `children`; titles must be globally
unique across all levels. Child matches are checked before their parent match.
An `index` points to the section's README source and controls its link ordering;
`first` lists sources to prioritize among pages tied by weight. Pager scope is
`section` for a child section, `group` for a top-level section and all its
children, `site` for all pages, or `none` to disable adjacent-page links.

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
incorrect types and unclosed front matter fail builds; duplicate YAML keys fail.
Excluded files stay excluded even if their metadata says `publish: true`.

Navigation sorts by weight, then places a section's configured index source and
first-listed sources before other equal-weight pages, then places README
introductions before remaining pages and sorts by source path. Previous/next
links use the configured pager scope and skip pages hidden from navigation.
Heading IDs default to `unicode`, preserving Unicode, removing punctuation,
replacing whitespace with hyphens and suffixing duplicates with `-1`, `-2`, and
so on. `github` uses GitHub-compatible slug rules; `goldmark` uses Goldmark's
automatic IDs. The TOC includes H2 through H4.

The `markdown` object controls extensions: `rawHTML` is disabled by default,
`alerts` and `headingLinks` are disabled by default, and `alignAttributes`
defaults to true. `headingIDs` accepts `unicode` (default), `github`, or
`goldmark`. Set `searchHeadings` to include heading fragments in search results;
it defaults to false.

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

Raw HTML in Markdown is disabled by default. Fenced code, tables, task lists,
strikethrough and normal Markdown are supported through Goldmark GFM. Language-
specific syntax highlighting, math, Mermaid rendering, API reference generation
and WebAssembly builds are extension points rather than bundled project
features. Record indexes can generate design and optimization record lists.
Write custom pages with a template or explicitly copied static HTML.
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

Nanosite began as the site tooling of `ornew/template-basic` and uses that OSS
documentation template itself. Adopting projects can pin a published commit
without waiting for a tagged release.

## Pinned builds and skill synchronization

Use the same revision for generation and the accompanying agent skill:

```sh
go run github.com/ornew/nanosite@FULL_COMMIT_SHA -repo . -config site/config.json -out site/dist
go run github.com/ornew/nanosite@FULL_COMMIT_SHA skill -out .agents/skills/nanosite
go run github.com/ornew/nanosite@FULL_COMMIT_SHA version
```

The command embeds `SKILL.md` and its configuration reference. `skill` installs
only that version's files; it refuses unrelated files and symlink destinations.
The default destination is `.agents/skills/nanosite`. The installed
`.nanosite-version` records the module version. Commit installed skill files with
the generator pin. The template's `hack/update-nanosite-skill.sh` synchronizes
both and also updates the dependency of Go extension drivers.

## Go extension drivers

Projects that need generated pages or custom rendering can use the `sitegen`
package. `sitegen.BuildWith` runs the standard publication pipeline with
`Prepare`, `Decorate`, `View`, and template-function hooks while retaining shared
configuration, routing, link validation, navigation, search, and safe output
replacement. The [Go extension guide](docs/guide/extensions.md) includes a
complete example. The lower-level `render` and `publish` packages remain
available for projects that need a custom pipeline. Pin the module and exported
skill to the same revision.

## Verification

```sh
go test ./...
go vet ./...
```

Tests exercise publication controls, metadata defaults, routing, source-relative
links, fragments, assets, navigation, search data, custom templates, output
ownership and preservation on failure. Also build the actual repository Markdown
and inspect the site after changing layout or browser behavior.
