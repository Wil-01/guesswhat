package main

import (
	"slices"
	"testing"
	"testing/fstest"
)

func TestThemesLoaded(t *testing.T) {
	if themesErr != nil {
		t.Fatal(themesErr)
	}
	seen := make(map[string]string)
	for name, list := range themes {
		for _, w := range list {
			n := normalize(w)
			if n == "" {
				t.Fatalf("bad word %q in %s", w, name)
			}
			if other, ok := seen[n]; ok {
				t.Fatalf("%q is in both %s and %s", w, other, name)
			}
			seen[n] = name
		}
	}
}

func TestLoadThemesErrors(t *testing.T) {
	if _, err := loadThemes(fstest.MapFS{}); err == nil {
		t.Fatal("expected error without themes")
	}
	empty := fstest.MapFS{"words/empty.txt": {Data: []byte("\n  \n")}}
	if _, err := loadThemes(empty); err == nil {
		t.Fatal("expected error for an empty theme")
	}
}

func TestRandomWord(t *testing.T) {
	for _, name := range themeNames() {
		word, theme := randomWord(name)
		if theme != name || !slices.Contains(themes[name], word) {
			t.Fatalf("randomWord(%q) = %q, %q", name, word, theme)
		}
	}
	word, theme := randomWord("")
	if !slices.Contains(themes[theme], word) {
		t.Fatalf("randomWord(\"\") = %q, %q", word, theme)
	}
}

func TestValidTheme(t *testing.T) {
	if !validTheme("") || !validTheme("animaux") || validTheme("nope") {
		t.Fatal("unexpected validTheme result")
	}
}

func TestNormalize(t *testing.T) {
	if got := normalize("  ÉLÉphant "); got != "elephant" {
		t.Fatalf("got %q", got)
	}
}

func TestHint(t *testing.T) {
	if got := hint("arc-en-ciel"); got != "___-__-____" {
		t.Fatalf("got %q", got)
	}
	if got := hint("château"); got != "_______" {
		t.Fatalf("got %q", got)
	}
}
