package main

import (
	_ "embed"
	"math/rand/v2"
	"strings"
	"unicode"
)

//go:embed words.txt
var wordsFile string

var words = parseWords(wordsFile)

var accents = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i",
	"ô", "o", "ö", "o",
	"ù", "u", "û", "u", "ü", "u",
	"ç", "c", "œ", "oe",
)

func parseWords(s string) []string {
	var list []string
	for _, line := range strings.Split(s, "\n") {
		if w := strings.TrimSpace(line); w != "" {
			list = append(list, w)
		}
	}
	return list
}

func randomWord() string {
	return words[rand.IntN(len(words))]
}

func normalize(s string) string {
	return accents.Replace(strings.ToLower(strings.TrimSpace(s)))
}

func hint(word string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return '_'
		}
		return r
	}, word)
}
