package render

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// SourcePolicy is the configurable discovery portion of an extension driver's
// site/config.json. Project themes and special page builders remain Go callbacks.
type SourcePolicy struct {
	Exclude []string `json:"exclude"`
}

func ReadSourcePolicy(repo string) (SourcePolicy, error) {
	var p SourcePolicy
	data, err := os.ReadFile(filepath.Join(repo, "site/config.json"))
	if os.IsNotExist(err) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, nil
}
func Glob(pattern, name string) bool {
	p, n := strings.Split(pattern, "/"), strings.Split(name, "/")
	var match func(int, int) bool
	match = func(i, j int) bool {
		if i == len(p) {
			return j == len(n)
		}
		if p[i] == "**" {
			return match(i+1, j) || j < len(n) && match(i, j+1)
		}
		if j == len(n) {
			return false
		}
		ok, _ := path.Match(p[i], n[j])
		return ok && match(i+1, j+1)
	}
	return match(0, 0)
}
func Matches(patterns []string, name string) bool {
	for _, p := range patterns {
		if Glob(p, name) {
			return true
		}
	}
	return false
}
func Discover(repo string, include, exclude []string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(repo, func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".claude" || d.Name() == ".agents" || strings.HasPrefix(d.Name(), ".nanosite-build-") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(repo, file)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if Matches(exclude, rel) || !Matches(include, rel) {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("selected source must not be a symlink: %s", rel)
		}
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}
func MarkdownOutput(source string, pretty bool, indexNames []string) string {
	for _, name := range indexNames {
		if path.Base(source) == name {
			return path.Join(path.Dir(source), "index.html")
		}
	}
	stem := strings.TrimSuffix(source, path.Ext(source))
	if pretty {
		return path.Join(stem, "index.html")
	}
	return stem + ".html"
}

// OrderSources preserves explicit first entries followed by links in a Markdown
// index and then alphabetical source paths. It does not invent missing pages.
func OrderSources(files []string, index string, first []string, body []byte) {
	rank := map[string]int{}
	for i, f := range first {
		rank[f] = i + 1
	}
	if index != "" {
		rank[index] = 1
		for _, m := range mdLinkRe.FindAllSubmatch(body, -1) {
			f := path.Join(path.Dir(index), string(m[1]))
			if _, ok := rank[f]; !ok {
				rank[f] = len(rank) + 1
			}
		}
	}
	sort.SliceStable(files, func(i, j int) bool {
		ri, rj := rank[files[i]], rank[files[j]]
		switch {
		case ri > 0 && rj > 0:
			return ri < rj
		case ri > 0 || rj > 0:
			return ri > 0
		}
		return files[i] < files[j]
	})
}

var mdLinkRe = regexp.MustCompile(`\]\(([^)#\s]+\.md)(?:#[^)]*)?\)`)
