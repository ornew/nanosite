package sitegen

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ornew/nanosite/render"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

// Index pages are virtual Markdown inputs; source documents are never rewritten.
func recordPages(repo string, c Config, pages []*Page, selected map[string]bool) ([]*Page, error) {
	published := map[string]*Page{}
	for _, p := range pages {
		published[p.Source] = p
	}
	for _, index := range c.RecordIndexes {
		if !matches(c.Include, index.Source) || matches(c.Exclude, index.Source) {
			continue
		}
		previous := published[index.Source]
		if selected[index.Source] && previous == nil {
			continue
		}
		directory, err := localInput(repo, index.Directory)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(directory); os.IsNotExist(err) && index.Optional {
			continue
		} else if err != nil {
			return nil, err
		}
		rule := render.RecordIndex{GeneratedNotice: index.Notice, IncludeTemplate: index.IncludeTemplate, Directory: index.Directory, Heading: index.Title, Introduction: index.Introduction, Columns: index.Columns, Required: index.Required, Template: index.Template, TemplateText: index.TemplateText, Select: func(source string) bool { return published[source] != nil }}
		records, hasTemplate, err := render.ReadRecords(repo, rule)
		if err != nil {
			return nil, err
		}
		// Records can have custom permalinks. Link to source paths so normal resolution
		// still checks publication and fragments instead of guessing an output route.
		for i := range records {
			source, err := filepath.Rel(filepath.FromSlash(filepath.Dir(index.Source)), filepath.FromSlash(filepath.Join(index.Directory, records[i].URL)))
			if err != nil {
				return nil, err
			}
			records[i].URL = filepath.ToSlash(source)
		}
		data := render.FormatRecordIndex(records, hasTemplate || index.Template == "", rule)
		if previous != nil {
			meta, err := yaml.Marshal(previous.Meta)
			if err != nil {
				return nil, err
			}
			data = append(append(append([]byte("---\n"), meta...), []byte("---\n")...), data...)
		}
		p, err := parsePage(index.Source, data, c, markdownWith(c.Markdown))
		if err != nil {
			return nil, err
		}
		if p == nil {
			continue
		}
		if previous != nil {
			for i, old := range pages {
				if old == previous {
					pages[i] = p
					break
				}
			}
		} else {
			pages = append(pages, p)
		}
		published[index.Source] = p
	}
	return pages, nil
}

func orderedNavigation(repo string, pages []*Page, current *Page, c Config) ([]NavGroup, error) {
	seen := map[string]bool{}
	var makeGroups func([]Section) ([]NavGroup, error)
	makeGroups = func(sections []Section) ([]NavGroup, error) {
		var result []NavGroup
		for _, s := range sections {
			seen[s.Title] = true
			g := NavGroup{Title: s.Title, Collapsed: s.Collapsed}
			for _, p := range pages {
				if p.Section == s.Title && boolValue(p.Meta.Nav, true) {
					g.Pages = append(g.Pages, p)
					g.Active = g.Active || p == current
				}
			}
			if s.Index != "" || len(s.First) > 0 {
				var body []byte
				if s.Index != "" {
					file, err := localInput(repo, s.Index)
					if err != nil {
						return nil, err
					}
					found := false
					for _, page := range pages {
						if page.Source == s.Index {
							body = page.Body
							found = true
							break
						}
					}
					if !found {
						body, err = os.ReadFile(file)
					}
					if err != nil {
						return nil, fmt.Errorf("navigation index: %w", err)
					}
				}
				rank := map[string]int{}
				prioritize := func(source string) {
					if source != "" && rank[source] == 0 {
						rank[source] = len(rank) + 1
					}
				}
				prioritize(s.Index)
				for _, source := range s.First {
					prioritize(source)
				}
				doc := markdown().Parser().Parse(text.NewReader(body))
				ast.Walk(doc, func(node ast.Node, enter bool) (ast.WalkStatus, error) {
					if link, ok := node.(*ast.Link); ok && enter {
						target, err := url.Parse(string(link.Destination))
						if err == nil && target.Scheme == "" && target.Host == "" && target.Path != "" {
							source := path.Clean(path.Join(path.Dir(s.Index), target.Path))
							if path.IsAbs(target.Path) {
								source = path.Clean(strings.TrimPrefix(target.Path, "/"))
							}
							prioritize(source)
							prioritize(path.Join(source, "README.md"))
							prioritize(path.Join(source, "index.md"))
						}
					}
					return ast.WalkContinue, nil
				})
				sort.SliceStable(g.Pages, func(i, j int) bool {
					a, b := g.Pages[i], g.Pages[j]
					if weight(a.Meta) != weight(b.Meta) {
						return weight(a.Meta) < weight(b.Meta)
					}
					ra, rb := rank[a.Source], rank[b.Source]
					if ra != rb {
						if ra == 0 {
							return false
						}
						if rb == 0 {
							return true
						}
						return ra < rb
					}
					introA, introB := path.Base(a.Source) == "README.md", path.Base(b.Source) == "README.md"
					if introA != introB {
						return introA
					}
					return a.Source < b.Source
				})
			}
			children, err := makeGroups(s.Children)
			if err != nil {
				return nil, err
			}
			g.Children = children
			for _, child := range children {
				g.Active = g.Active || child.Active
			}
			if len(g.Pages) > 0 || len(children) > 0 {
				result = append(result, g)
			}
		}
		return result, nil
	}
	groups, err := makeGroups(c.Sections)
	if err != nil {
		return nil, err
	}
	for _, p := range pages {
		if !seen[p.Section] {
			extra, e := makeGroups([]Section{{Title: p.Section}})
			if e != nil {
				return nil, e
			}
			groups = append(groups, extra...)
		}
	}
	return groups, nil
}

func (g NavGroup) allPages() []*Page {
	pages := append([]*Page{}, g.Pages...)
	for _, child := range g.Children {
		pages = append(pages, child.allPages()...)
	}
	return pages
}
func (g NavGroup) FirstPage() *Page {
	pages := g.allPages()
	if len(pages) == 0 {
		return nil
	}
	return pages[0]
}
func pageNeighbors(groups []NavGroup, current *Page, mode string) (*Page, *Page) {
	if mode == "none" || !boolValue(current.Meta.Nav, true) {
		return nil, nil
	}
	var all []*Page
	for _, group := range groups {
		pages := group.allPages()
		if mode == "group" {
			for _, p := range pages {
				if p == current {
					return render.Adjacent(pages, current, nil)
				}
			}
		}
		all = append(all, pages...)
	}
	if mode == "group" {
		return nil, nil
	}
	eligible := func(p *Page) bool { return mode == "site" || p.Section == current.Section }
	return render.Adjacent(all, current, eligible)
}
