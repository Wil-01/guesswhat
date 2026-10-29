package main

import (
	"fmt"
	"time"
)

const (
	minRounds   = 1
	maxRounds   = 10
	minDuration = 30
	maxDuration = 180
)

var (
	errInvalidRounds   = fmt.Errorf("rounds must be between %d and %d", minRounds, maxRounds)
	errInvalidDuration = fmt.Errorf("duration must be between %d and %d seconds", minDuration, maxDuration)
)

// Settings are chosen by the admin before a game; Duration is in seconds.
type Settings struct {
	Theme    string `json:"theme"`
	Rounds   int    `json:"rounds"`
	Duration int    `json:"duration"`
}

func defaultSettings() Settings {
	return Settings{Rounds: defaultRounds, Duration: int(defaultRoundDuration / time.Second)}
}

func (s Settings) validate(words *WordStore) error {
	switch {
	case !words.Valid(s.Theme):
		return fmt.Errorf("%w %q", errUnknownTheme, s.Theme)
	case s.Rounds < minRounds || s.Rounds > maxRounds:
		return errInvalidRounds
	case s.Duration < minDuration || s.Duration > maxDuration:
		return errInvalidDuration
	}
	return nil
}
