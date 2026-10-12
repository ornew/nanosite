package render

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
)

type Column struct {
	Title string `json:"title"`
	Field string `json:"field"`
}
type Record struct {
	Title, URL string
	Fields     map[string]string
}
type RecordIndex struct {
	IncludeTemplate                  bool
	Directory, Heading, Introduction string
	Columns                          []Column
	Required                         []string
	Template                         string
	TemplateText                     string
	GeneratedNotice                  string
	Select                           func(string) bool // optional repository-relative record selection
}

var recordName = regexp.MustCompile(`^([0-9]{3})-.+\.md$`)
var recordMetadata = regexp.MustCompile(`^\*\*([^*]+)\*\*:[ \t]*(.*?)[ \t]*\r?$`)

func ReadRecords(repo string, c RecordIndex) ([]Record, bool, error) {
	if c.Directory == "" || filepath.IsAbs(c.Directory) || strings.Contains(c.Directory, "\\") || strings.HasPrefix(filepath.Clean(c.Directory), "..") {
		return nil, false, fmt.Errorf("record directory must be repository-relative")
	}
	dir := filepath.Join(repo, filepath.FromSlash(c.Directory))
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var records []Record
	ids := map[string]string{}
	hasTemplate := false
	fields := map[string]bool{}
	for _, col := range c.Columns {
		fields[col.Field] = true
	}
	for _, key := range c.Required {
		fields[key] = true
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if c.Select != nil && !c.Select(filepath.ToSlash(filepath.Join(c.Directory, name))) {
			continue
		}
		if name == c.Template {
			hasTemplate = true
		}
		match := recordName.FindStringSubmatch(name)
		if match == nil || match[1] == "000" && !c.IncludeTemplate {
			continue
		}
		if previous, exists := ids[match[1]]; exists {
			return nil, false, fmt.Errorf("%s: duplicate record number %s in %s and %s", c.Directory, match[1], previous, name)
		}
		ids[match[1]] = name
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, false, fmt.Errorf("record must not be a symlink: %s", name)
		}
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, false, err
		}
		d := Parse(src, Options{})
		record := Record{URL: url.PathEscape(name), Fields: map[string]string{}}
		for n := d.AST.FirstChild(); n != nil; n = n.NextSibling() {
			if h, ok := n.(*ast.Heading); ok && h.Level == 1 {
				record.Title = strings.TrimSpace(string(h.Text(src)))
				break
			}
		}
		if record.Title == "" {
			return nil, false, fmt.Errorf("%s/%s: missing level-one title", c.Directory, name)
		}
		for n := d.AST.FirstChild(); n != nil; n = n.NextSibling() {
			if _, ok := n.(*ast.List); !ok {
				continue
			}
			for item := n.FirstChild(); item != nil; item = item.NextSibling() {
				block := item.FirstChild()
				if block == nil || block.Lines().Len() != 1 {
					continue
				}
				match := recordMetadata.FindSubmatch(bytes.TrimSpace(block.Lines().Value(src)))
				if match == nil {
					continue
				}
				key, value := string(match[1]), strings.TrimSpace(string(match[2]))
				if !fields[key] {
					continue
				}
				if _, exists := record.Fields[key]; exists {
					return nil, false, fmt.Errorf("%s/%s: duplicate %s metadata", c.Directory, name, key)
				}
				record.Fields[key] = value
			}
		}
		for _, key := range c.Required {
			if record.Fields[key] == "" {
				return nil, false, fmt.Errorf("%s/%s: missing %s metadata", c.Directory, name, key)
			}
		}
		records = append(records, record)
	}
	return records, hasTemplate, nil
}

func RecordIndexMarkdown(repo string, c RecordIndex) ([]byte, error) {
	if _, err := os.Stat(filepath.Join(repo, c.Directory)); os.IsNotExist(err) {
		return nil, nil
	}
	records, hasTemplate, err := ReadRecords(repo, c)
	if err != nil {
		return nil, err
	}
	return FormatRecordIndex(records, hasTemplate, c), nil
}

// FormatRecordIndex renders already-selected records without reading source files.
func FormatRecordIndex(records []Record, hasTemplate bool, c RecordIndex) []byte {
	var b strings.Builder
	if c.GeneratedNotice != "" {
		fmt.Fprintf(&b, "<!-- %s -->\n\n", c.GeneratedNotice)
	}
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", c.Heading, c.Introduction)
	b.WriteByte('|')
	for _, col := range c.Columns {
		fmt.Fprintf(&b, " %s |", TableText(col.Title))
	}
	b.WriteByte('\n')
	b.WriteByte('|')
	for range c.Columns {
		b.WriteString(" --- |")
	}
	b.WriteByte('\n')
	for _, r := range records {
		b.WriteByte('|')
		for _, col := range c.Columns {
			value := TableText(r.Fields[col.Field])
			if col.Field == "Title" {
				value = "[" + TableText(r.Title) + "](" + r.URL + ")"
			}
			fmt.Fprintf(&b, " %s |", value)
		}
		b.WriteByte('\n')
	}
	if hasTemplate && c.TemplateText != "" {
		fmt.Fprintf(&b, "\n%s\n", c.TemplateText)
	}
	return []byte(b.String())
}

func TableText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\r' {
			b.WriteByte(' ')
			continue
		}
		if strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// RecordTableHTML preserves caller-selected columns and ordering without making
// headings, introduction text or metadata keys project-specific.
func RecordTableHTML(records []Record, columns []Column, class string) string {
	var b strings.Builder
	if class != "" {
		fmt.Fprintf(&b, "<table class=\"%s\">", html.EscapeString(class))
	} else {
		b.WriteString("<table>")
	}
	b.WriteString("<thead><tr>")
	for _, c := range columns {
		fmt.Fprintf(&b, "<th>%s</th>", html.EscapeString(c.Title))
	}
	b.WriteString("</tr></thead><tbody>\n")
	for _, r := range records {
		b.WriteString("<tr>")
		for _, c := range columns {
			if c.Field == "Title" {
				fmt.Fprintf(&b, "<td><a href=\"%s\">%s</a></td>", html.EscapeString(r.URL), html.EscapeString(r.Title))
			} else {
				fmt.Fprintf(&b, "<td>%s</td>", html.EscapeString(r.Fields[c.Field]))
			}
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</tbody></table>\n")
	return b.String()
}
