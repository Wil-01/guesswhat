package main

import (
	"errors"
	"io/fs"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const maxWordLen = 30

var (
	errUnknownTheme  = errors.New("unknown theme")
	errThemeExists   = errors.New("theme already exists")
	errWordExists    = errors.New("word already exists")
	errWordNotFound  = errors.New("word not found")
	errInvalidTheme  = errors.New("theme names use 1 to 30 lowercase letters, digits or dashes")
	errInvalidWord   = errors.New("words use 1 to 30 letters, spaces, dashes or apostrophes")
	errEmptyTheme    = errors.New("a theme needs at least one word")
	errLastTheme     = errors.New("the last theme cannot be deleted")
	errLastWord      = errors.New("the last word of a theme cannot be deleted, delete the theme instead")
	themeNamePattern = regexp.MustCompile(`^[a-z0-9-]{1,30}$`)
)

// WordStore holds the themed words, persisted as one file per theme when dir is set.
type WordStore struct {
	mu     sync.RWMutex
	dir    string
	themes map[string][]string
}

func newMemoryWordStore() (*WordStore, error) {
	themes, err := loadThemes(defaultWords, "words/*.txt")
	if err != nil {
		return nil, err
	}
	return &WordStore{themes: themes}, nil
}

// newWordStore loads the themes from dir, seeding it with the default words when it has none.
func newWordStore(dir string) (*WordStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	themes, err := loadThemes(os.DirFS(dir), "*.txt")
	if err == nil {
		return &WordStore{dir: dir, themes: themes}, nil
	}
	if !errors.Is(err, errNoThemes) {
		return nil, err
	}
	s, err := newMemoryWordStore()
	if err != nil {
		return nil, err
	}
	s.dir = dir
	for name := range s.themes {
		if err := s.save(name); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *WordStore) Themes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Sorted(maps.Keys(s.themes))
}

// Valid accepts an empty theme, which means all themes mixed.
func (s *WordStore) Valid(theme string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.themes[theme]
	return ok || theme == ""
}

func (s *WordStore) Words(theme string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.themes[theme])
}

func (s *WordStore) All() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make(map[string][]string, len(s.themes))
	for name, list := range s.themes {
		all[name] = slices.Clone(list)
	}
	return all
}

// Random returns a word and its theme; an unknown or empty theme picks among all words.
func (s *WordStore) Random(theme string) (string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if list, ok := s.themes[theme]; ok {
		return list[rand.IntN(len(list))], theme
	}
	total := 0
	for _, list := range s.themes {
		total += len(list)
	}
	n := rand.IntN(total)
	for _, name := range slices.Sorted(maps.Keys(s.themes)) {
		if n < len(s.themes[name]) {
			return s.themes[name][n], name
		}
		n -= len(s.themes[name])
	}
	panic("unreachable")
}

func (s *WordStore) AddTheme(name string, words []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !themeNamePattern.MatchString(name) {
		return errInvalidTheme
	}
	if _, ok := s.themes[name]; ok {
		return errThemeExists
	}
	if len(words) == 0 {
		return errEmptyTheme
	}
	var list []string
	for _, w := range words {
		w, err := cleanWord(w)
		if err != nil {
			return err
		}
		if s.contains(w) || slices.ContainsFunc(list, func(o string) bool { return normalize(o) == normalize(w) }) {
			return errWordExists
		}
		list = append(list, w)
	}
	s.themes[name] = list
	if err := s.save(name); err != nil {
		delete(s.themes, name)
		return err
	}
	return nil
}

func (s *WordStore) DeleteTheme(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.themes[name]; !ok {
		return errUnknownTheme
	}
	if len(s.themes) == 1 {
		return errLastTheme
	}
	if s.dir != "" {
		err := os.Remove(filepath.Join(s.dir, name+".txt"))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	delete(s.themes, name)
	return nil
}

func (s *WordStore) AddWord(theme, word string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.themes[theme]
	if !ok {
		return errUnknownTheme
	}
	word, err := cleanWord(word)
	if err != nil {
		return err
	}
	if s.contains(word) {
		return errWordExists
	}
	s.themes[theme] = append(slices.Clone(old), word)
	if err := s.save(theme); err != nil {
		s.themes[theme] = old
		return err
	}
	return nil
}

func (s *WordStore) DeleteWord(theme, word string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.themes[theme]
	if !ok {
		return errUnknownTheme
	}
	i := slices.IndexFunc(old, func(w string) bool { return normalize(w) == normalize(word) })
	if i < 0 {
		return errWordNotFound
	}
	if len(old) == 1 {
		return errLastWord
	}
	s.themes[theme] = slices.Delete(slices.Clone(old), i, i+1)
	if err := s.save(theme); err != nil {
		s.themes[theme] = old
		return err
	}
	return nil
}

// The helpers below must be called with s.mu held.

func (s *WordStore) contains(word string) bool {
	for _, list := range s.themes {
		if slices.ContainsFunc(list, func(w string) bool { return normalize(w) == normalize(word) }) {
			return true
		}
	}
	return false
}

func (s *WordStore) save(theme string) error {
	if s.dir == "" {
		return nil
	}
	file := filepath.Join(s.dir, theme+".txt")
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(s.themes[theme], "\n")+"\n"), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, file)
}

func cleanWord(w string) (string, error) {
	w = strings.ToLower(strings.Join(strings.Fields(w), " "))
	if w == "" || utf8.RuneCountInString(w) > maxWordLen {
		return "", errInvalidWord
	}
	for _, r := range w {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' {
			return "", errInvalidWord
		}
	}
	return w, nil
}
