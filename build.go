package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ornew/nanosite/publish"
	"github.com/ornew/nanosite/render"
	"github.com/yuin/goldmark/ast"
)

//go:embed layout.html static/*
var embedded embed.FS

const marker = ".nanosite"
const markerContent = "nanosite v1\n"

type Result struct {
	Pages []*Page
	Out   string
}
type NavGroup struct {
	Title             string
	Collapsed, Active bool
	Pages             []*Page
}
type View struct {
	Styles, Scripts                                     []string
	Config                                              Config
	Page                                                *Page
	Body                                                template.HTML
	Groups                                              []NavGroup
	TOC                                                 []Heading
	Previous, Next                                      *Page
	Home, CSS, Script, Theme, Search, Source, Canonical string
	NoIndex                                             bool
	URL                                                 func(string) string
}

func inside(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func relative(from, to string) string {
	r, _ := filepath.Rel(filepath.FromSlash(path.Dir(from)), filepath.FromSlash(to))
	return filepath.ToSlash(r)
}
func encodePath(p string) string { return (&url.URL{Path: p}).EscapedPath() }
func pageURL(output string) string {
	if output == "index.html" {
		return ""
	}
	if strings.HasSuffix(output, "/index.html") {
		return strings.TrimSuffix(output, "index.html")
	}
	return output
}
func publicURL(c Config, output string) string { return c.BasePath + encodePath(pageURL(output)) }

func validateOutput(repo, out string) error {
	if out == repo || inside(out, repo) {
		return fmt.Errorf("output must not contain the input repository")
	}
	// Resolve symlinks before any ownership check; never replace a symlink itself.
	if info, err := os.Lstat(out); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output must not be a symlink")
	}
	entries, err := os.ReadDir(out)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(out, marker))
	if err != nil || string(data) != markerContent {
		return fmt.Errorf("refusing nonempty output without %s ownership marker: %s", marker, out)
	}
	return nil
}

