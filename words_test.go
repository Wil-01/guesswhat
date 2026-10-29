package main

import "testing"

func TestWordsLoaded(t *testing.T) {
	if len(words) == 0 {
		t.Fatal("no words loaded")
	}
	for _, w := range words {
		if normalize(w) == "" {
			t.Fatalf("bad word %q", w)
		}
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
