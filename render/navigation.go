package render

// Adjacent selects neighbors in the supplied order, restricted by eligible.
// Themes may flatten multiple groups or restrict to the current section.
func Adjacent[T comparable](pages []T, current T, eligible func(T) bool) (previous, next T) {
	found := -1
	for i, p := range pages {
		if p == current {
			found = i
			break
		}
	}
	if found < 0 {
		return
	}
	for i := found - 1; i >= 0; i-- {
		if eligible == nil || eligible(pages[i]) {
			previous = pages[i]
			break
		}
	}
	for i := found + 1; i < len(pages); i++ {
		if eligible == nil || eligible(pages[i]) {
			next = pages[i]
			break
		}
	}
	return
}

type Group[T any] struct {
	Name              string
	Pages             []T
	Children          []Group[T]
	Collapsed, Active bool
}
type GroupRule[T any] struct {
	Name      string
	Match     func(T) bool
	Children  []GroupRule[T]
	Collapsed bool
}

// Navigation groups pages by configured rules and opens the current page's
// parent groups. It retains input page order within every group.
func Navigation[T comparable](pages []T, current T, rules []GroupRule[T]) []Group[T] {
	var groups []Group[T]
	for _, r := range rules {
		g := Group[T]{Name: r.Name, Collapsed: r.Collapsed}
		for _, p := range pages {
			if r.Match != nil && r.Match(p) {
				g.Pages = append(g.Pages, p)
				g.Active = g.Active || p == current
			}
		}
		g.Children = Navigation(g.Pages, current, r.Children)
		if len(g.Pages) > 0 || len(g.Children) > 0 {
			groups = append(groups, g)
		}
	}
	return groups
}
