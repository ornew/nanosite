package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedSkillInstallation(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "skills/nanosite")
	if err := writeSkill(out); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SKILL.md", "references/config.md"} {
		want, err := skillFiles.ReadFile(".agents/skills/nanosite/" + name)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatalf("installed %s differs from embedded command version", name)
		}
	}
	if err := os.WriteFile(filepath.Join(out, "SKILL.md"), []byte("outdated"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(out); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(out, "SKILL.md"))
	if err != nil || string(b) == "outdated" {
		t.Fatal("existing skill was not synchronized")
	}
	if err := os.WriteFile(filepath.Join(out, "notes.txt"), []byte("personal"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(out); err == nil {
		t.Fatal("unrelated file should prevent replacement")
	}
	b, err = os.ReadFile(filepath.Join(out, "notes.txt"))
	if err != nil || string(b) != "personal" {
		t.Fatal("unrelated contents changed")
	}
}

func TestSkillRejectsUnrelatedDirectoriesAndSymlinks(t *testing.T) {
	out := filepath.Join(t.TempDir(), "skill")
	if err := writeSkill(out); err != nil {
		t.Fatal(err)
	}
	personal := filepath.Join(out, "personal")
	if err := os.Mkdir(personal, 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(out); err == nil {
		t.Fatal("unrelated empty directory accepted")
	}
	if _, err := os.Stat(personal); err != nil {
		t.Fatal("personal directory removed")
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(out, alias); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(alias); err == nil {
		t.Fatal("symlink accepted")
	}
}
