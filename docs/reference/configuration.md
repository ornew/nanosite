# Configuration and command-line reference

The standard Nanosite command reads one JSON object. Unknown keys, invalid
types and trailing JSON values fail configuration loading. Duplicate JSON keys
are accepted by Go's decoder; avoid them because later values can obscure the
intended setting. The command's `-config` and `-out` paths are resolved from its working
directory. Paths within the config are relative to the input repository unless
noted otherwise.

## Command-line flags

| Flag | Default | Meaning |
|:--|:--|:--|
| `-repo` | `.` | Input repository, read only |
| `-config` | `site/config.json` | JSON configuration file |
| `-out` | `site/dist` | Generated output directory |
| `-check` | false | Validate parsing, routes, rendering and links without writing output |
| `-drafts` | false | Include draft pages for this build or preview |
| `-base-path` | unset | Override the configured deployment prefix, for example `/manual/` |
| `-serve` | unset | Serve the built site at the given host and port, for example `localhost:8080` |

The `skill` and `version` subcommands are separate commands, not build flags.
`-serve` builds once and serves the output; rebuild or restart after edits.

### Skill and version subcommands

| Command | Default | Meaning |
|:--|:--|:--|
| `nanosite version` | — | Print the selected binary's module version, VCS revision (with `+dirty` when applicable), or `(devel)` |
| `nanosite skill` | `-out .agents/skills/nanosite` | Export the skill embedded in this exact Nanosite build |
| `nanosite skill -out PATH` | — | Export to a dedicated skill directory at `PATH` |

For example, export the skill and inspect the matching generator version:

```sh
go run github.com/ornew/nanosite@FULL_COMMIT_SHA skill -out .agents/skills/nanosite
go run github.com/ornew/nanosite@FULL_COMMIT_SHA version
```

Replace `FULL_COMMIT_SHA` with the revision used to build the site. The export
includes `.nanosite-skill` and `.nanosite-version` metadata. The destination
must be a dedicated skill directory; unrelated files and symlink destinations
are rejected. The source of the embedded skill is the `.agents/skills/nanosite/`
directory in the selected Nanosite revision. Keep the exported skill and
generator pin together at that revision.

## Site identity and deployment

| Key | JSON type | Default | Meaning and constraints |
|:--|:--|:--|:--|
| `title` | string | `Documentation` | Required nonempty site name |
| `description` | string | empty | Fallback page description |
| `lang` | string | `en` | Required nonempty HTML language value |
| `repository` | string | empty | Optional repository URL for source links |
| `branch` | string | `main` | Branch used in generated source links |
| `siteURL` | string | empty | Optional HTTP(S) origin only; enables canonical URLs, sitemap and robots output |
| `basePath` | string | `/` | Deployment URL prefix, independent of `siteURL`; must be an absolute path without traversal, query, fragment or backslash |

## Publication and routes

| Key | JSON type | Default | Meaning |
|:--|:--|:--|:--|
| `include` | string array | `README.md`, `docs/**/*.md`, `spec/**/*.md` | Select Markdown inputs by repository-relative glob |
| `exclude` | string array | `**/TEMPLATE.md`, `**/000-template.md` | Exclude pages and assets; exclusions take precedence |
| `assets` | string array | empty | Copy selected local assets at their repository-relative output paths |
| `drafts` | boolean | false | Include pages marked `draft: true`; `-drafts` enables this for one invocation |
| `prettyURLs` | boolean | true | Map `topic.md` to `topic/index.html`; false maps it to `topic.html` |
| `redirects` | object of strings | empty | Map old local routes to published routes; both routes must end in `/` or `.html` |
| `staticDir` | string | empty | Copy a repository directory's contents to the output root |

Glob syntax supports `*`, `?`, character classes, and `**` as a whole path
segment matching zero or more directories. Braces and negated pattern lists are
not supported. `README.md` and `index.md` always map to their directory's
`index.html`. All output routes, aliases, assets and generated files are checked
for collisions.

## Navigation, search, and page layout

