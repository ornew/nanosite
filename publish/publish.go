// Package publish plans and atomically publishes documentation output produced
// by Nanosite or a project extension driver. Source and theme policy stay with
// the caller; ownership, staging and failure recovery are shared.
package publish

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Repo, Out     string
	Marker        string
	MarkerContent []byte
	Protected     []string // repository-relative files/directories the output must not contain
}

func within(parent, child string) bool {
	r, e := filepath.Rel(parent, child)
	return e == nil && r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator))
}

func validate(o Options) error {
	if o.Marker == "" || o.Marker == "." || o.Marker == ".." || filepath.Base(o.Marker) != o.Marker {
		return fmt.Errorf("invalid output marker")
	}
	if info, err := os.Lstat(o.Out); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output must not be a symlink: %s", o.Out)
	}
	repo, e := canonicalPath(o.Repo)
	if e != nil {
		return e
	}
	out, e := canonicalPath(o.Out)
	if e != nil {
		return e
	}
	if within(out, repo) {
		return fmt.Errorf("output directory contains source: %s", out)
	}
	for _, p := range o.Protected {
		source, err := canonicalPath(filepath.Join(repo, p))
		if err != nil {
			return err
		}
		if within(out, source) {
			return fmt.Errorf("output directory contains source: %s", out)
		}
	}
	if info, e := os.Lstat(out); e == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output must not be a symlink: %s", out)
	}
	entries, e := os.ReadDir(out)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	if len(entries) == 0 {
		return nil
	}
	info, e := os.Lstat(filepath.Join(out, o.Marker))
	if e != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to replace unmarked output directory: %s", out)
	}
	b, e := os.ReadFile(filepath.Join(out, o.Marker))
	if e != nil || string(b) != string(o.MarkerContent) {
		return fmt.Errorf("refusing to replace output with an invalid ownership marker: %s", out)
	}
	return nil
}

// Run invokes generate in a new sibling staging directory, then replaces only
// owned output. Failed generation preserves the previous build. Calls targeting
// the same output directory must be serialized.
func Run(o Options, generate func(stage string) error) error {
	if err := validate(o); err != nil {
		return err
	}
	out, err := canonicalPath(o.Out)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(out), ".nanosite-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := os.WriteFile(filepath.Join(stage, o.Marker), o.MarkerContent, 0644); err != nil {
		return err
	}
	if err := generate(stage); err != nil {
		return err
	}
	info, err := os.Lstat(stage)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("generator replaced staging directory")
	}
	markerInfo, err := os.Lstat(filepath.Join(stage, o.Marker))
	if err != nil || !markerInfo.Mode().IsRegular() {
		return fmt.Errorf("generator replaced staging ownership marker")
	}
	marker, err := os.ReadFile(filepath.Join(stage, o.Marker))
	if err != nil || string(marker) != string(o.MarkerContent) {
		return fmt.Errorf("generator changed staging ownership marker")
	}
	if err := validate(o); err != nil {
		return err
	}
	backup := ""
	if _, err := os.Stat(out); err == nil {
		backup = stage + "-previous"
		if err := os.Rename(out, backup); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(stage, out); err != nil {
		if backup != "" {
			if restore := os.Rename(backup, out); restore != nil {
				return fmt.Errorf("publish: %v; restore %s: %w", err, backup, restore)
			}
		}
		return err
	}
	if backup != "" {
		return os.RemoveAll(backup)
	}
	return nil
}

func WriteFile(name string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return os.WriteFile(name, data, 0644)
}

// Resolve existing ancestor symlinks even when the output does not exist yet.
func canonicalPath(p string) (string, error) {
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(a); os.IsNotExist(err) {
		parent, err := canonicalPath(filepath.Dir(a))
		if err != nil {
			return "", err
		}
		return filepath.Join(parent, filepath.Base(a)), nil
	}
	return filepath.EvalSymlinks(a)
}
