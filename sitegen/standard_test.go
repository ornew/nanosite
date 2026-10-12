package sitegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ornew/nanosite/render"
)

func TestStandardRecordIndexRespectsPublicationAndRoutes(t *testing.T) {
	repo, out, c := fixture(t)
	c.Sections = append(c.Sections, Section{Title: "Design", Match: []string{"docs/design/**"}})
	c.RecordIndexes = []RecordIndexConfig{{Source: "docs/design/README.md", Directory: "docs/design", Title: "Design Records", Columns: []render.Column{{Title: "Record", Field: "Title"}, {Title: "Status | phase", Field: "Status"}}, Required: []string{"Status"}}}
	put(t, repo, "README.md", "# Home\n[Records](docs/design/)\n")
	source := "---\ntitle: Decision index\npermalink: /decisions/\n---\n# Old hand-maintained index\n"
	put(t, repo, "docs/design/README.md", source)
	put(t, repo, "docs/design/001-first.md", "---\npermalink: /chosen/\n---\n# 001. A [record] | choice\n\n- **Status**: Draft | WIP\n")
	put(t, repo, "docs/design/002-private.md", "---\npublish: false\n---\n# 002. Private\n")
	put(t, repo, "docs/design/003-draft.md", "---\ndraft: true\n---\n# 003. Draft\n")
	put(t, repo, "docs/design/004-TEMPLATE.md", "# 004. Invalid excluded record\n")
	c.Exclude = append(c.Exclude, "docs/design/004-TEMPLATE.md")
	run(t, repo, out, c)
	page := read(t, out, "decisions/index.html")
	for _, want := range []string{"<table>", "Decision index", `href="../chosen/index.html"`, "Draft | WIP", "Status | phase"} {
		if !strings.Contains(page, want) {
			t.Fatalf("missing %q: %s", want, page)
		}
	}
	if strings.Contains(page, "Private") || strings.Contains(page, "003.") {
		t.Fatal("unpublished record included")
	}
	if read(t, repo, "docs/design/README.md") != source {
		t.Fatal("source index modified")
	}
	if !strings.Contains(read(t, out, "index.html"), `href="decisions/index.html"`) {
		t.Fatal("generated index permalink not resolved")
	}
	put(t, repo, "docs/design/001-first.md", "# 001. Missing metadata\n")
	expectError(t, repo, out, c, "missing Status")
	if read(t, out, "decisions/index.html") != page {
		t.Fatal("failed index generation replaced usable output")
	}
}

func TestSyntheticRecordIndexAndUnpublishedIndex(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home\n")
	put(t, repo, "docs/design/001-first.md", "# 001. First\n\n- **Status**: Accepted\n")
	c.RecordIndexes = []RecordIndexConfig{{Source: "docs/design/README.md", Directory: "docs/design", Title: "Records", Columns: []render.Column{{Title: "Record", Field: "Title"}}, TemplateText: "Available records."}}
	c.Sections = append(c.Sections, Section{Title: "Records", Match: []string{"docs/design/**"}, Index: "docs/design/README.md"})
	run(t, repo, out, c)
	if !strings.Contains(read(t, out, "docs/design/index.html"), "Available records.") {
		t.Fatal("standalone templateText ignored")
	}
	if !strings.Contains(read(t, out, "docs/design/index.html"), "001. First") {
		t.Fatal("synthetic index missing")
	}
	if _, err := os.Stat(filepath.Join(repo, "docs/design/README.md")); !os.IsNotExist(err) {
		t.Fatal("synthetic source written")
	}
	put(t, repo, "docs/design/README.md", "---\npublish: false\n---\n# Hidden index\n")
	run(t, repo, out, c)
	if _, err := os.Stat(filepath.Join(out, "docs/design/index.html")); !os.IsNotExist(err) {
		t.Fatal("unpublished index resurrected")
	}
}

