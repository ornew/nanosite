# Configure a site

Nanosite takes a repository of Markdown and assets plus a JSON configuration.
Start with `site/config.json`. Paths in configuration are relative to the
repository unless an option says otherwise. Unknown keys and invalid types fail
the build.

## Choose what to publish

`include` selects repository-relative Markdown paths. The default is
`README.md`, `docs/**/*.md`, and `spec/**/*.md`. `exclude` takes precedence over
both included Markdown and assets. The default excludes `TEMPLATE.md` and
`000-template.md`; underscore-prefixed files are not excluded automatically.
Patterns support `*`, `?`, character classes and `**` as a whole path segment;
there are no brace expansions or negated patterns.

The global `drafts` setting defaults to false. A page with `draft: true` is
omitted unless drafts are enabled in configuration or by the `-drafts` preview
flag. `publish: false` always omits it. File exclusions still apply. A hidden
page can still be fetched directly if it is published: `nav: false` and
`search: false` only affect those features, so use publication controls to keep
content out of the output.

## Set page metadata

Place optional YAML front matter at the beginning of a Markdown file:

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
draft: false
publish: true
nav: true
search: true
toc: true
---
```

`title` does not replace the Markdown H1. Without a title, the first H1 supplies
it; without an H1, the source filename is used. `navTitle` is the shorter label
shown in navigation. `section` selects a navigation group, and `weight` sorts
lower values first. `layout` is `home` or `doc`. `permalink` and each alias must
be a local route ending in `/` or `.html`. Aliases redirect to the page's
published route.

Configuration `defaults` can set any page metadata for matching source globs.
Values are applied in this order: built-in defaults, each matching config rule
in list order, then the page's own front matter. Later values replace earlier
ones; explicit `false`, `0`, and empty values are meaningful overrides.
Front matter accepts only known keys and expected types.

## Routes and links

With `prettyURLs: true` (the default), `docs/start.md` becomes
`docs/start/index.html`; with false, it becomes `docs/start.html`. A source named
`README.md` or `index.md` becomes its directory's `index.html` in either mode.
Routes, aliases and generated files are checked for collisions. A `siteURL`
provides the canonical URL origin and enables a sitemap and robots file;
`basePath` independently sets a deployment prefix such as `/manual/`.

Markdown links resolve from the source file's directory even when the target has
a custom permalink. Local pages, assets and heading fragments must resolve.
External URLs are preserved, but Nanosite does not check their availability.
Add image and download paths to `assets`; exclusions apply to them too. The
optional `staticDir` copies a public directory to the output root, keeping its
relative file paths.

## Organize navigation and paging

`sections` is an ordered tree of groups. Each section has a globally unique
`title`, `match` globs and optional `collapsed` and `children` fields. Child
matches are tested before the containing section's own match. Pages not matched
by a section are grouped by their `section` metadata or under `Other`.

Within a section, pages sort by weight. An optional `index` names the README
source to put first among tied pages, and `first` lists additional source paths
to prioritize in that tie. Remaining tied pages put README introductions before
sorting by source path. `nav: false` hides a page from navigation and page
neighbors.

`pager` controls previous/next links: `section` (default) stays within the same
page section; `group` spans a top-level section and its nested children; `site`
spans all navigation groups; `none` disables the links. A page hidden from
navigation has no page neighbors.
Use `pagerOrder` to prioritize neighbor order with source-path glob patterns.
Patterns apply within each top-level group and the first matching pattern in
the array has priority; navigation ordering remains unchanged.

## Configure Markdown and search

`markdown` controls parsing and presentation:

- `rawHTML` allows raw HTML and defaults to false.
- `alerts` enables alert block syntax and defaults to false.
- `headingLinks` adds links to headings and defaults to false.
- `headingIDs` is `unicode` (default), `github`, or `goldmark`.
- `alignAttributes` controls table alignment attributes and defaults to true.

The table of contents is on by default and includes H2 through H4 at the default
`tocDepth` of 4; `tocDepth` accepts values from 2 to 6. Global `toc` or
per-page `toc` can turn it off. Search is on by default. `searchHeadings: true`
adds heading fragments as search results; it defaults to false. Per-page
`search: false` omits the page from search without unpublishing it.

Additional `sitegen` settings include `tocDepth`, which sets the deepest TOC and
search heading level (2–6, default 4), and `directoryLinks`, which uses
directory URLs for index pages and search results. `repositoryLinks` can turn
unresolved local source links into repository links when `repository` is
configured; repository file/directory links take precedence over an incidental
published route with the same filename. `searchIndex` customizes the search output path, JavaScript
variable, entry properties and text/section format. See the
[Go extension guide](extensions.md) and the full
[configuration reference](../reference/configuration.md) for these fields.

## Add CSS, scripts, or a template

Use `styles` and `scripts` for output-relative local assets. They load after
Nanosite's built-in stylesheet and scripts, so a CSS override can reuse its
color variables and theme modes. `staticDir` is convenient for a small asset
directory. See the [custom-theme recipe](../cookbook/custom-theme.md) for a
complete example. Setting `template` replaces the entire HTML layout; its Go
template receives the page, rendered body, navigation, TOC and URL helper.

For every documented field and command-line option, see the
[configuration reference](../reference/configuration.md).