| Key | JSON type | Default | Meaning |
|:--|:--|:--|:--|
| `search` | boolean | true | Build and load the local full-text search index |
| `searchHeadings` | boolean | false | Add heading-fragment entries to search results when search is enabled |
| `toc` | boolean | true | Show each page's heading TOC by default; depth is set by `tocDepth` |
| `tocDepth` | integer | 4 | Maximum heading level included in TOC and search headings; valid range 2–6 |
| `pager` | string | `section` | Adjacent-page scope: `section`, `group`, `site`, or `none` |
| `pagerOrder` | string array | empty | Glob patterns that prioritize page neighbors within each top-level group, in listed order |
| `sections` | section array | empty | Ordered navigation tree; see section fields below |
| `order` | string array | empty | Source paths prioritized among equal-weight pages, before README ordering |
| `template` | string | empty | Repository-relative Go HTML template replacing the default full layout |
| `templates` | string array | empty | Repository-relative Go template files parsed as named templates in the layout |
| `styles` | string array | empty | Existing output-relative stylesheets, loaded after built-in CSS |
| `scripts` | string array | empty | Existing output-relative scripts, loaded after built-in JavaScript |
| `directoryLinks` | boolean | false | Use directory URLs for index pages and search result links |
| `repositoryLinks` | boolean | false | When a local link is not a published page or asset, link to its file or directory in the configured repository |
| `bundledAssets` | boolean | true | Copy built-in CSS, JavaScript and theme assets |
| `redirectRules` | boolean | true | Emit the Netlify `_redirects` file for page redirects and aliases |
| `redirectTemplate` | string | empty | Repository-relative Go template for redirect HTML; receives `.Config`, `.Href`, and `.Lang` |
| `searchIndex` | object | see below | Select search output path, JavaScript variable and entry fields |

Setting `redirectRules: false` suppresses only the `_redirects` file; individual
redirect HTML pages are still generated. Disabling `bundledAssets` also removes
the assets referenced by the default layout, so provide a compatible custom
layout and its assets when turning it off. A file in `templates` is parsed as a
named Go template using its basename and can be called from the layout.

Each `sections` object supports:

| Field | JSON type | Default | Meaning |
|:--|:--|:--|:--|
| `title` | string | required | Nonempty title unique across every nesting level |
| `match` | string array | empty | Source globs assigned to this section |
| `collapsed` | boolean | false | Start this navigation group collapsed |
| `children` | section array | empty | Nested sections; child matches are checked before the parent's match |
| `index` | string | empty | README source prioritized among equal-weight pages; its links can inform order |
| `first` | string array | empty | Source paths prioritized after `index` among equal-weight pages |

Pages not matched by configured sections form groups from their `section`
metadata, or `Other` if unset. Pager `section` stays within the page's section;
`group` spans one top-level section and its descendants; `site` spans every
group. Pages hidden with `nav: false` are skipped, and a hidden current page has
no neighbors.
`pagerOrder` patterns match source paths. Matching pages are ordered by the
first matching pattern's position, independently of navigation order; unmatched
pages retain their stable order after matched pages.

With `directoryLinks: true`, directory index pages use trailing-slash routes.
When `repositoryLinks` is enabled, unresolved source links that refer to a
repository file or directory prefer the configured repository URL over an
incidental published route with the same filename.

`tocDepth` controls the deepest heading included in a page TOC and search
heading list. It must be from 2 through 6. `order` lists logical source paths;
listed paths sort before other pages of the same weight, with earlier entries
first.

### Search index schema

`searchIndex` can customize the generated search file and its entry format:

| Field | JSON type | Default | Meaning |
|:--|:--|:--|:--|
| `path` | string | `assets/search-index.js` | Output-relative path for search data |
| `variable` | string | `globalThis.NANOSITE_SEARCH` | Dotted JavaScript identifier assigned the data; when omitted and `path` ends in `.json`, output raw JSON |
| `fields` | object of strings | `title`, `url`, `text` | Map supported entry fields to output property names |
| `text` | string | `markdown` | `markdown` source text or `html` rendered text |
| `headingTitle` | string | `context` | `context` uses “page › heading”; `heading` uses only the heading |
| `headingText` | string | `heading` | `heading` uses heading text; `page` uses page text |
| `section` | string | `leaf` | `leaf` uses the page section; `group` uses its top-level navigation group |

`fields` may map `title`, `url`, `text`, `section`, and `headings` to nonempty
output property names; duplicate output property names are rejected. The
default three properties are `title`, `url`, and `text`. The `headings` value
is a list of heading ID/title pairs. Search heading
entries are included only when `searchHeadings` is true. `headingTitle` accepts
empty/default or `context` and `heading`; `headingText` accepts empty/default,
`heading`, or `page`; `section` accepts empty/default, `leaf`, or `group`.

## Markdown processing

| Key under `markdown` | JSON type | Default | Meaning |
|:--|:--|:--|:--|
| `rawHTML` | boolean | false | Permit raw HTML in Markdown |
| `alerts` | boolean | false | Enable alert block post-processing |
| `headingLinks` | boolean | false | Add links to rendered headings |
| `headingIDs` | string | `unicode` | ID scheme: `unicode`, `github`, or `goldmark` |
| `alignAttributes` | boolean | true | Use HTML alignment attributes for aligned table cells |

