package main

import (
	"errors"
	"math/rand/v2"
)

const (
	avatarColors = 12
	avatarEyes   = 6
	avatarMouths = 6
)

var errInvalidAvatar = errors.New("invalid avatar")

// Avatar holds indexes into the palettes drawn by the front.
type Avatar struct {
	Color int `json:"color"`
	Eyes  int `json:"eyes"`
	Mouth int `json:"mouth"`
}

func randomAvatar() Avatar {
	return Avatar{
		Color: rand.IntN(avatarColors),
		Eyes:  rand.IntN(avatarEyes),
		Mouth: rand.IntN(avatarMouths),
	}
}

func (a Avatar) valid() bool {
	return a.Color >= 0 && a.Color < avatarColors &&
		a.Eyes >= 0 && a.Eyes < avatarEyes &&
		a.Mouth >= 0 && a.Mouth < avatarMouths
}
