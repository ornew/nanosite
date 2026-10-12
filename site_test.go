package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) (string, string, Config) {
	t.Helper()
	root, err := canonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	c := defaultConfig()
	c.Sections = []Section{{Title: "Overview", Match: []string{"README.md"}}, {Title: "Guides", Match: []string{"docs/guide/**"}}}
	return repo, filepath.Join(root, "output"), c
}
func put(t *testing.T, root, name, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}
func read(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func run(t *testing.T, repo, out string, c Config) *Result {
	t.Helper()
	r, err := build(repo, out, c, false)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func expectError(t *testing.T, repo, out string, c Config, want string) {
	t.Helper()
	_, err := build(repo, out, c, false)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v; want %q", err, want)
	}
}

func TestMarkdownLinksHeadingsAndAssets(t *testing.T) {
	repo, out, c := fixture(t)
	c.Assets = []string{"docs/**/*.svg"}
	put(t, repo, "README.md", "# Example\n\n[Guide](docs/guide/)\n[日本語](docs/guide/first.md#日本語)\n")
	put(t, repo, "docs/guide/README.md", "# Guides\n[First](first.md?q=yes#same-1)\n")
	put(t, repo, "docs/guide/first.md", "# First\n## Same\n## Same\n## 日本語\n![diagram](../media/diagram.svg)\n[Home](../../README.md)\n[Heading](#same)\n\n| A | B |\n| - | - |\n| x | y |\n\n```text\n<unsafe>\n```\n")
	put(t, repo, "docs/media/diagram.svg", "<svg xmlns=\"http://www.w3.org/2000/svg\"/>")
	r := run(t, repo, out, c)
	if len(r.Pages) != 3 {
		t.Fatalf("pages=%d", len(r.Pages))
	}
	home := read(t, out, "index.html")
	if !strings.Contains(home, `href="docs/guide/index.html"`) {
		t.Fatal("directory Markdown link not rewritten")
	}
	guide := read(t, out, "docs/guide/index.html")
	if !strings.Contains(guide, `href="first/index.html?q=yes#same-1"`) {
		t.Fatal("query/fragment not preserved")
	}
	page := read(t, out, "docs/guide/first/index.html")
	for _, want := range []string{`id="same-1"`, `id="日本語"`, `src="../../media/diagram.svg"`, `href="../../../index.html"`, `href="#same"`, `<table>`, `&lt;unsafe&gt;`} {
		if !strings.Contains(page, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if read(t, out, "docs/media/diagram.svg") == "" {
		t.Fatal("asset not copied")
	}
}

func TestPublicationAndDefaults(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home\n")
	put(t, repo, "docs/guide/hidden.md", "---\ntitle: Hidden\nnav: false\nsearch: false\nweight: 0\ndraft: false\n---\n# Hidden\n")
	put(t, repo, "docs/guide/draft.md", "---\ndraft: true\n---\n# Draft\n")
	put(t, repo, "docs/guide/private.md", "---\npublish: false\n---\n# Private\n")
	put(t, repo, "docs/guide/TEMPLATE.md", "# Template\n[Invalid](missing.md)")
	yes, no := true, false
	w := 20
	c.Defaults = []Defaults{{Match: []string{"docs/**"}, Values: Metadata{Nav: &yes, Search: &yes, Weight: &w}}, {Match: []string{"README.md"}, Values: Metadata{TOC: &no}}}
	r := run(t, repo, out, c)
	if len(r.Pages) != 2 {
		t.Fatal("publication controls ignored")
	}
	if strings.Contains(read(t, out, "index.html"), ">Hidden</a>") {
		t.Fatal("hidden page appears in nav")
	}
	index := read(t, out, "assets/search-index.js")
	if strings.Contains(index, "Hidden") || strings.Contains(index, "Draft") {
		t.Fatal("search exclusions ignored")
	}
	if strings.Contains(read(t, out, "docs/guide/hidden/index.html"), `content="noindex"`) {
		t.Fatal("explicit draft:false became noindex")
	}
	c.Drafts = true
	r = run(t, repo, out, c)
	if len(r.Pages) != 3 {
		t.Fatal("draft preview did not include drafts")
	}
	if !strings.Contains(read(t, out, "docs/guide/draft/index.html"), `content="noindex"`) {
		t.Fatal("draft preview should be noindex")
	}
}

func TestFailuresPreservePreviousOutput(t *testing.T) {
	for _, bad := range []string{"[Missing](missing.md)", "[Anchor](#absent)", "[Template](docs/guide/TEMPLATE.md)", "![Asset](docs/image.png)", "[Outside](../outside.md)"} {
		t.Run(bad, func(t *testing.T) {
			repo, out, c := fixture(t)
			put(t, repo, "README.md", "# Good\n")
			run(t, repo, out, c)
			before := read(t, out, "index.html")
			put(t, repo, "README.md", "# Broken\n"+bad)
			expectError(t, repo, out, c, "link")
			if read(t, out, "index.html") != before {
				t.Fatal("failed build changed output")
			}
		})
	}
}

func TestUnknownMetadataAndMalformedFrontMatter(t *testing.T) {
	for _, body := range []string{"---\ndraf: true\n---\n# H", "---\nweight: wrong\n---\n# H", "---\ndraft: true\n# H", "---\nlayout: unknown\n---\n# H"} {
		repo, out, c := fixture(t)
		put(t, repo, "README.md", body)
		if _, err := build(repo, out, c, false); err == nil {
			t.Fatalf("accepted invalid metadata: %s", body)
		}
	}
}

func TestRoutesAliasesAndFlatURLs(t *testing.T) {
	repo, out, c := fixture(t)
	c.PrettyURLs = false
	c.BasePath = "/project/"
	c.SiteURL = "https://example.com"
	put(t, repo, "README.md", "# Home\n[Intro](docs/guide/intro.md#heading)\n")
	put(t, repo, "docs/guide/intro.md", "---\npermalink: /learn/start/\naliases: [/old.html]\n---\n# Intro\n## Heading\n[Flat](flat.md)\n")
	put(t, repo, "docs/guide/flat.md", "# Flat\n[Alias](/old.html#heading)\n")
	run(t, repo, out, c)
	if !strings.Contains(read(t, out, "index.html"), `href="learn/start/index.html#heading"`) {
		t.Fatal("custom permalink ignored")
	}
	if !strings.Contains(read(t, out, "learn/start/index.html"), `href="../../docs/guide/flat.html"`) {
		t.Fatal("source-relative link followed output directory")
	}
	if !strings.Contains(read(t, out, "docs/guide/flat.html"), `href="../../learn/start/index.html#heading"`) {
		t.Fatal("alias link not resolved")
	}
	if !strings.Contains(read(t, out, "_redirects"), "/project/old.html /project/learn/start/ 301") {
		t.Fatal("deployment prefix missing")
	}
	if !strings.Contains(read(t, out, "sitemap.xml"), "https://example.com/project/learn/start/") {
		t.Fatal("canonical deployment prefix missing")
	}
}

func TestOutputCollisions(t *testing.T) {
	for _, name := range []string{"docs/guide/index.md", "docs/guide/README.md"} {
		repo, out, c := fixture(t)
		put(t, repo, "README.md", "# Home")
		put(t, repo, "docs/guide.md", "# Guide")
		put(t, repo, name, "# Index")
		expectError(t, repo, out, c, "collision")
	}
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "---\naliases: [/index.html]\n---\n# Home")
	expectError(t, repo, out, c, "collision")
	repo, out, c = fixture(t)
	c.Assets = []string{"assets/**"}
	put(t, repo, "README.md", "# H")
	put(t, repo, "assets/site.css", "custom")
	expectError(t, repo, out, c, "collision")
}

func TestOutputProtectionStaleRemovalAndCheck(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home")
	put(t, out, "personal.txt", "Keep me")
	expectError(t, repo, out, c, "refusing nonempty")
	if read(t, out, "personal.txt") != "Keep me" {
		t.Fatal("unrelated output changed")
	}
	if err := os.RemoveAll(out); err != nil {
		t.Fatal(err)
	}
	put(t, repo, "docs/guide/old.md", "# Old")
	run(t, repo, out, c)
	if err := os.Remove(filepath.Join(repo, "docs/guide/old.md")); err != nil {
		t.Fatal(err)
	}
	run(t, repo, out, c)
	if _, err := os.Stat(filepath.Join(out, "docs/guide/old/index.html")); !os.IsNotExist(err) {
		t.Fatal("stale page retained")
	}
	before := read(t, out, "index.html")
	put(t, repo, "README.md", "# New")
	if _, err := build(repo, out, c, true); err != nil {
		t.Fatal(err)
	}
	if read(t, out, "index.html") != before {
		t.Fatal("check wrote output")
	}
	expectError(t, repo, repo, c, "must not contain")
}

func TestSymlinksAndTraversal(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home")
	external := filepath.Join(filepath.Dir(repo), "secret.md")
	put(t, filepath.Dir(repo), "secret.md", "secret")
	if err := os.MkdirAll(filepath.Join(repo, "docs/guide"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(repo, "docs/guide/leak.md")); err != nil {
		t.Fatal(err)
	}
	expectError(t, repo, out, c, "symlink")
	if err := os.Remove(filepath.Join(repo, "docs/guide/leak.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(filepath.Dir(repo), "alias")
	if err := os.Symlink(out, alias); err != nil {
		t.Fatal(err)
	}
	expectError(t, repo, alias, c, "symlinks")
	put(t, repo, "README.md", "---\npermalink: /../escape/\n---\n# Home")
	expectError(t, repo, out, c, "traversal")
}

func TestRawHTMLIsDisabledAndMetadataEscaped(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "---\ntitle: '<script>bad()</script>'\ndescription: '\" onload=\"bad()'\n---\n# Home\n<script>alert(1)</script>\n")
	run(t, repo, out, c)
	page := read(t, out, "index.html")
	if strings.Contains(page, "<script>bad()") || strings.Contains(page, "<script>alert(1)") {
		t.Fatal("HTML injection")
	}
	if !strings.Contains(page, "&lt;script&gt;bad()") {
		t.Fatal("metadata not escaped")
	}
}

func TestConfigValidation(t *testing.T) {
	for _, body := range []string{`{"unknown":true}`, `{"basePath":"/../escape"}`, `{"include":["../*.md"]}`, `{"include":["**/["]}`, `{"sections":[{"title":"x"},{"title":"x"}]}`, `{} {}`} {
		root := t.TempDir()
		file := filepath.Join(root, "config.json")
		put(t, root, "config.json", body)
		if _, err := readConfig(file); err == nil {
			t.Fatalf("accepted config %s", body)
		}
	}
}

func TestUnicodeSpacesAndExternalURLs(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home\n[Local](docs/guide/a%20b.md#日本語)\n[Remote](https://example.com/missing#anchor)\n[Mail](mailto:hello@example.com)\n")
	put(t, repo, "docs/guide/a b.md", "# Page\n## 日本語\n")
	run(t, repo, out, c)
	if !strings.Contains(read(t, out, "index.html"), `href="docs/guide/a%20b/index.html#`) {
		t.Fatal("spaces not encoded")
	}
}

func TestGlobs(t *testing.T) {
	for _, tt := range []struct {
		p, n string
		want bool
	}{{"docs/**/*.md", "docs/a.md", true}, {"docs/**/*.md", "docs/a/b.md", true}, {"**/_*/**", "docs/_draft/a.md", true}, {"**/_*.md", "_a.md", true}, {"docs/*.md", "docs/a/b.md", false}, {"docs/**", "spec/a.md", false}} {
		if got := glob(tt.p, tt.n); got != tt.want {
			t.Fatalf("glob(%q,%q)=%v", tt.p, tt.n, got)
		}
	}
}

func TestDeterministicBuildAndNoSourceWrites(t *testing.T) {
	repo, out, c := fixture(t)
	body := "# Home\n## Hello\n"
	put(t, repo, "README.md", body)
	run(t, repo, out, c)
	before := read(t, out, "index.html")
	run(t, repo, out, c)
	if read(t, out, "index.html") != before {
		t.Fatal("identical inputs produced different output")
	}
	if read(t, repo, "README.md") != body {
		t.Fatal("source Markdown changed")
	}
}

func TestSectionPagingAndSearchText(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home")
	put(t, repo, "docs/guide/a.md", "---\nweight: 20\n---\n# A\nUniqueSearchToken\n")
	put(t, repo, "docs/guide/b.md", "---\nweight: 10\n---\n# B\n")
	run(t, repo, out, c)
	a := read(t, out, "docs/guide/a/index.html")
	if !strings.Contains(a, "← B") {
		t.Fatal("section paging ignores weight")
	}
	if strings.Contains(a, "← Home") {
		t.Fatal("paging crossed section")
	}
	index := read(t, out, "assets/search-index.js")
	raw := strings.TrimSuffix(strings.TrimPrefix(index, "globalThis.NANOSITE_SEARCH = "), ";\n")
	var entries []map[string]string
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e["title"] == "A" {
			if !strings.Contains(e["text"], "UniqueSearchToken") || strings.Contains(e["text"], "weight") {
				t.Fatal("search must contain body, not front matter")
			}
			return
		}
	}
	t.Fatal("missing search page")
}

func TestCustomTemplateAndStaticDirectory(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home\n[Download](/download.txt)\n")
	put(t, repo, "site/public/download.txt", "example")
	put(t, repo, "site/public/assets/custom.css", "body{}")
	put(t, repo, "site/public/assets/custom.js", "void 0;")
	c.StaticDir = "site/public"
	c.Styles = []string{"assets/custom.css"}
	c.Scripts = []string{"assets/custom.js"}
	run(t, repo, out, c)
	if !strings.Contains(read(t, out, "index.html"), `href="assets/custom.css"`) {
		t.Fatal("custom styles missing")
	}
	if read(t, out, "download.txt") != "example" {
		t.Fatal("public file missing")
	}
	put(t, repo, "site/custom.html", `<!doctype html><title>{{.Config.Title}}</title><article>{{.Body}}</article>`)
	c.Template = "site/custom.html"
	run(t, repo, out, c)
	if strings.Contains(read(t, out, "index.html"), "sidebar") {
		t.Fatal("custom template ignored")
	}
	c.Styles = []string{"assets/missing.css"}
	expectError(t, repo, out, c, "missing styles/scripts")
}

func TestSearchPreservesInlineIdentifiers(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# PROJECT_NAME\n\n`some_identifier` and snake_case with **bold** text.\n")
	run(t, repo, out, c)
	index := read(t, out, "assets/search-index.js")
	for _, word := range []string{"PROJECT_NAME", "some_identifier", "snake_case", "bold text"} {
		if !strings.Contains(index, word) {
			t.Fatalf("search text split %s", word)
		}
	}
}

func TestTableAlignmentAndEscapedCells(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home\n\n| Left | Center | Right |\n| :--- | :---: | ---: |\n| a\\|b | **bold** | `<tag>` |\n")
	run(t, repo, out, c)
	body := read(t, out, "index.html")
	for _, want := range []string{`align="left"`, `align="center"`, `align="right"`, `a|b`, `<strong>bold</strong>`, `&lt;tag&gt;`} {
		if !strings.Contains(body, want) {
			t.Fatalf("table lacks %s", want)
		}
	}
	if strings.Contains(body, `style="text-align`) {
		t.Fatal("inline table style violates default CSP")
	}
}
