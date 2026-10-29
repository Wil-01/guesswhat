package main

import (
	"testing"
	"testing/fstest"
)

func TestDefaultWords(t *testing.T) {
	seen := make(map[string]string)
	for name, list := range testWords.All() {
		for _, w := range list {
			if _, err := cleanWord(w); err != nil {
				t.Fatalf("bad word %q in %s", w, name)
			}
			if other, ok := seen[normalize(w)]; ok {
				t.Fatalf("%q is in both %s and %s", w, other, name)
			}
			seen[normalize(w)] = name
		}
	}
}

func TestLoadThemesErrors(t *testing.T) {
	if _, err := loadThemes(fstest.MapFS{}, "*.txt"); err != errNoThemes {
		t.Fatalf("got %v, want errNoThemes", err)
	}
	empty := fstest.MapFS{"empty.txt": {Data: []byte("\n  \n")}}
	if _, err := loadThemes(empty, "*.txt"); err == nil {
		t.Fatal("expected error for an empty theme")
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
