package sitegen

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtensionPublicationAnchorsAndAtomicValidation(t *testing.T) {
	repo, out, c := fixture(t)
	put(t, repo, "README.md", "# Home\n[Generated](docs/generated.md#a&b)\n")
	c.Exclude = append(c.Exclude, "docs/excluded.md")
	ext := Extensions{Prepare: func(ctx *Context) error {
		hidden, err := ctx.AddHTMLPage("docs/excluded.md", "excluded.html", "Hidden", nil)
		if err != nil {
			return err
		}
		if hidden != nil {
			t.Fatal("excluded extension published")
		}
		p, err := ctx.AddHTMLPage("docs/generated.md", "generated.html", "Generated", []byte(`<h2 id="a&amp;b">Anchor</h2><div data-id="ghost"></div>`))
		if err != nil {
			return err
		}
		if p.Anchors["ghost"] || !p.Anchors["a&b"] {
			t.Fatal("incorrect generated anchors")
		}
		ctx.Assets["generated.txt"] = []byte("asset")
		return nil
	}}
	if _, err := BuildWith(repo, out, c, false, ext); err != nil {
		t.Fatal(err)
	}
	initial := read(t, out, "index.html")
	called := false
	ext.Validate = func(stage string) error {
		called = true
		if read(t, stage, "generated.html") == "" {
			t.Fatal("missing staged page")
		}
		return errors.New("validation failed")
	}
	for _, check := range []bool{true, false} {
		called = false
		if _, err := BuildWith(repo, out, c, check, ext); err == nil || !strings.Contains(err.Error(), "validation failed") {
			t.Fatal(err)
		}
		if !called || read(t, out, "index.html") != initial {
			t.Fatal("failed validation changed output or skipped hook")
		}
	}
	ext.Validate = nil
	missing := filepath.Join(filepath.Dir(out), "dry-run")
	if _, err := BuildWith(repo, missing, c, true, ext); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("dry run published")
	}
}
func TestDuplicateSearchFieldNamesRejected(t *testing.T) {
	repo, _, _ := fixture(t)
	put(t, repo, "config.json", `{"searchIndex":{"fields":{"title":"x","url":"x"}}}`)
	if _, err := ReadConfig(filepath.Join(repo, "config.json")); err == nil {
		t.Fatal("duplicate mapped fields accepted")
	}
}
