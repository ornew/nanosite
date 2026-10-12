# Documentation Site

Nanosite is the independent Go documentation generator module at the repository root. It reads
repository Markdown and creates static HTML without changing source documents.
Its configuration contains the project name, publication rules and navigation;
the generator itself does not depend on the project's implementation language.

## Build and preview

Use Go 1.27.1 or later, from the repository root:

```sh
sh site/build.sh
sh site/build.sh -check
sh site/build.sh -serve localhost:8080
```

The output is `site/dist/`, ignored by Git. Preview builds once; restart after
editing source. The first build needs dependencies downloaded or cached.

## Publication controls

Edit `site/config.json` to choose Markdown `include` and `exclude` patterns.
Templates and underscore-prefixed drafts are excluded by the starter config.
YAML front matter can set `title`, `description`, `navTitle`, `weight`, `section`,
`layout`, `permalink`, `aliases`, `publish`, `draft`, `nav`, `search`, and `toc`.
Matching `defaults` rules supply metadata before page-specific overrides.

```yaml
---
navTitle: Streaming
weight: 20
draft: true
---
```

Drafts are omitted by default; `sh site/build.sh -drafts` includes them in a local
preview. `publish: false` omits a page in every mode. Hiding navigation/search
alone does not omit the page from the output.

The default site has configurable section navigation, section-local paging,
H2–H4 tables of contents, local full-text search, light/dark/system colors,
responsive navigation and print styles. Page links and heading fragments are
checked before output is replaced. Failed builds retain the previous output;
successful builds remove stale pages. Do not hand-edit generated files.

## Assets and customization

Use `assets` globs for repository images and downloads. `staticDir` copies a
public directory directly into the output. Optional `styles`, `scripts` and a
custom Go HTML `template` extend the presentation without changing the engine.
See the root `README.md` for the full configuration, precedence, supported syntax,
link rules, output protections, extension limits and extraction instructions.

## Netlify

`netlify.toml` builds with Go and publishes `site/dist/`. Connect the actual
repository and configure the production branch and custom domain in Netlify.
Set `siteURL` to the actual HTTP(S) origin to emit canonical URLs and a sitemap;
set `basePath` separately for a subpath deployment. Leave `siteURL` blank until
it is known. Add real legacy redirects rather than copying another project's
hostnames. A local build verifies generated output, not deployment or DNS.

## Verification

```sh
go test ./...
go vet ./...
sh site/build.sh -check
sh site/build.sh
```

Inspect desktop and mobile navigation, search, color modes and any custom
browser assets after presentation changes. External URL availability and
custom raw HTML assets are outside the Markdown link checker.