func TestNestedNavigationOrderingAndPagerModes(t *testing.T) {
	repo, out, c := fixture(t)
	c.Sections = []Section{{Title: "Overview", Match: []string{"README.md"}}, {Title: "Backends", Match: []string{"docs/backends/**"}, Children: []Section{{Title: "Local", Match: []string{"docs/backends/local/**"}, Index: "docs/backends/local/README.md", Collapsed: true}, {Title: "Remote", Match: []string{"docs/backends/remote/**"}, Collapsed: true}}}}
	put(t, repo, "README.md", "# Home\n")
	put(t, repo, "docs/backends/README.md", "# Backends\n")
	put(t, repo, "docs/backends/local/README.md", "# Local\n[Z][last]\n[A](a.md)\n\n[last]: z.md\n")
	put(t, repo, "docs/backends/local/a.md", "# A\n")
	put(t, repo, "docs/backends/local/z.md", "# Z\n")
	put(t, repo, "docs/backends/remote/README.md", "# Remote\n")
	r := run(t, repo, out, c)
	var z *Page
	for _, p := range r.Pages {
		if p.Title == "Z" {
			z = p
		}
	}
	groups, err := orderedNavigation(repo, r.Pages, z, c)
	if err != nil {
		t.Fatal(err)
	}
	group := groups[1]
	if !group.Active || !group.Children[0].Active || group.Children[1].Active {
		t.Fatal("ancestor activation failed")
	}
	local := group.Children[0].Pages
	if len(local) != 3 || local[1].Title != "Z" || local[2].Title != "A" {
		t.Fatal("index-link order not applied")
	}
	prev, next := pageNeighbors(groups, z, Config{Pager: "section"})
	if prev == nil || prev.Title != "Local" || next == nil || next.Title != "A" {
		t.Fatal("section pager order failed")
	}
	last := local[2]
	_, next = pageNeighbors(groups, last, Config{Pager: "section"})
	if next != nil {
		t.Fatal("section pager escaped child")
	}
	_, next = pageNeighbors(groups, last, Config{Pager: "group"})
	if next == nil || next.Title != "Remote" {
		t.Fatal("group pager missed child boundary")
	}
	prev, next = pageNeighbors(groups, z, Config{Pager: "none"})
	if prev != nil || next != nil {
		t.Fatal("disabled pager enabled")
	}
	no := false
	z.Meta.Nav = &no
	prev, next = pageNeighbors(groups, z, Config{Pager: "site"})
	if prev != nil || next != nil {
		t.Fatal("hidden page received pager")
	}
	if !strings.Contains(read(t, out, "docs/backends/local/z/index.html"), `<summary>Local</summary>`) {
		t.Fatal("nested navigation not rendered")
	}
}

func TestConfigurableMarkdownAndHeadingSearch(t *testing.T) {
	repo, out, c := fixture(t)
	c.Markdown = MarkdownOptions{RawHTML: true, Alerts: true, HeadingLinks: true, HeadingIDs: "github"}
	c.SearchHeadings = true
	put(t, repo, "README.md", "# Home\n\n## A &amp; `B`\n\n> [!NOTE]\n> Useful.\n\n<div>Raw</div>\n")
	run(t, repo, out, c)
	html := read(t, out, "index.html")
	for _, want := range []string{`id="a--b"`, `class="anchor"`, `callout-note`, `<div>Raw</div>`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s", want)
		}
	}
	if !strings.Contains(read(t, out, "assets/search-index.js"), `index.html#a--b`) {
		t.Fatal("heading search absent")
	}
	c.Markdown = MarkdownOptions{}
	run(t, repo, out, c)
	html = read(t, out, "index.html")
	if strings.Contains(html, "<div>Raw</div>") || strings.Contains(html, `class="anchor"`) {
		t.Fatal("opt-in rendering leaked into default")
	}
}

func TestExtendedConfigurationRejectsInvalidContracts(t *testing.T) {
	for _, body := range []string{`{"pager":"unknown"}`, `{"markdown":{"headingIDs":"unknown"}}`, `{"sections":[{"title":"A","children":[{"title":"A"}]}]}`, `{"recordIndexes":[{"source":"docs/README.md","directory":"..","title":"X","columns":[{"title":"Title","field":"Title"}]}]}`} {
		file := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := readConfig(file); err == nil {
			t.Errorf("invalid config accepted: %s", body)
		}
	}
	// Public JSON keys for columns must round-trip as documented.
	data, err := json.Marshal(render.Column{Title: "Status", Field: "Status"})
	if err != nil || !strings.Contains(string(data), `"field"`) {
		t.Fatal(string(data), err)
	}
}

func TestRootRecordDirectoryDefault(t *testing.T) {
	repo, out, c := fixture(t)
	c.Include = []string{"*.md"}
	put(t, repo, "001-record.md", "# Root record\n")
	config := `{"include":["*.md"],"recordIndexes":[{"source":"README.md","title":"Root records","columns":[{"title":"Record","field":"Title"}]}]}`
	file := filepath.Join(repo, "config.json")
	put(t, repo, "config.json", config)
	var err error
	c, err = readConfig(file)
	if err != nil {
		t.Fatal(err)
	}
	run(t, repo, out, c)
	if !strings.Contains(read(t, out, "index.html"), "Root record") {
		t.Fatal("root-directory records missing")
	}
}
