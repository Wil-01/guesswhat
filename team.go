package main

import (
	"errors"
	"fmt"
)

const (
	teamCount      = 2
	minTeamPlayers = 2
)

var (
	errInvalidTeam   = errors.New("invalid team")
	errTeamsTooSmall = fmt.Errorf("each team needs at least %d players", minTeamPlayers)
	errNotYourTeam   = errors.New("only the drawer's team can guess")
)

func (r *Room) setTeam(c *Client, team int) error {
	if team < 1 || team > teamCount {
		return errInvalidTeam
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != StateWaiting {
		return errAlreadyStarted
	}
	c.team = team
	r.broadcastPlayers()
	return nil
}

// The helpers below must be called with r.mu held.

func (r *Room) teamSize(team int) int {
	n := 0
	for _, c := range r.clients {
		if c.team == team {
			n++
		}
	}
	return n
}

func (r *Room) smallestTeam() int {
	best := 1
	for team := 2; team <= teamCount; team++ {
		if r.teamSize(team) < r.teamSize(best) {
			best = team
		}
	}
	return best
}

func (r *Room) teamsReady() bool {
	for team := 1; team <= teamCount; team++ {
		if r.teamSize(team) < minTeamPlayers {
			return false
		}
	}
	return true
}

func (r *Room) teamScoreList() []int {
	if !r.teams {
		return nil
	}
	return append([]int(nil), r.teamScores[1:]...)
}

func nextTeam(team int) int {
	return team%teamCount + 1
}
