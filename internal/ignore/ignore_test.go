package ignore

import "testing"

func TestMatchEmptyPatterns(t *testing.T) {
	m := New(nil)
	if m.Match("foo.go") {
		t.Error("empty matcher should not match anything")
	}

	m = New([]string{})
	if m.Match("foo.go") {
		t.Error("empty matcher should not match anything")
	}
}

func TestMatchSimpleGlob(t *testing.T) {
	m := New([]string{"*.log"})
	tests := []struct {
		path string
		want bool
	}{
		{"debug.log", true},
		{"error.log", true},
		{"foo.go", false},
		{"src/main.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchDirectoryName(t *testing.T) {
	m := New([]string{"node_modules"})
	tests := []struct {
		path string
		want bool
	}{
		{"node_modules", true},
		{"node_modules/react/index.js", true},
		{"src/node_modules/pkg", true},
		{"src/main.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchAbsolutePathPattern(t *testing.T) {
	m := New([]string{"/vendor"})
	tests := []struct {
		path string
		want bool
	}{
		{"vendor", true},
		{"vendor/pkg", true},
		{"src/vendor", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchTrailingSlash(t *testing.T) {
	m := New([]string{"dist/"})
	tests := []struct {
		path string
		want bool
	}{
		{"dist/output.js", true},
		{"build/dist/output.js", true},
		{"dist", false},
		{"distant/file.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchMultiplePatterns(t *testing.T) {
	m := New([]string{"*.log", "node_modules", ".git"})
	tests := []struct {
		path string
		want bool
	}{
		{"debug.log", true},
		{"node_modules/react/index.js", true},
		{".git/config", true},
		{"src/main.go", false},
		{"src/app_test.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchQuestionMark(t *testing.T) {
	m := New([]string{"test?.go"})
	tests := []struct {
		path string
		want bool
	}{
		{"tests.go", true},
		{"test2.go", true},
		{"test.go", false},
		{"testing.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchNegation(t *testing.T) {
	m := New([]string{"*.log", "!error.log"})
	tests := []struct {
		path string
		want bool
	}{
		{"debug.log", true},
		{"error.log", false},
		{"src/error.log", false},
		{"foo.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchNegationDirectory(t *testing.T) {
	m := New([]string{"node_modules", "!node_modules/important"})
	tests := []struct {
		path string
		want bool
	}{
		{"node_modules/react/index.js", true},
		{"node_modules/important/pkg.go", false},
		{"node_modules", true},
		{"src/main.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchNegationReinclude(t *testing.T) {
	m := New([]string{"*.log", "!*.log", "important.log"})
	tests := []struct {
		path string
		want bool
	}{
		{"debug.log", false},
		{"important.log", true},
		{"foo.go", false},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path); got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestMatchNegationOnly(t *testing.T) {
	m := New([]string{"!*.log"})
	if m.Match("debug.log") {
		t.Error("negation-only pattern with no prior match should not match")
	}
	if m.Match("foo.go") {
		t.Error("negation-only pattern should not match unrelated paths")
	}
}

func TestMatchWindowsPaths(t *testing.T) {
	m := New([]string{"node_modules", "*.log"})
	if !m.Match(`node_modules\react\index.js`) {
		t.Error("should match Windows-style paths")
	}
	if !m.Match(`src\debug.log`) {
		t.Error("should match glob against Windows path components")
	}
}
