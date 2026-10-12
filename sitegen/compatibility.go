package sitegen

import (
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ornew/nanosite/publish"
	"github.com/ornew/nanosite/render"
)

func outputOptions(repo, out string, e Extensions) publish.Options {
	markerName, content := marker, []byte(markerContent)
	if e.Marker != "" {
		markerName, content = e.Marker, e.MarkerContent
	}
	return publish.Options{Repo: repo, Out: out, Marker: markerName, MarkerContent: content, Protected: e.Protected, AllowedOutput: e.AllowedOutput}
}
func validateOutputWith(repo, out string, e Extensions) error {
	if e.Marker == "" {
		if err := validateOutput(repo, out); err != nil {
			return err
		}
	}
	return publish.Validate(outputOptions(repo, out, e))
}
func writeOutputWith(repo, out string, files map[string][]byte, e Extensions) error {
	return publish.Run(outputOptions(repo, out, e), func(stage string) error {
		for _, name := range sortedKeys(files) {
			if err := publish.WriteFile(filepath.Join(stage, filepath.FromSlash(name)), files[name]); err != nil {
				return err
			}
		}
		if e.Validate != nil {
			return e.Validate(stage)
		}
		return nil
	})
}
func executeLayout(layout *template.Template, w io.Writer, v View, custom func(View) (any, error)) error {
	var data any = v
	if custom != nil {
		var err error
		data, err = custom(v)
		if err != nil {
			return err
		}
	}
	return layout.ExecuteTemplate(w, "layout.html", data)
}
func tocDepth(c Config) int {
	if c.TOCDepth == 0 {
		return 4
	}
	return c.TOCDepth
}
func indexTableClass(c Config, source string) string {
	for _, index := range c.RecordIndexes {
		if index.Source == source {
			return index.TableClass
		}
	}
	return ""
}

func resolveConfiguredLink(repo string, c Config, p *Page, raw string, bySource, byOutput map[string]*Page, assets map[string]string, redirects map[string]string, generated map[string][]byte) (string, error) {
	result, err := resolveLink(p, raw, bySource, byOutput, assets, redirects, generated)
	if err == nil {
		if c.DirectoryLinks {
			u, e := url.Parse(result)
			if e == nil && !u.IsAbs() && u.Host == "" && path.Base(u.Path) == "index.html" {
				u.Path = strings.TrimSuffix(u.Path, "index.html")
				if u.Path == "" {
					u.Path = "./"
				}
				u.RawPath = ""
				result = u.String()
			}
		}
		return result, nil
	}
	if !c.RepositoryLinks || c.Repository == "" {
		return "", err
	}
	u, e := url.Parse(raw)
	if e != nil || u.IsAbs() || u.Host != "" || u.Path == "" || strings.Contains(u.Path, "\\") {
		return "", err
	}
	source := path.Clean(path.Join(path.Dir(p.Source), u.Path))
	if path.IsAbs(u.Path) {
		source = path.Clean(strings.TrimPrefix(u.Path, "/"))
	}
	if source == ".." || strings.HasPrefix(source, "../") {
		return "", err
	}
	if bySource[source] != nil || bySource[path.Join(source, "README.md")] != nil || bySource[path.Join(source, "index.md")] != nil {
		return "", err
	}
	info, e := os.Stat(filepath.Join(repo, filepath.FromSlash(source)))
	if e != nil {
		return "", err
	}
	kind := "blob"
	if info.IsDir() {
		kind = "tree"
	}
	target := strings.TrimSuffix(c.Repository, "/") + "/" + kind + "/" + encodePath(c.Branch) + "/" + encodePath(source)
	if u.RawQuery != "" {
		target += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		target += (&url.URL{Fragment: u.Fragment}).String()
	}
	return target, nil
}

func searchPath(c Config) string {
	if c.SearchIndex.Path != "" {
		return c.SearchIndex.Path
	}
	return "assets/search-index.js"
}
func searchBytes(c Config, data []byte) []byte {
	variable := c.SearchIndex.Variable
	if variable == "" && strings.HasSuffix(searchPath(c), ".json") {
		return data
	}
	if variable == "" {
		variable = "globalThis.NANOSITE_SEARCH"
	}
	return append(append([]byte(variable+" = "), data...), ';', '\n')
}
func searchEntries(p *Page, body string, groups []NavGroup, c Config) []map[string]any {
	text := plainText(p)
	if c.SearchIndex.Text == "html" || p.RenderedHTML != nil {
		text = render.PlainText(body)
	}
	if p.SearchText != nil {
		text = *p.SearchText
	}
	target := encodePath(p.Output)
	if c.DirectoryLinks {
		target = encodePath(pageURL(p.Output))
	}
	section := p.Section
	if c.SearchIndex.Section == "group" {
		for _, group := range groups {
			for _, q := range group.allPages() {
				if q == p {
					section = group.Title
				}
			}
		}
	}
	headings := [][2]string{}
	for _, h := range p.Headings {
		if p.RenderedHTML != nil || h.Level >= 2 && h.Level <= tocDepth(c) {
			headings = append(headings, [2]string{h.ID, h.Title})
		}
	}
	fields := c.SearchIndex.Fields
	if fields == nil {
		fields = map[string]string{"title": "title", "url": "url", "text": "text"}
	}
	makeEntry := func(title, url, text string) map[string]any {
		values := map[string]any{"title": title, "url": url, "text": text, "section": section, "headings": headings}
		entry := map[string]any{}
		for field, key := range fields {
			entry[key] = values[field]
		}
		return entry
	}
	entries := []map[string]any{makeEntry(p.Title, target, text)}
	if c.SearchHeadings {
		for _, h := range p.Headings {
			if h.Level < 2 || h.Level > tocDepth(c) {
				continue
			}
			title := p.Title + " › " + h.Title
			if c.SearchIndex.HeadingTitle == "heading" {
				title = h.Title
			}
			headingText := h.Title
			if c.SearchIndex.HeadingText == "page" {
				headingText = text
			}
			entries = append(entries, makeEntry(title, target+(&url.URL{Fragment: h.ID}).String(), headingText))
		}
	}
	return entries
}

func validateExtensionAssets(context *Context) error {
	for source := range context.Assets {
		if path.IsAbs(source) || path.Clean(source) != source || source == ".." || strings.HasPrefix(source, "../") {
			return fmt.Errorf("invalid generated asset: %s", source)
		}
	}
	return nil
}
