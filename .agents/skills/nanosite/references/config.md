# Nanosite Configuration

- Read `site/build.sh`, `site/config.json` and `site/nanosite.rev` first. The standard CLI and Go extension drivers have different contracts; do not apply unsupported settings.
- `site/nanosite.rev` pins the generator and installed skill to the same commit. The selected command embeds its own skill; run `nanosite skill -out .agents/skills/nanosite` through `hack/update-nanosite-skill.sh`. Do not copy a newer skill independently.

## Standard CLI

- Paths: `-repo` is the input root; `-config`/`-out` are relative to the working directory. Configured source paths are repository-relative.
- Flags: `-check` validates without writes; `-drafts` includes drafts; `-base-path` overrides deployment prefix; `-serve` builds once, then previews.
- Identity: `title`, `description`, `lang`; optional `repository`, `branch` for source links.
- Deployment: `siteURL` is the HTTP(S) origin, `basePath` the separate prefix. An origin enables canonical URLs, sitemap and robots output.
- Publication: `include` selects Markdown, `exclude` wins over includes/assets, `drafts` defaults false. Globs: `*`, `?`, character classes, whole-segment `**`; no braces/negation.
- Routes: `prettyURLs` selects `topic/index.html` or `topic.html`; README/index become directory indexes. Collisions fail.
- Navigation: ordered `sections` have globally unique `title`, `match`, optional `collapsed`, `index`, `first`, and nested `children`. Child matches are checked before parent matches. `index` names the README source for ordering; `first` lists sources to prioritize among equal-weight pages.
- Paging: `pager` is `section` (default), `group` (top-level section and descendants), `site`, or `none`.
- Record indexes: `recordIndexes` entries define a virtual Markdown `source`, optional `directory` (defaults to the source directory), required `title` and `columns` (`title`/`field` pairs), optional `introduction`, `required` metadata fields, a template basename, and Markdown footer `templateText`. The generated source is immutable and receives normal front matter/defaults, routing, link and fragment handling. Records follow page publication filtering (`include`, `exclude`, `publish`, and draft rules); numbered NNN records are listed in order and `000` templates are omitted. `templateText` is appended alone unconditionally; when `template` is set, it is appended only if that named template is published.
- Markdown: `markdown.rawHTML`, `alerts`, and `headingLinks` default false. `headingIDs` defaults to `unicode`; supported values are `unicode`, `github`, and `goldmark`. `alignAttributes` defaults true and accepts explicit false. `searchHeadings` adds heading-fragment results to search and defaults false.
- Assets: `assets` copies at repository-relative paths; `staticDir` maps a public directory to the output root. `styles`/`scripts` list output-relative assets loaded after defaults. `template` replaces the Go HTML layout.
- UI: `search`, `toc`. `redirects` maps old local routes to published destinations; chains/missing destinations fail.

## Standard front matter

- `title`, `description`, `navTitle`: metadata/short navigation label; title does not replace the H1.
- `section`, `weight`: group and lower-first ordering; configured section index/first sources sort first among tied pages, followed by README introductions and source paths.
- `layout`: `home` or `doc`; `permalink`/`aliases`: local routes ending in `/` or `.html`.
- `publish: false`: always omit; `draft: true`: omit unless enabled. File exclusions still win.
- `nav`, `search`, `toc`: navigation/paging, search and H2–H4 TOC controls. Hiding navigation/search does not prevent publication.
- Precedence: engine defaults → matching config defaults → page metadata. Explicit false/zero override earlier values; unknown fields/types fail.
- Markdown links resolve against source paths, including custom permalinks. Validate targets/fragments; include local assets explicitly. Excluded-page links fail; external availability is not checked. Raw HTML is disabled by default.

## Go extension drivers

- Drivers use `render`/`publish` APIs with project themes and special builders. Their flags, routes, heading IDs, navigation, search schemas and Markdown policy come from their code/site guide.
- `site/config.json` source exclusions do not imply support for standard front matter. Unsupported YAML can become visible page content; check the driver's contract first.
- `render.RecordIndexMarkdown`: numbered-record tables, configurable columns/required metadata; edit records, not generated indexes.
- `publish.Run`: owned staging/replacement, preserving prior output on failure. Generated output is disposable; never hand-edit it.
- Preserve theme assets, source links and API/CLI/Playground extensions. For standard CSS overrides, add a local asset to `styles`; handle dark/system modes explicitly.
- Build actual docs and run driver tests. Check routes, fragments, tables, navigation, search, assets and desktop/mobile/color modes; local generation is not deployment evidence.
