package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"unicode"
)

//go:embed words/*.txt
var defaultWords embed.FS

var errNoThemes = errors.New("no word themes found")

var accents = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i",
	"ô", "o", "ö", "o",
	"ù", "u", "û", "u", "ü", "u",
	"ç", "c", "œ", "oe",
)

func loadThemes(fsys fs.FS, pattern string) (map[string][]string, error) {
	files, err := fs.Glob(fsys, pattern)
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
		return nil, errNoThemes
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
