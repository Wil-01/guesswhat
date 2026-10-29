package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func newDiskStore(t *testing.T) (*WordStore, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "words")
	s, err := newWordStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s, dir
}

func reload(t *testing.T, dir string) *WordStore {
	t.Helper()
	s, err := newWordStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestStoreSeedsEmptyDir(t *testing.T) {
	s, dir := newDiskStore(t)
	if !slices.Equal(s.Themes(), testWords.Themes()) {
		t.Fatalf("got themes %v", s.Themes())
	}
	if _, err := os.Stat(filepath.Join(dir, "animaux.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestStoreRandom(t *testing.T) {
	for _, name := range testWords.Themes() {
		word, theme := testWords.Random(name)
		if theme != name || !slices.Contains(testWords.Words(name), word) {
			t.Fatalf("Random(%q) = %q, %q", name, word, theme)
		}
	}
	for _, name := range []string{"", "deleted-theme"} {
		word, theme := testWords.Random(name)
		if !slices.Contains(testWords.Words(theme), word) {
			t.Fatalf("Random(%q) = %q, %q", name, word, theme)
		}
	}
}

func TestStoreValid(t *testing.T) {
	if !testWords.Valid("") || !testWords.Valid("animaux") || testWords.Valid("nope") {
		t.Fatal("unexpected Valid result")
	}
}

func TestStoreThemeLifecycle(t *testing.T) {
	s, dir := newDiskStore(t)
	if err := s.AddTheme("sport", []string{" Ballon ", "raquette"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddWord("sport", "Ski   de fond"); err != nil {
		t.Fatal(err)
	}
	want := []string{"ballon", "raquette", "ski de fond"}
	if got := reload(t, dir).Words("sport"); !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	if err := s.DeleteWord("sport", "RAQUETTE"); err != nil {
		t.Fatal(err)
	}
	if got := reload(t, dir).Words("sport"); !slices.Equal(got, []string{"ballon", "ski de fond"}) {
		t.Fatalf("got %v", got)
	}

	if err := s.DeleteTheme("sport"); err != nil {
		t.Fatal(err)
	}
	if reload(t, dir).Valid("sport") {
		t.Fatal("theme still on disk")
	}
}

func TestStoreErrors(t *testing.T) {
	s, err := newMemoryWordStore()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"bad theme name", s.AddTheme("Sport!", []string{"ballon"}), errInvalidTheme},
		{"existing theme", s.AddTheme("animaux", []string{"ballon"}), errThemeExists},
		{"empty theme", s.AddTheme("sport", nil), errEmptyTheme},
		{"duplicate in new theme", s.AddTheme("sport", []string{"but", "BUT"}), errWordExists},
		{"word from another theme", s.AddTheme("sport", []string{"elephant"}), errWordExists},
		{"bad word", s.AddWord("animaux", "chat2"), errInvalidWord},
		{"long word", s.AddWord("animaux", "abcdefghijklmnopqrstuvwxyzabcde"), errInvalidWord},
		{"existing word", s.AddWord("objets", "Éléphant"), errWordExists},
		{"word in unknown theme", s.AddWord("nope", "ballon"), errUnknownTheme},
		{"unknown word", s.DeleteWord("animaux", "ballon"), errWordNotFound},
		{"unknown theme", s.DeleteTheme("nope"), errUnknownTheme},
	}
	for _, c := range cases {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, c.err, c.want)
		}
	}
}

func TestStoreKeepsLastThemeAndWord(t *testing.T) {
	s, err := newMemoryWordStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddTheme("solo", []string{"unique"}); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteWord("solo", "unique"); !errors.Is(err, errLastWord) {
		t.Fatalf("got %v", err)
	}
	for _, name := range s.Themes()[1:] {
		if err := s.DeleteTheme(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.DeleteTheme(s.Themes()[0]); !errors.Is(err, errLastTheme) {
		t.Fatalf("got %v", err)
	}
}

func TestStoreRejectsBrokenDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "empty.txt"), []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := newWordStore(dir); err == nil {
		t.Fatal("expected an error")
	}
}