func build(repo, out string, c Config, check bool) (*Result, error) {
	repo, err := canonicalPath(repo)
	if err != nil {
		return nil, err
	}
	originalOut, err := filepath.Abs(out)
	if err != nil {
		return nil, err
	}
	out, err = canonicalPath(out)
	if err != nil {
		return nil, err
	}
	if originalOut != out {
		return nil, fmt.Errorf("output path must not traverse symlinks")
	}
	if err := validateOutput(repo, out); err != nil {
		return nil, err
	}
	md := markdown()
	var pages []*Page
	assets := map[string]string{}
	err = filepath.WalkDir(repo, func(file string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(repo, file)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if file == out || strings.HasPrefix(d.Name(), ".site-build-") || d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		selected := matches(c.Include, rel) && !matches(c.Exclude, rel)
		asset := matches(c.Assets, rel) && !matches(c.Exclude, rel)
		if !selected && !asset {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("selected input must not be a symlink: %s", rel)
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("selected input is not a regular file: %s", rel)
		}
		if inside(out, file) {
			return fmt.Errorf("output would overwrite an input: %s", rel)
		}
		if selected {
			if !strings.EqualFold(path.Ext(rel), ".md") {
				return fmt.Errorf("include matched non-Markdown input: %s", rel)
			}
			data, e := os.ReadFile(file)
			if e != nil {
				return e
			}
			p, e := parsePage(rel, data, c, md)
			if e != nil {
				return e
			}
			if p != nil {
				pages = append(pages, p)
			}
		} else if asset {
			assets[rel] = file
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no published Markdown pages matched include patterns")
	}
	// An optional public directory maps files directly to output-relative paths.
	if c.StaticDir != "" {
		dir, e := localInput(repo, c.StaticDir)
		if e != nil {
			return nil, e
		}
		if inside(dir, out) || inside(out, dir) {
			return nil, fmt.Errorf("staticDir overlaps output")
		}
		info, e := os.Stat(dir)
		if e != nil {
			return nil, e
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("staticDir must be a directory")
		}
		e = filepath.WalkDir(dir, func(file string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("static input must not be a symlink: %s", file)
			}
			info, e := d.Info()
			if e != nil {
				return e
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("static input is not regular: %s", file)
			}
			rel, e := filepath.Rel(dir, file)
			if e != nil {
				return e
			}
			rel = filepath.ToSlash(rel)
			if _, exists := assets[rel]; exists {
				return fmt.Errorf("duplicate asset: %s", rel)
			}
			assets[rel] = file
			return nil
		})
		if e != nil {
			return nil, e
		}
	}
	// Never permit output to replace source directories, even if prior output has a marker.
	for _, p := range pages {
		if inside(out, filepath.Join(repo, filepath.FromSlash(p.Source))) {
			return nil, fmt.Errorf("output contains a source page")
		}
	}
	sort.Slice(pages, func(i, j int) bool {
		if weight(pages[i].Meta) != weight(pages[j].Meta) {
			return weight(pages[i].Meta) < weight(pages[j].Meta)
		}
		ai, aj := path.Base(pages[i].Source) == "README.md", path.Base(pages[j].Source) == "README.md"
		if ai != aj {
			return ai
		}
		return pages[i].Source < pages[j].Source
	})
	bySource := map[string]*Page{}
	byOutput := map[string]*Page{}
	generated := map[string][]byte{}
	reserve := func(name string, data []byte) error {
		name = path.Clean(name)
		if name == "." || name == ".." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") {
			return fmt.Errorf("invalid output path: %s", name)
		}
		if _, ok := generated[name]; ok {
			return fmt.Errorf("output collision: %s", name)
		}
		// A file cannot also be an ancestor directory of another generated file.
		for existing := range generated {
			if strings.HasPrefix(existing, name+"/") || strings.HasPrefix(name, existing+"/") {
				return fmt.Errorf("output file/directory collision: %s and %s", name, existing)
			}
		}
		generated[name] = data
		return nil
	}
	for _, p := range pages {
		if err := reserve(p.Output, nil); err != nil {
			return nil, err
		}
		bySource[p.Source] = p
		byOutput[p.Output] = p
	}
	for name, file := range assets {
		data, e := os.ReadFile(file)
		if e != nil {
			return nil, e
		}
		if err := reserve(name, data); err != nil {
			return nil, err
		}
	}
	static, err := fs.ReadDir(embedded, "static")
	if err != nil {
		return nil, err
	}
	for _, f := range static {
		data, e := embedded.ReadFile("static/" + f.Name())
		if e != nil {
			return nil, e
		}
		if err := reserve("assets/"+f.Name(), data); err != nil {
			return nil, err
		}
	}
	redirects := map[string]string{}
	for from, to := range c.Redirects {
		output, e := routeOutput(from)
		if e != nil {
			return nil, e
		}
		target, e := routeOutput(to)
		if e != nil {
			return nil, e
		}
		redirects[output] = target
	}
	for _, p := range pages {
		if p.Meta.Aliases != nil {
			for _, a := range *p.Meta.Aliases {
				from, e := routeOutput(a)
				if e != nil {
					return nil, fmt.Errorf("%s: alias: %w", p.Source, e)
				}
				if _, ok := redirects[from]; ok {
					return nil, fmt.Errorf("duplicate redirect: %s", a)
				}
				redirects[from] = p.Output
			}
		}
	}
	for from, to := range redirects {
		if byOutput[to] == nil {
			return nil, fmt.Errorf("redirect %s must target a published page: %s", from, to)
		}
		href := encodePath(relative(from, to))
		data := []byte(fmt.Sprintf("<!doctype html><html lang=\"%s\"><head><meta charset=\"utf-8\"><meta http-equiv=\"refresh\" content=\"0;url=%s\"><title>Page moved</title></head><body><a href=\"%s\">Continue to the new page</a></body></html>", html.EscapeString(c.Lang), html.EscapeString(href), html.EscapeString(href)))
		if err := reserve(from, data); err != nil {
			return nil, err
		}
	}
	// Resolve and validate every Markdown link before rendering or replacing output.
	for _, p := range pages {
		err = ast.Walk(p.Document, func(n ast.Node, enter bool) (ast.WalkStatus, error) {
			if !enter {
				return ast.WalkContinue, nil
			}
			switch l := n.(type) {
			case *ast.Link:
				dest, e := resolveLink(p, string(l.Destination), bySource, byOutput, assets, redirects, generated)
				if e != nil {
					return ast.WalkStop, e
				}
				l.Destination = []byte(dest)
			case *ast.Image:
				dest, e := resolveLink(p, string(l.Destination), bySource, byOutput, assets, redirects, generated)
				if e != nil {
					return ast.WalkStop, e
				}
				l.Destination = []byte(dest)
			}
			return ast.WalkContinue, nil
		})
		if err != nil {
			return nil, err
		}
	}
	layout, err := template.ParseFS(embedded, "layout.html")
	if err != nil {
		return nil, err
	}
	if c.Template != "" {
		file, e := localInput(repo, c.Template)
		if e != nil {
			return nil, e
		}
		if inside(out, file) {
			return nil, fmt.Errorf("template overlaps output")
		}
		data, e := os.ReadFile(file)
		if e != nil {
			return nil, e
		}
		layout, err = template.New("layout.html").Parse(string(data))
		if err != nil {
			return nil, err
		}
	}
	for _, name := range append(append([]string{}, c.Styles...), c.Scripts...) {
		if name == "" || strings.ContainsAny(name, "?#\\") || path.Clean(name) != name || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") {
			return nil, fmt.Errorf("styles/scripts require output-relative local paths: %q", name)
		}
		if _, ok := generated[name]; !ok {
			return nil, fmt.Errorf("missing styles/scripts asset: %s", name)
		}
	}
	home := "index.html"
	if byOutput[home] == nil {
		home = pages[0].Output
	}
	var search []map[string]string
	for _, p := range pages {
		body, e := renderMarkdown(md, p)
		if e != nil {
			return nil, e
		}
		groups := navigation(pages, p, c)
		v := View{Config: c, Page: p, Body: template.HTML(body), Groups: groups, Home: encodePath(relative(p.Output, home)),
			CSS: encodePath(relative(p.Output, "assets/site.css")), Script: encodePath(relative(p.Output, "assets/site.js")), Theme: encodePath(relative(p.Output, "assets/theme.js"))}
		v.NoIndex = boolValue(p.Meta.Draft, false)
		for _, name := range c.Styles {
			v.Styles = append(v.Styles, encodePath(relative(p.Output, name)))
		}
		for _, name := range c.Scripts {
			v.Scripts = append(v.Scripts, encodePath(relative(p.Output, name)))
		}
		v.URL = func(to string) string { return encodePath(relative(p.Output, to)) }
		if boolValue(p.Meta.TOC, c.TOC) {
			for _, h := range p.Headings {
				if h.Level >= 2 && h.Level <= 4 {
					v.TOC = append(v.TOC, h)
				}
			}
		}
		if c.Search {
			v.Search = encodePath(relative(p.Output, "assets/search-index.js"))
		}
		if c.Repository != "" {
			v.Source = strings.TrimSuffix(c.Repository, "/") + "/blob/" + encodePath(c.Branch) + "/" + encodePath(p.Source)
		}
		if c.SiteURL != "" {
			u, e := url.Parse(c.SiteURL)
			if e != nil || u.Scheme != "https" && u.Scheme != "http" || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || u.Path != "" && u.Path != "/" {
				return nil, fmt.Errorf("siteURL must be an absolute http(s) origin; set deployment prefix with basePath")
			}
			v.Canonical = strings.TrimSuffix(c.SiteURL, "/") + publicURL(c, p.Output)
		}
		for _, g := range groups {
			previous, next := render.Adjacent(g.Pages, p, nil)
			if previous != nil || next != nil {
				v.Previous, v.Next = previous, next
			}
		}

		var b strings.Builder
		if err := layout.ExecuteTemplate(&b, "layout.html", v); err != nil {
			return nil, err
		}
		generated[p.Output] = []byte(b.String())
		if c.Search && boolValue(p.Meta.Search, true) {
			search = append(search, map[string]string{"title": p.Title, "url": p.Output, "text": plainText(p)})
		}
	}
	if c.Search {
		data, e := json.Marshal(search)
		if e != nil {
			return nil, e
		}
		if err := reserve("assets/search-index.js", append(append([]byte("globalThis.NANOSITE_SEARCH = "), data...), ';', '\n')); err != nil {
			return nil, err
		}
	}
	if len(redirects) > 0 {
		var b strings.Builder
		keys := sortedKeys(redirects)
		for _, from := range keys {
			fmt.Fprintf(&b, "%s %s 301\n", publicURL(c, from), publicURL(c, redirects[from]))
		}
		if err := reserve("_redirects", []byte(b.String())); err != nil {
			return nil, err
		}
	}
	if c.SiteURL != "" {
		var b strings.Builder
		b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
		for _, p := range pages {
			if boolValue(p.Meta.Draft, false) {
				continue
			}
			fmt.Fprintf(&b, "<url><loc>%s</loc></url>\n", html.EscapeString(strings.TrimSuffix(c.SiteURL, "/")+publicURL(c, p.Output)))
		}
		b.WriteString("</urlset>\n")
		if err := reserve("sitemap.xml", []byte(b.String())); err != nil {
			return nil, err
		}
		if err := reserve("robots.txt", []byte("User-agent: *\nAllow: /\nSitemap: "+strings.TrimSuffix(c.SiteURL, "/")+c.BasePath+"sitemap.xml\n")); err != nil {
			return nil, err
		}
	}
	if err := reserve(marker, []byte(markerContent)); err != nil {
		return nil, err
	}
	result := &Result{pages, out}
	if check {
		return result, nil
	}
	if err := writeOutput(repo, out, generated); err != nil {
		return nil, err
	}
	return result, nil
}

