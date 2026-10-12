# Generate an index of numbered records

Use `recordIndexes` to build a Markdown table from numbered files in a
directory. Nanosite generates the configured source in memory, so you do not
need to maintain a duplicate index file. Existing source Markdown is never
rewritten.

## Add the records

Create `docs/decisions/001-cache-policy.md`:

```markdown
# Cache policy

- **Status**: Accepted
- **Date**: 2026-08-14

Cache immutable responses for one hour.
```

Create `docs/decisions/002-search-index.md`:

```markdown
# Search index

- **Status**: Proposed
- **Date**: 2026-09-02

Build a compact index for published guides.
```

Each listed filename must match `NNN-name.md`, where `NNN` is a three-digit
number other than `000`. The first H1 is the record title. Metadata uses a
single-item Markdown list: `- **Field:** value`. The configured `required` keys
must be present and nonempty. Duplicate record numbers in one directory fail
the build.

## Configure the generated page

Add this object to the site's configuration. The source path must match an
`include` pattern and not match `exclude`. This example assumes there is no
physical `docs/decisions/README.md`; Nanosite supplies that virtual Markdown
page and routes it as the directory index.

```json
{
  "recordIndexes": [
    {
      "source": "docs/decisions/README.md",
      "directory": "docs/decisions",
      "title": "Decisions",
      "introduction": "Accepted and proposed decisions for this project.",
      "columns": [
        { "title": "Decision", "field": "Title" },
        { "title": "Status", "field": "Status" },
        { "title": "Date", "field": "Date" }
      ],
      "required": ["Status", "Date"]
    }
  ],
  "sections": [
    {
      "title": "Decisions",
      "match": ["docs/decisions/**"],
      "index": "docs/decisions/README.md"
    }
  ]
}
```

`directory` defaults to the directory containing `source`; use `"."` to index
the repository root. `columns` controls both labels and values. The special
`Title` field links each row to its record. Other fields come from the metadata
list items. The optional `template` is a basename in the records directory. It
does not contribute its file contents; when that named record template is
selected for publication, `templateText` is appended as a Markdown footer. If
`template` is omitted, `templateText` is appended unconditionally.

## Publication and routes

The index is a virtual Markdown page. Its front matter, matching `defaults`,
permalink, aliases and links use the normal page rules. It does not create or
modify a source file. A record appears only when it is included and passes the
same publication rules as other pages: file exclusions, `publish: false`, and
draft settings are honored. A `000` template file is omitted from the numbered
list. If records use custom permalinks, generated links resolve through their
source paths and still lead to the published routes.

To place the index first among equal-weight pages in its navigation group, set
that group's `index` field to the same virtual source path. Validate the
configuration with `nanosite -check` (or the equivalent `go run` command in the
[first-site tutorial](../tutorial/first-site.md)) before building the site.

The generated Markdown table contains the two records in numeric order:

```markdown
| Decision | Status | Date |
|:--|:--|:--|
| [Cache policy](001-cache-policy.md) | Accepted | 2026-08-14 |
| [Search index](002-search-index.md) | Proposed | 2026-09-02 |
```
