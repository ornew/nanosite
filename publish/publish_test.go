package publish

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPreservesOutputOnFailureAndProtectsAliasedSources(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	out := filepath.Join(root, "dist")
	if err := os.MkdirAll(filepath.Join(repo, "docs"), 0755); err != nil {
		t.Fatal(err)
	}
	o := Options{Repo: repo, Out: out, Marker: ".owned", MarkerContent: []byte("v1"), Protected: []string{"docs"}}
	if err := Run(o, func(stage string) error { return WriteFile(filepath.Join(stage, "index.html"), []byte("previous")) }); err != nil {
		t.Fatal(err)
	}
	if err := Run(o, func(stage string) error {
		_ = WriteFile(filepath.Join(stage, "index.html"), []byte("partial"))
		return errors.New("failed extension")
	}); err == nil {
		t.Fatal("failed generation accepted")
	}
	got, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil || string(got) != "previous" {
		t.Fatal("failed build replaced usable output")
	}
	if err := Run(o, func(stage string) error { return os.Remove(filepath.Join(stage, o.Marker)) }); err == nil {
		t.Fatal("removed marker accepted")
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(repo, alias); err != nil {
		t.Fatal(err)
	}
	o.Out = filepath.Join(alias, "docs")
	called := false
	if err := Run(o, func(string) error { called = true; return nil }); err == nil || called {
		t.Fatal("aliased source accepted")
	}
}

func TestCheckLinksWithQueryAndRootRoutes(t *testing.T) {
	out := t.TempDir()
	_ = WriteFile(filepath.Join(out, "index.html"), []byte(`<h1 id="intro">Intro</h1>`))
	_ = WriteFile(filepath.Join(out, "guide/index.html"), []byte(`<a href="/?q=hello#intro">Home</a>`))
	if problems := CheckLinks(out); len(problems) != 0 {
		t.Fatal(problems)
	}
}