func resolveLink(p *Page, raw string, bySource, byOutput map[string]*Page, assets map[string]string, redirects map[string]string, generated map[string][]byte) (string, error) {
	fail := func(message string) (string, error) { return "", fmt.Errorf("%s: link %q: %s", p.Source, raw, message) }
	u, err := url.Parse(raw)
	if err != nil {
		return fail("invalid URL")
	}
	if u.IsAbs() || u.Host != "" {
		return raw, nil
	}
	target := u.Path
	var dest string
	var page *Page
	if target == "" {
		dest = p.Output
		page = p
	} else {
		if strings.Contains(target, "\\") {
			return fail("backslash in local path")
		}
		source := path.Clean(path.Join(path.Dir(p.Source), target))
		output := path.Clean(path.Join(path.Dir(p.Output), target))
		if strings.HasPrefix(target, "/") {
			source = path.Clean(strings.TrimPrefix(target, "/"))
			output = source
		}
		if source == ".." || strings.HasPrefix(source, "../") {
			return fail("path leaves repository")
		}
		if page = bySource[source]; page != nil {
			dest = page.Output
		} else if page = bySource[path.Join(source, "README.md")]; page != nil {
			dest = page.Output
		} else if page = bySource[path.Join(source, "index.md")]; page != nil {
			dest = page.Output
		} else if _, ok := assets[source]; ok {
			dest = source
		} else {
			if strings.HasSuffix(target, "/") {
				output = path.Join(output, "index.html")
			}
			if page = byOutput[output]; page != nil {
				dest = output
			} else if to, ok := redirects[output]; ok {
				page = byOutput[to]
				dest = to
			} else if _, ok := generated[output]; ok {
				dest = output
			} else {
				return fail("missing, excluded or unpublished target (include an asset explicitly if needed)")
			}
		}
	}
	if page != nil && u.Fragment != "" && !page.Anchors[u.Fragment] {
		return fail("missing heading #" + u.Fragment)
	}
	u.Path = relative(p.Output, dest)
	u.RawPath = ""
	if target == "" {
		u.Path = ""
	}
	return u.String(), nil
}

