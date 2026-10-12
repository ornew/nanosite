package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordIndexMetadataAndMarkdownEscaping(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, "docs/design")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	src := "# 001. A [link] | record\n\n- **Status**: Draft | WIP\n- **Date**: 2026-10-12\n\n- **Example**: first\n- **Example**: second\n"
	if err := os.WriteFile(filepath.Join(dir, "001-record.md"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	c := RecordIndex{Directory: "docs/design", Heading: "Design Records", Columns: []Column{{"Title", "Title"}, {"Status", "Status"}, {"Date", "Date"}}, Required: []string{"Status", "Date"}}
	got, err := RecordIndexMarkdown(repo, c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `Draft \| WIP`) || !strings.Contains(string(got), `A \[link\] \| record`) {
		t.Fatalf("unsafe table metadata: %s", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "001-duplicate.md"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadRecords(repo, c); err == nil {
		t.Fatal("duplicate IDs accepted")
	}
}

func TestMarkdownModesAndNavigation(t *testing.T) {
	src := []byte("# 日本語 `API`\n\n## Topic\n\n## Topic\n\n> [!NOTE]\n> Keep it.\n\n<div>raw</div>\n")
	d := Parse(src, Options{RawHTML: true, Alerts: true, HeadingLinks: true})
	got, err := d.Render(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`id="日本語-api"`, `id="topic-1"`, `class="callout callout-note"`, `<div>raw</div>`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s", want)
		}
	}
	safe := Parse(src, Options{})
	got, err = safe.Render(nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "<div>raw</div>") {
		t.Fatal("raw HTML enabled by default")
	}
	prev, next := Adjacent([]int{1, 2, 3, 4}, 3, func(i int) bool { return i%2 == 0 })
	if prev != 2 || next != 4 {
		t.Fatal(prev, next)
	}
}
