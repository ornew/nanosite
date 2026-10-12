package main

import (
	"encoding/json"
	"fmt"
	"github.com/ornew/nanosite/render"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
)

// All repository-specific choices live in Config. The generator can be moved
// without importing the adopted project's implementation or tooling.
type Config struct {
	Template    string            `json:"template"`
	StaticDir   string            `json:"staticDir"`
	Styles      []string          `json:"styles"`
	Scripts     []string          `json:"scripts"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Lang        string            `json:"lang"`
	Repository  string            `json:"repository"`
	Branch      string            `json:"branch"`
	SiteURL     string            `json:"siteURL"`
	BasePath    string            `json:"basePath"`
	Include     []string          `json:"include"`
	Exclude     []string          `json:"exclude"`
	Assets      []string          `json:"assets"`
	Drafts      bool              `json:"drafts"`
	PrettyURLs  bool              `json:"prettyURLs"`
	Search      bool              `json:"search"`
	TOC         bool              `json:"toc"`
	Sections    []Section         `json:"sections"`
	Defaults    []Defaults        `json:"defaults"`
	Redirects   map[string]string `json:"redirects"`
}

type Section struct {
	Title     string   `json:"title"`
	Match     []string `json:"match"`
	Collapsed bool     `json:"collapsed"`
}

type Defaults struct {
	Match  []string `json:"match"`
	Values Metadata `json:"values"`
}

// Pointers distinguish absent values from explicit false, zero and empty text.
type Metadata struct {
	Title       *string   `json:"title,omitempty" yaml:"title"`
	Description *string   `json:"description,omitempty" yaml:"description"`
	NavTitle    *string   `json:"navTitle,omitempty" yaml:"navTitle"`
	Weight      *int      `json:"weight,omitempty" yaml:"weight"`
	Section     *string   `json:"section,omitempty" yaml:"section"`
	Layout      *string   `json:"layout,omitempty" yaml:"layout"`
	Permalink   *string   `json:"permalink,omitempty" yaml:"permalink"`
	Aliases     *[]string `json:"aliases,omitempty" yaml:"aliases"`
	Publish     *bool     `json:"publish,omitempty" yaml:"publish"`
	Draft       *bool     `json:"draft,omitempty" yaml:"draft"`
	Nav         *bool     `json:"nav,omitempty" yaml:"nav"`
	Search      *bool     `json:"search,omitempty" yaml:"search"`
	TOC         *bool     `json:"toc,omitempty" yaml:"toc"`
}

func defaultConfig() Config {
	return Config{Title: "Documentation", Lang: "en", Branch: "main", BasePath: "/",
		Include:    []string{"README.md", "docs/**/*.md", "spec/**/*.md"},
		Exclude:    []string{"**/TEMPLATE.md", "**/000-template.md"},
		PrettyURLs: true, Search: true, TOC: true}
}

func readConfig(file string) (Config, error) {
	c := defaultConfig()
	f, err := os.Open(file)
	if err != nil {
		return c, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return c, fmt.Errorf("config: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return c, fmt.Errorf("config must contain one JSON object")
	}
	if c.Title == "" || c.Lang == "" {
		return c, fmt.Errorf("config title and lang must be nonempty")
	}
	if c.BasePath == "" {
		c.BasePath = "/"
	}
	if !strings.HasPrefix(c.BasePath, "/") || strings.ContainsAny(c.BasePath, "?#\\") || strings.Contains(c.BasePath, "..") {
		return c, fmt.Errorf("basePath must be an absolute URL path without traversal, query or fragment")
	}
	c.BasePath = "/" + strings.Trim(path.Clean(c.BasePath), "/")
	if c.BasePath != "/" {
		c.BasePath += "/"
	}
	seen := map[string]bool{}
	for _, s := range c.Sections {
		if s.Title == "" || seen[s.Title] {
			return c, fmt.Errorf("section titles must be nonempty and unique: %q", s.Title)
		}
		seen[s.Title] = true
	}
	patterns := append(append(append([]string{}, c.Include...), c.Exclude...), c.Assets...)
	for _, s := range c.Sections {
		patterns = append(patterns, s.Match...)
	}
	for _, d := range c.Defaults {
		patterns = append(patterns, d.Match...)
	}
	for _, p := range patterns {
		if err := validatePattern(p); err != nil {
			return c, err
		}
	}
	return c, nil
}

func mergeMetadata(base, next Metadata) Metadata {
	b, n := reflect.ValueOf(&base).Elem(), reflect.ValueOf(next)
	for i := 0; i < b.NumField(); i++ {
		if !n.Field(i).IsNil() {
			b.Field(i).Set(n.Field(i))
		}
	}
	return base
}

func boolValue(v *bool, fallback bool) bool {
	if v != nil {
		return *v
	}
	return fallback
}
func textValue(v *string, fallback string) string {
	if v != nil {
		return *v
	}
	return fallback
}
func weight(m Metadata) int {
	if m.Weight != nil {
		return *m.Weight
	}
	return 0
}

func validatePattern(p string) error {
	if p == "" || strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return fmt.Errorf("invalid repository-relative glob %q", p)
	}
	for _, s := range strings.Split(p, "/") {
		if s == ".." {
			return fmt.Errorf("glob leaves repository: %q", p)
		}
		if s != "**" {
			if _, err := path.Match(s, "x"); err != nil {
				return fmt.Errorf("glob %q: %w", p, err)
			}
		}
	}
	return nil
}

// ** matches zero or more whole directories. Other segments use path.Match.
func glob(pattern, name string) bool { return render.Glob(pattern, name) }

func matches(patterns []string, name string) bool {
	for _, p := range patterns {
		if glob(p, name) {
			return true
		}
	}
	return false
}