func navigation(pages []*Page, current *Page, c Config) []NavGroup {
	var result []NavGroup
	seen := map[string]bool{}
	add := func(title string, collapsed bool) {
		g := NavGroup{Title: title, Collapsed: collapsed}
		for _, p := range pages {
			if p.Section == title && boolValue(p.Meta.Nav, true) {
				g.Pages = append(g.Pages, p)
				g.Active = g.Active || p == current
			}
		}
		if len(g.Pages) > 0 {
			result = append(result, g)
		}
		seen[title] = true
	}
	for _, s := range c.Sections {
		add(s.Title, s.Collapsed)
	}
	for _, p := range pages {
		if !seen[p.Section] {
			add(p.Section, false)
		}
	}
	return result
}

func plainText(p *Page) string {
	var b strings.Builder
	// Preserve inline token boundaries: adding a space after each Text node splits
	// identifiers such as PROJECT_NAME when Goldmark tokenizes punctuation.
	ast.Walk(p.Document, func(n ast.Node, enter bool) (ast.WalkStatus, error) {
		if !enter {
			if n.Type() == ast.TypeBlock {
				b.WriteByte(' ')
			}
			return ast.WalkContinue, nil
		}
		switch t := n.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(p.Body))
			if t.SoftLineBreak() || t.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(t.Value)
		case *ast.FencedCodeBlock:
			for i := 0; i < t.Lines().Len(); i++ {
				s := t.Lines().At(i)
				b.Write(s.Value(p.Body))
			}
		case *ast.CodeBlock:
			for i := 0; i < t.Lines().Len(); i++ {
				s := t.Lines().At(i)
				b.Write(s.Value(p.Body))
			}
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(strings.Fields(b.String()), " ")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func writeOutput(repo, out string, files map[string][]byte) error {
	return publish.Run(publish.Options{Repo: repo, Out: out, Marker: marker, MarkerContent: []byte(markerContent)}, func(stage string) error {
		for _, name := range sortedKeys(files) {
			if err := publish.WriteFile(filepath.Join(stage, filepath.FromSlash(name)), files[name]); err != nil {
				return err
			}
		}
		return nil
	})
}

// localInput ensures configured templates and public directories are read only
// from the input repository; symlinks cannot escape it or hide source ownership.
func localInput(repo, name string) (string, error) {
	if filepath.IsAbs(name) || strings.Contains(name, "\\") {
		return "", fmt.Errorf("input path must be repository-relative: %s", name)
	}
	file := filepath.Join(repo, filepath.FromSlash(name))
	if !inside(repo, file) {
		return "", fmt.Errorf("input leaves repository: %s", name)
	}
	real, err := canonicalPath(file)
	if err != nil {
		return "", err
	}
	if real != file {
		return "", fmt.Errorf("configured input must not traverse symlinks: %s", name)
	}
	return file, nil
}
