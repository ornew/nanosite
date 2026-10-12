package sitegen

import (
	"fmt"
	"github.com/ornew/nanosite/render"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
	"html"
	"html/template"
	"path"
	"regexp"
)

// Extensions adds trusted project-specific content to the standard pipeline.
// Markdown publication, routing, record indexes, navigation and output ownership
// remain controlled by the standard configuration and builder.
type Extensions struct {
	Validate      func(string) error
	Prepare       func(*Context) error
	Decorate      func(*Context, *Page, string) (string, error)
	View          func(View) (any, error)
	TemplateFuncs template.FuncMap
	Marker        string
	MarkerContent []byte
	Protected     []string
	AllowedOutput []string
}

type Context struct {
	Repo     string
	Config   Config
	Markdown goldmark.Markdown
	Pages    []*Page
	Assets   map[string][]byte
}

// AddHTMLPage registers trusted generated content before links are resolved.
// source is a logical Markdown path used for matching links and section rules.
func (c *Context) AddHTMLPage(source, output, title string, body []byte) (*Page, error) {
	if !matches(c.Config.Include, source) || matches(c.Config.Exclude, source) {
		return nil, nil
	}
	if err := relativeSource(source); err != nil {
		return nil, err
	}
	if path.Clean(output) != output || path.IsAbs(output) {
		return nil, fmt.Errorf("invalid generated page output: %s", output)
	}
	for _, p := range c.Pages {
		if p.Source == source {
			return nil, fmt.Errorf("duplicate generated source: %s", source)
		}
	}
	p, err := parsePage(source, []byte("# "+title+"\n"), c.Config, c.Markdown)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	p.Output = output
	p.RenderedHTML = append([]byte{}, body...)
	p.NoSource = true
	p.Headings = nil
	p.Anchors = map[string]bool{}
	for _, m := range generatedID.FindAllSubmatch(body, -1) {
		p.Anchors[html.UnescapeString(string(m[1]))] = true
	}
	p.Document = c.Markdown.Parser().Parse(text.NewReader(nil))
	c.Pages = append(c.Pages, p)
	return p, nil
}

var generatedID = regexp.MustCompile(`(?:^|[\s<])id="([^"]+)"`)

// BuildWith runs the same builder used by the CLI, with narrowly scoped hooks.
func BuildWith(repo, out string, c Config, check bool, extensions Extensions) (*Result, error) {
	return buildWith(repo, out, c, check, extensions)
}

func (c *Context) Render(node *Page) (string, error) { return renderMarkdown(c.Markdown, node) }
func HTMLText(body string) string                    { return render.PlainText(body) }
