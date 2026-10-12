package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/ornew/nanosite/render"
)

// All repository-specific choices live in Config. The generator can be moved
// without importing the adopted project's implementation or tooling.
type Config struct {
	RecordIndexes  []RecordIndexConfig `json:"recordIndexes"`
	Markdown       MarkdownOptions     `json:"markdown"`
	Pager          string              `json:"pager"`
	SearchHeadings bool                `json:"searchHeadings"`
	Template       string              `json:"template"`
	StaticDir      string              `json:"staticDir"`
	Styles         []string            `json:"styles"`
	Scripts        []string            `json:"scripts"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	Lang           string              `json:"lang"`
	Repository     string              `json:"repository"`
	Branch         string              `json:"branch"`
	SiteURL        string              `json:"siteURL"`
	BasePath       string              `json:"basePath"`
	Include        []string            `json:"include"`
	Exclude        []string            `json:"exclude"`
	Assets         []string            `json:"assets"`
	Drafts         bool                `json:"drafts"`
	PrettyURLs     bool                `json:"prettyURLs"`
	Search         bool                `json:"search"`
	TOC            bool                `json:"toc"`
	Sections       []Section           `json:"sections"`
	Defaults       []Defaults          `json:"defaults"`
	Redirects      map[string]string   `json:"redirects"`
}

type Section struct {
	Children  []Section `json:"children"`
	Index     string    `json:"index"`
	First     []string  `json:"first"`
	Title     string    `json:"title"`
	Match     []string  `json:"match"`
	Collapsed bool      `json:"collapsed"`
}

type MarkdownOptions struct {
	RawHTML         bool   `json:"rawHTML"`
	Alerts          bool   `json:"alerts"`
	HeadingLinks    bool   `json:"headingLinks"`
	HeadingIDs      string `json:"headingIDs"`
	AlignAttributes *bool  `json:"alignAttributes"`
}

type RecordIndexConfig struct {
	Source       string          `json:"source"`
	Directory    string          `json:"directory"`
	Title        string          `json:"title"`
	Introduction string          `json:"introduction"`
	Columns      []render.Column `json:"columns"`
	Required     []string        `json:"required"`
	Template     string          `json:"template"`
	TemplateText string          `json:"templateText"`
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
		PrettyURLs: true, Search: true, TOC: true, Pager: "section"}
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
	if c.Pager != "section" && c.Pager != "group" && c.Pager != "site" && c.Pager != "none" {
		return c, fmt.Errorf("pager must be section, group, site or none")
	}
	if c.Markdown.HeadingIDs != "" && c.Markdown.HeadingIDs != "unicode" && c.Markdown.HeadingIDs != "github" && c.Markdown.HeadingIDs != "goldmark" {
		return c, fmt.Errorf("markdown.headingIDs must be unicode, github or goldmark")
	}
	seen := map[string]bool{}
	patterns := append(append(append([]string{}, c.Include...), c.Exclude...), c.Assets...)
	var validateSections func([]Section) error
	validateSections = func(sections []Section) error {
		for _, section := range sections {
			if section.Title == "" || seen[section.Title] {
				return fmt.Errorf("section titles must be nonempty and globally unique: %q", section.Title)
			}
			seen[section.Title] = true
			patterns = append(patterns, section.Match...)
			if section.Index != "" {
				if err := relativeSource(section.Index); err != nil {
					return err
				}
			}
			for _, source := range section.First {
				if err := relativeSource(source); err != nil {
					return err
				}
			}
			if err := validateSections(section.Children); err != nil {
				return err
			}
		}
		return nil
	}
	if err := validateSections(c.Sections); err != nil {
		return c, err
	}
	indexes := map[string]bool{}
	for i := range c.RecordIndexes {
		index := &c.RecordIndexes[i]
		if err := relativeSource(index.Source); err != nil {
			return c, err
		}
		if path.Ext(index.Source) != ".md" || indexes[index.Source] {
			return c, fmt.Errorf("record index sources must be unique Markdown paths")
		}
		indexes[index.Source] = true
		if index.Directory == "" {
			index.Directory = path.Dir(index.Source)
		}
		if err := relativeSource(index.Directory); err != nil && index.Directory != "." {
			return c, err
		}
		if index.Title == "" || len(index.Columns) == 0 {
			return c, fmt.Errorf("record index requires title and columns")
		}
		for _, col := range index.Columns {
			if col.Title == "" || col.Field == "" {
				return c, fmt.Errorf("record index columns require title and field")
			}
		}
		if index.Template != "" && path.Base(index.Template) != index.Template {
			return c, fmt.Errorf("record template must be a filename")
		}
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

func relativeSource(source string) error {
	if source == "" || source == "." || source == ".." || path.Clean(source) != source || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "../") || strings.ContainsAny(source, "\\?#") {
		return fmt.Errorf("invalid repository-relative source path: %q", source)
	}
	return nil
}
