package ignore

import (
	"path/filepath"
	"strings"
)

type Matcher struct {
	patterns []string
}

func New(patterns []string) *Matcher {
	return &Matcher{patterns: patterns}
}

func (m *Matcher) Match(path string) bool {
	if len(m.patterns) == 0 {
		return false
	}

	cleaned := strings.ReplaceAll(path, "\\", "/")
	result := false

	for _, pattern := range m.patterns {
		p := strings.ReplaceAll(pattern, "\\", "/")

		negate := false
		if strings.HasPrefix(p, "!") {
			negate = true
			p = p[1:]
		}

		matched := m.matchPattern(p, cleaned)
		if matched {
			result = !negate
		}
	}

	return result
}

func (m *Matcher) matchPattern(p, cleaned string) bool {
	if strings.HasPrefix(p, "/") {
		anchor := p[1:]
		if matched, _ := filepath.Match(anchor, cleaned); matched {
			return true
		}
		if cleaned == anchor || strings.HasPrefix(cleaned, anchor+"/") {
			return true
		}
		return false
	}

	if strings.Contains(p, "/") {
		if cleaned == p || strings.HasPrefix(cleaned, p+"/") || strings.Contains(cleaned, "/"+p+"/") {
			return true
		}
	}

	if matched, _ := filepath.Match(p, filepath.Base(cleaned)); matched {
		return true
	}

	if strings.HasSuffix(p, "/") {
		prefix := p
		if strings.HasPrefix(cleaned, prefix) || strings.Contains(cleaned, "/"+prefix) {
			return true
		}
	}

	for part := range strings.SplitSeq(cleaned, "/") {
		if matched, _ := filepath.Match(p, part); matched {
			return true
		}
	}

	return false
}
