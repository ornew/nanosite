package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

//go:embed all:.agents/skills/nanosite
var skillFiles embed.FS

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		revision := ""
		dirty := false
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				revision = s.Value
			}
			if s.Key == "vcs.modified" {
				dirty = s.Value == "true"
			}
		}
		if revision != "" {
			if dirty {
				revision += "+dirty"
			}
			return revision
		}
	}
	return "(devel)"
}

func skillCommand(args []string) error {
	flags := flag.NewFlagSet("nanosite skill", flag.ContinueOnError)
	out := flags.String("out", ".agents/skills/nanosite", "destination skill directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected skill arguments")
	}
	if err := writeSkill(*out); err != nil {
		return err
	}
	fmt.Printf("installed Nanosite skill (%s) in %s\n", version(), *out)
	return nil
}

// writeSkill updates only a dedicated skill directory, refusing unrelated files.
// The skill is embedded in the selected command version, not fetched separately.
func writeSkill(out string) error {
	out, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if info, e := os.Lstat(out); e == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("skill destination must not be a symlink")
	}
	out, err = canonicalPath(out)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cwd, err = canonicalPath(cwd)
	if err != nil {
		return err
	}
	if out == cwd || out == filepath.Dir(out) {
		return fmt.Errorf("choose a dedicated skill directory")
	}
	allowed := map[string]bool{".nanosite-skill": true, ".nanosite-version": true, ".revision": true}
	dirs := map[string]bool{".": true}
	files := map[string][]byte{}
	const base = ".agents/skills/nanosite"
	err = fs.WalkDir(skillFiles, base, func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			rel := strings.TrimPrefix(name, base+"/")
			if name == base {
				rel = "."
			}
			dirs[rel] = true
			return nil
		}
		rel := strings.TrimPrefix(name, base+"/")
		allowed[rel] = true
		b, e := skillFiles.ReadFile(name)
		if e != nil {
			return e
		}
		files[rel] = b
		return nil
	})
	if err != nil {
		return err
	}
	validate := func() error {
		if info, e := os.Lstat(out); e == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("skill destination must not be a symlink")
		}
		e := filepath.WalkDir(out, func(name string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			rel, e := filepath.Rel(out, name)
			if e != nil {
				return e
			}
			if d.IsDir() {
				if !dirs[filepath.ToSlash(rel)] {
					return fmt.Errorf("refusing unrelated skill directory: %s", rel)
				}
				return nil
			}
			if !d.Type().IsRegular() || !allowed[filepath.ToSlash(rel)] {
				return fmt.Errorf("refusing unrelated skill destination entry: %s", rel)
			}
			return nil
		})
		if os.IsNotExist(e) {
			return nil
		}
		return e
	}
	if err := validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(out), ".nanosite-skill-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	files[".nanosite-skill"] = []byte("Nanosite embedded skill v1\n")
	files[".nanosite-version"] = []byte(version() + "\n")
	for name, data := range files {
		p := filepath.Join(stage, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(p, data, 0644); err != nil {
			return err
		}
	}
	if err := validate(); err != nil {
		return err
	}
	backup := ""
	if _, e := os.Stat(out); e == nil {
		backup = stage + "-previous"
		if e := os.Rename(out, backup); e != nil {
			return e
		}
	} else if !os.IsNotExist(e) {
		return e
	}
	if e := os.Rename(stage, out); e != nil {
		if backup != "" {
			if restore := os.Rename(backup, out); restore != nil {
				return fmt.Errorf("install: %v; restore %s: %w", e, backup, restore)
			}
		}
		return e
	}
	if backup != "" {
		return os.RemoveAll(backup)
	}
	return nil
}