The `markdown` object is optional. Unicode IDs lowercase text, keep Unicode
letters/digits/marks plus hyphens and underscores, turn whitespace into
hyphens, remove other punctuation, and suffix duplicate IDs. Goldmark mode
uses Goldmark automatic IDs. The table of contents includes heading levels 2
through 4.

## Configuration defaults and metadata

`defaults` is an ordered array of `{ "match": ["glob"], "values": { ... } }`.
Every matching rule applies in order. Built-in defaults apply first, matching
configuration rules next, and page front matter last. A later explicit value
overrides an earlier value, including false, zero, an empty string, or an empty
array.

The same metadata fields are accepted under `defaults[].values` and as YAML
front matter:

| Field | YAML/JSON type | Default | Meaning |
|:--|:--|:--|:--|
| `title` | string | first H1, then filename | Page title; does not replace the Markdown H1 |
| `description` | string | site `description` | Page description metadata |
| `navTitle` | string | resolved title | Short navigation label |
| `weight` | integer | 0 | Lower values sort first |
| `section` | string | matched section, then `Other` | Navigation group name |
| `layout` | string | `doc` | `home` or `doc` |
| `permalink` | string | source-derived route | Local route ending in `/` or `.html` |
| `aliases` | string array | empty | Old local routes redirected to this page |
| `publish` | boolean | true | False omits the page in every build mode |
| `draft` | boolean | false | True omits the page unless drafts are enabled |
| `nav` | boolean | true | False hides the page from navigation and adjacent-page links |
| `search` | boolean | true | False excludes the page from the search index |
| `toc` | boolean | global `toc` value | Override the page's table of contents |

Front matter is one YAML mapping between opening and closing `---` lines.
Unknown keys, duplicate YAML keys, wrong types, or an unclosed mapping fail the build.
Excluded files remain excluded regardless of metadata. Hiding a page from
navigation or search does not unpublish it.

## Record indexes

`recordIndexes` is an array of virtual Markdown page definitions:

| Field | JSON type | Meaning |
|:--|:--|:--|
| `source` | string, required | Unique repository-relative `.md` source path; must match `include` and not `exclude` |
| `directory` | string | Record directory; defaults to the directory containing `source`; `.` means repository root |
| `title` | string, required | Generated H1 |
| `introduction` | string | Introductory Markdown below the H1 |
| `columns` | array, required | Nonempty list of `{ "title": string, "field": string }` columns |
| `required` | string array | Metadata fields that each listed record must provide |
| `template` | basename | Optional record template filename in `directory`; the file contents are not included |
| `templateText` | string | Markdown footer; appended unconditionally without `template`, or when the named template is published |
| `notice` | string | Generated HTML comment prepended to the index; default empty |
| `optional` | boolean | Skip the index when its record directory does not exist; default false |
| `includeTemplate` | boolean | Include `000-...md` files in the record list; default false |
| `tableClass` | string | CSS class added to the generated record table; default empty |

Records are files named `NNN-name.md`, with numbers sorted numerically; `000`
files are skipped unless `includeTemplate` is true. Configured exclude patterns
still apply. Each needs a level-one heading. Metadata is read from single-line
list items such as `- **Status**: Accepted`; required fields must be
nonempty. Duplicate record numbers or duplicate configured fields in one
record fail. `Title` is a special column field linking to each record.

The index source is generated in memory and is never created or overwritten on
disk. Existing front matter at that source supplies page metadata while the
body is generated. The page follows ordinary front matter/default, route, link
and fragment handling. Records are limited to pages that pass include/exclude,
publication and draft filtering.

## Templates and output paths

Custom layouts use Go `html/template` and receive a `View` with `.Config`,
`.Page`, trusted rendered `.Body`, `.Groups`, `.TOC`, `.Previous`, `.Next`, asset
URLs, `.Source`, `.Canonical`, and `.URL`. Call `.URL` for a path relative to the
current output page. Generated search, CSS, JS and theme URLs are escaped paths
relative to the current output page, which supports deployment under a base path.
The default layout and scripts provide responsive navigation,
search, and light/dark/system theme selection; a replacement template owns the
markup for features it wants to retain.

Markdown links resolve relative to the source path, even for a custom permalink.
Local targets and fragments must exist and be published; external links are not
checked for availability. Generated output must be new, empty, or contain the
Nanosite ownership marker. Failed builds preserve prior output; successful
builds replace it after validation.
