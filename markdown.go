package main

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode"

	"github.com/ornew/nanosite/render"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

type Heading struct {
	ID, Title string
	Level     int
}
type Page struct {
	Source, Output, Title, Description, NavTitle, Section, Layout string
	Meta                                                          Metadata
	Body                                                          []byte
	Document                                                      ast.Node
	Headings                                                      []Heading
	Anchors                                                       map[string]bool
}

func frontMatter(src []byte) (Metadata, []byte, error) {
	var m Metadata
	s := strings.ReplaceAll(strings.TrimPrefix(string(src), "\ufeff"), "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return m, []byte(s), nil
	}
	lines := strings.Split(s, "\n")
	end := 1
	for end < len(lines) && lines[end] != "---" {
		end++
	}
	if end == len(lines) {
		return m, nil, fmt.Errorf("unclosed YAML front matter")
	}
	d := yaml.NewDecoder(strings.NewReader(strings.Join(lines[1:end], "\n")))
	d.KnownFields(true)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return m, nil, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return m, nil, fmt.Errorf("front matter must be a single mapping")
	}
	return m, []byte(strings.Join(lines[end+1:], "\n")), nil
}

func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsSpace(r):
			b.WriteByte('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.M, r):
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "section"
	}
	return b.String()
}

func parsePage(source string, data []byte, c Config, md goldmark.Markdown) (*Page, error) {
	m, body, err := frontMatter(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}
	defaults := Metadata{}
	for _, d := range c.Defaults {
		if matches(d.Match, source) {
			defaults = mergeMetadata(defaults, d.Values)
		}
	}
	m = mergeMetadata(defaults, m)
	if !boolValue(m.Publish, true) || (boolValue(m.Draft, false) && !c.Drafts) {
		return nil, nil
	}
	p := &Page{Source: source, Meta: m, Body: body, Document: md.Parser().Parse(text.NewReader(body)), Anchors: map[string]bool{}}
	err = ast.Walk(p.Document, func(n ast.Node, enter bool) (ast.WalkStatus, error) {
		if !enter {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok {
			title := string(h.Text(body))
			id := slug(title)
			if c.Markdown.HeadingIDs == "github" {
				d := render.Document{Markdown: md, Source: body}
				title = d.Text(h)
				id = render.Slug(title)
			}
			if c.Markdown.HeadingIDs == "goldmark" {
				if value, ok := h.AttributeString("id"); ok {
					id = string(value.([]byte))
				}
			}
			base := id
			for i := 1; p.Anchors[id]; i++ {
				id = fmt.Sprintf("%s-%d", base, i)
			}
			p.Anchors[id] = true
			h.SetAttributeString("id", []byte(id))
			p.Headings = append(p.Headings, Heading{id, title, h.Level})
			if h.Level == 1 && p.Title == "" {
				p.Title = title
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}
	p.Title = textValue(m.Title, p.Title)
	if p.Title == "" {
		p.Title = strings.TrimSuffix(path.Base(source), path.Ext(source))
	}
	p.NavTitle = textValue(m.NavTitle, p.Title)
	p.Description = textValue(m.Description, c.Description)
	p.Layout = textValue(m.Layout, "doc")
	if p.Layout != "home" && p.Layout != "doc" {
		return nil, fmt.Errorf("%s: layout must be home or doc", source)
	}
	p.Section = textValue(m.Section, "")
	if p.Section == "" {
		p.Section = matchSection(source, c.Sections)
	}
	if p.Section == "" {
		p.Section = "Other"
	}
	if m.Permalink != nil {
		p.Output, err = routeOutput(*m.Permalink)
	} else {
		p.Output = pageOutput(source, c.PrettyURLs)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}
	return p, nil
}

func markdown() goldmark.Markdown { return markdownWith(MarkdownOptions{}) }
func markdownWith(o MarkdownOptions) goldmark.Markdown {
	return render.NewMarkdown(render.Options{RawHTML: o.RawHTML, AlignAttributes: boolValue(o.AlignAttributes, true), AutomaticHeadingIDs: o.HeadingIDs == "goldmark"})
}
func matchSection(source string, sections []Section) string {
	for _, section := range sections {
		if child := matchSection(source, section.Children); child != "" {
			return child
		}
		if matches(section.Match, source) {
			return section.Title
		}
	}
	return ""
}

func renderMarkdown(md goldmark.Markdown, p *Page) (string, error) {
	var b bytes.Buffer
	err := md.Renderer().Render(&b, p.Body, p.Document)
	return b.String(), err
}

func pageOutput(source string, pretty bool) string {
	if strings.EqualFold(path.Base(source), "README.md") || strings.EqualFold(path.Base(source), "index.md") {
		return path.Join(path.Dir(source), "index.html")
	}
	stem := strings.TrimSuffix(source, path.Ext(source))
	if pretty {
		return path.Join(stem, "index.html")
	}
	return stem + ".html"
}

func routeOutput(route string) (string, error) {
	if route == "" || strings.ContainsAny(route, "?#\\") {
		return "", fmt.Errorf("route must be a nonempty local path: %q", route)
	}
	route = strings.TrimPrefix(route, "/")
	for _, s := range strings.Split(route, "/") {
		if s == ".." || s == "." {
			return "", fmt.Errorf("route must not contain traversal: %q", route)
		}
	}
	if route == "" {
		return "index.html", nil
	}
	if strings.HasSuffix(route, "/") {
		return path.Join(route, "index.html"), nil
	}
	if !strings.HasSuffix(route, ".html") {
		return "", fmt.Errorf("route must end in / or .html: %q", route)
	}
	return route, nil
}
