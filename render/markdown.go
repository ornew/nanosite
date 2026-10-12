// Package render supplies Markdown and record-index rendering for Nanosite
// and project extensions, without assuming a project language or theme.
package render

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Heading struct {
	Level    int
	ID, Text string
}
type Options struct {
	RawHTML              bool
	AlignAttributes      bool
	AutomaticHeadingIDs  bool // Goldmark IDs instead of GitHub-compatible Unicode IDs
	Alerts, HeadingLinks bool
	ResolveLink          func(string) string
}

type Document struct {
	Markdown goldmark.Markdown
	AST      ast.Node
	Source   []byte
	Headings []Heading
	options  Options
}

func NewMarkdown(o Options) goldmark.Markdown {
	var extensions []goldmark.Extender
	if o.AlignAttributes {
		extensions = append(extensions, extension.NewTable(extension.WithTableCellAlignMethod(extension.TableCellAlignAttribute)))
	} else {
		extensions = append(extensions, extension.Table)
	}
	extensions = append(extensions, extension.Strikethrough, extension.Linkify, extension.TaskList)
	opts := []goldmark.Option{goldmark.WithExtensions(extensions...)}
	if o.RawHTML {
		opts = append(opts, goldmark.WithRendererOptions(gmhtml.WithUnsafe()))
	}
	if o.AutomaticHeadingIDs {
		opts = append(opts, goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	}
	return goldmark.New(opts...)
}

func Parse(src []byte, o Options) *Document {
	d := &Document{Source: src, options: o}
	d.Markdown = NewMarkdown(o)
	d.Markdown.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(transformer{d}, 100)))
	d.AST = d.Markdown.Parser().Parse(text.NewReader(src))
	return d
}

type transformer struct{ d *Document }

func (t transformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	ids := map[string]int{}
	ast.Walk(doc, func(n ast.Node, enter bool) (ast.WalkStatus, error) {
		if !enter {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			title := t.d.Text(n)
			id := ""
			if t.d.options.AutomaticHeadingIDs {
				if v, ok := n.AttributeString("id"); ok {
					id = string(v.([]byte))
				}
			}
			if id == "" {
				id = Slug(title)
				if k := ids[id]; k > 0 {
					ids[id]++
					id = fmt.Sprintf("%s-%d", id, k)
				} else {
					ids[id] = 1
				}
				n.SetAttributeString("id", []byte(id))
			}
			t.d.Headings = append(t.d.Headings, Heading{n.Level, id, title})
		case *ast.Link:
			if t.d.options.ResolveLink != nil {
				n.Destination = []byte(t.d.options.ResolveLink(string(n.Destination)))
			}
		case *ast.Image:
			if t.d.options.ResolveLink != nil {
				n.Destination = []byte(t.d.options.ResolveLink(string(n.Destination)))
			}
		}
		return ast.WalkContinue, nil
	})
}

func (d *Document) Render(n ast.Node) (string, error) {
	if n == nil {
		n = d.AST
	}
	var b bytes.Buffer
	if err := d.Markdown.Renderer().Render(&b, d.Source, n); err != nil {
		return "", err
	}
	return PostProcess(b.String(), d.options.Alerts, d.options.HeadingLinks), nil
}
func (d *Document) Title() string {
	for _, h := range d.Headings {
		if h.Level == 1 {
			return h.Text
		}
	}
	return ""
}
func (d *Document) TOC(min, max int) []Heading {
	var hs []Heading
	for _, h := range d.Headings {
		if h.Level >= min && h.Level <= max {
			hs = append(hs, h)
		}
	}
	return hs
}
func (d *Document) Text(n ast.Node) string {
	var b bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if err := d.Markdown.Renderer().Render(&b, d.Source, c); err != nil {
			panic(err)
		}
	}
	s := tagRe.ReplaceAllString(anchorRe.ReplaceAllString(b.String(), ""), "")
	return strings.TrimSpace(spaceRe.ReplaceAllString(html.UnescapeString(s), " "))
}
func Slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.M, r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

var (
	alertRe   = regexp.MustCompile(`<blockquote>\s*<p>\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*`)
	headingRe = regexp.MustCompile(`<h([2-4]) id="([^"]+)">`)
	anchorRe  = regexp.MustCompile(`<a class="anchor"[^>]*>#</a>`)
	tagRe     = regexp.MustCompile(`<[^>]*>`)
	spaceRe   = regexp.MustCompile(`\s+`)
)

func PostProcess(s string, alerts, links bool) string {
	if alerts {
		s = alertRe.ReplaceAllStringFunc(s, func(m string) string {
			kind := alertRe.FindStringSubmatch(m)[1]
			label := kind[:1] + strings.ToLower(kind[1:])
			return fmt.Sprintf(`<blockquote class="callout callout-%s"><p class="callout-title">%s</p><p>`, strings.ToLower(kind), label)
		})
	}
	if links {
		s = headingRe.ReplaceAllString(s, `<h$1 id="$2"><a class="anchor" href="#$2" aria-hidden="true">#</a>`)
	}
	return s
}
func PlainText(s string) string {
	s = anchorRe.ReplaceAllString(s, "")
	return strings.TrimSpace(spaceRe.ReplaceAllString(html.UnescapeString(tagRe.ReplaceAllString(s, " ")), " "))
}
