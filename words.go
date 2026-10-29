package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"math/rand/v2"
	"path"
	"slices"
	"strings"
	"unicode"
)

//go:embed words/*.txt
var wordFiles embed.FS

var themes, themesErr = loadThemes(wordFiles)

var errUnknownTheme = errors.New("unknown theme")

var accents = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i",
	"ô", "o", "ö", "o",
	"ù", "u", "û", "u", "ü", "u",
	"ç", "c", "œ", "oe",
)

func loadThemes(fsys fs.FS) (map[string][]string, error) {
	files, err := fs.Glob(fsys, "words/*.txt")
	if err != nil {
		return nil, err
	}
	loaded := make(map[string][]string)
	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, err
		}
		list := parseWords(string(data))
		if len(list) == 0 {
			return nil, fmt.Errorf("%s: no words", file)
		}
		loaded[strings.TrimSuffix(path.Base(file), ".txt")] = list
	}
	if len(loaded) == 0 {
		return nil, errors.New("no word themes found")
	}
	return loaded, nil
}

func parseWords(s string) []string {
	var list []string
	for _, line := range strings.Split(s, "\n") {
		if w := strings.TrimSpace(line); w != "" {
			list = append(list, w)
		}
	}
	return list
}

func themeNames() []string {
	return slices.Sorted(maps.Keys(themes))
}

// validTheme accepts an empty theme, which means all themes mixed.
func validTheme(theme string) bool {
	_, ok := themes[theme]
	return ok || theme == ""
}

// randomWord returns a word and the theme it belongs to.
func randomWord(theme string) (string, string) {
	if list, ok := themes[theme]; ok {
		return list[rand.IntN(len(list))], theme
	}
	total := 0
	for _, list := range themes {
		total += len(list)
	}
	n := rand.IntN(total)
	for _, name := range themeNames() {
		if n < len(themes[name]) {
			return themes[name][n], name
		}
		n -= len(themes[name])
	}
	panic("unreachable")
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
