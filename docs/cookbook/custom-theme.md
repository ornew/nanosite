# Add a custom theme

Start with an additional stylesheet when you only need a brand color or a small
layout adjustment. Replace the full HTML template only when the page structure
needs to change; a custom template takes responsibility for the site controls
and responsive behavior it replaces.

## Add a CSS override

Put this file at `site/public/assets/custom.css`:

```css
:root {
  --accent: #6d3ec1;
  --accent-soft: #efe8fb;
}

:root[data-theme="dark"] {
  --accent: #c1a4ff;
  --accent-soft: #302344;
}

/* Follow the user's system preference unless they selected a theme. */
@media (prefers-color-scheme: dark) {
  :root:not([data-theme="light"]) {
    --accent: #c1a4ff;
    --accent-soft: #302344;
  }
}
```

Configure the public directory and stylesheet in `site/config.json`:

```json
{
  "staticDir": "site/public",
  "styles": ["assets/custom.css"]
}
```

`staticDir` copies the contents of `site/public` to the output root. Stylesheet
paths are relative to that output root and load after the built-in stylesheet.
Using the built-in CSS variables and keeping explicit and system dark-mode
values aligned preserves the theme switcher and contrast across color modes.
Give custom assets distinct names so they do not collide with bundled assets.

## Replace the HTML layout

For a complete layout change, create `site/layout.html`:

```html
<!doctype html>
<html lang="{{.Config.Lang}}">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light dark">
  <title>{{.Page.Title}} · {{.Config.Title}}</title>
  <meta name="description" content="{{.Page.Description}}">
  {{if .Canonical}}<link rel="canonical" href="{{.Canonical}}">{{end}}
  {{if .NoIndex}}<meta name="robots" content="noindex">{{end}}
  <script src="{{.Theme}}"></script>
  <link rel="stylesheet" href="{{.CSS}}">
  {{range .Styles}}<link rel="stylesheet" href="{{.}}">{{end}}
</head>
<body class="layout-{{.Page.Layout}}">
  <header>
    <a href="{{.Home}}">{{.Config.Title}}</a>
    <button id="theme" type="button">Theme: system</button>
    {{if .Config.Repository}}<a href="{{.Config.Repository}}">Source</a>{{end}}
  </header>
  <details id="navigation" open>
    <summary>Navigation</summary>
    <nav aria-label="Documentation">
      {{range .Groups}}
        <section><h2>{{.Title}}</h2><ul>
          {{range .Pages}}<li><a href="{{call $.URL .Output}}">{{.NavTitle}}</a></li>{{end}}
        </ul></section>
      {{end}}
    </nav>
  </details>
  {{if .Search}}
    <label for="search">Search</label>
    <input id="search" type="search" autocomplete="off" aria-controls="search-results">
    <div id="search-results" role="region" aria-live="polite" hidden></div>
  {{end}}
  <main id="main">
    <article>{{.Body}}</article>
    {{if .TOC}}<nav aria-label="On this page"><ul>{{range .TOC}}
      <li class="level-{{.Level}}"><a href="#{{.ID}}">{{.Title}}</a></li>
    {{end}}</ul></nav>{{end}}
    <nav aria-label="Adjacent pages">
      {{if .Previous}}<a href="{{call .URL .Previous.Output}}">← {{.Previous.NavTitle}}</a>{{end}}
      {{if .Next}}<a href="{{call .URL .Next.Output}}">{{.Next.NavTitle}} →</a>{{end}}
    </nav>
  </main>
  {{if .Search}}<script src="{{.Search}}"></script>{{end}}
  <script src="{{.Script}}" defer></script>
  {{range .Scripts}}<script src="{{.}}" defer></script>{{end}}
</body>
</html>
```

Set `"template": "site/layout.html"` in the configuration. The template uses
Go's `html/template`; `.Body` contains trusted rendered Markdown, while metadata
is escaped normally. Use `{{call .URL "index.html"}}` to build a path relative
to the current generated page. The `.Theme`, `.CSS`, `.Search`, `.Script`, and
`.Styles`/`.Scripts` values are escaped paths relative to the current output
page, so use those values instead of hard-coding root-relative asset paths.

A custom layout replaces the default navigation and other markup. Keep the
`#navigation` details element, theme script, stylesheet, theme button, search
elements, and client script hooks shown above if those features are enabled.
The built-in script expects `#navigation` and `#theme`; search also expects
`#search-results`. The sample navigation lists top-level groups; extend it to
render nested `.Children` if the configuration uses child sections. The page
body stays in `.Body`; do not render Markdown a second time. Verify the finished
site at desktop and mobile widths and in light, dark, and system color modes.
