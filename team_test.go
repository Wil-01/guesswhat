package main

import (
	"encoding/json"
	"slices"
	"testing"
)

// newTeamGame connects four players split into teams 1, 2, 1, 2.
func newTeamGame(t *testing.T, rounds int) (*Room, []*testClient) {
	t.Helper()
	room, srv := newTestServer(t)
	setRoom(room, func(r *Room) { r.teams, r.rounds = true, rounds })
	var all []*testClient
	for range 4 {
		all = append(all, connect(t, srv))
	}
	return room, all
}

func teamOf(all []*testClient, c *testClient) int {
	return slices.Index(all, c)%2 + 1
}

func teamScores(t *testing.T, p anyPacket) []int {
	t.Helper()
	var scores []int
	if err := json.Unmarshal(p.Teams, &scores); err != nil {
		t.Fatalf("got teams %s: %v", p.Teams, err)
	}
	return scores
}

func TestTeamsBalancedOnJoin(t *testing.T) {
	room, srv := newTestServer(t)
	setRoom(room, func(r *Room) { r.teams = true })
	alice := connect(t, srv)
	connect(t, srv)
	connect(t, srv)

	players := expect(t, alice.Conn, TypePlayers).Players
	expect(t, alice.Conn, TypePlayers)
	for i, p := range players {
		if p.Team != i%2+1 {
			t.Fatalf("got players %+v", players)
		}
	}
}

func TestTeamsHiddenWhenDisabled(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	connect(t, srv)

	p := expect(t, alice.Conn, TypePlayers)
	if p.Players[0].Team != 0 || p.Players[1].Team != 0 || p.Teams != nil {
		t.Fatalf("got players %+v", p)
	}
}

func TestSwitchTeam(t *testing.T) {
	_, all := newTeamGame(t, 1)
	send(t, all[0].Conn, TeamPacket{Type: TypeTeam, Team: 2})
	for {
		if players := expect(t, all[3].Conn, TypePlayers).Players; players[0].Team == 2 {
			break
		}
	}

	send(t, all[0].Conn, TeamPacket{Type: TypeTeam, Team: 3})
	if got := expect(t, all[0].Conn, TypeError).Message; got != errInvalidTeam.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestStartRequiresFullTeams(t *testing.T) {
	room, srv := newTestServer(t)
	setRoom(room, func(r *Room) { r.teams = true })
	alice := connect(t, srv)
	connect(t, srv)
	connect(t, srv)

	send(t, alice.Conn, Envelope{Type: TypeStart})
	if got := expect(t, alice.Conn, TypeError).Message; got != errTeamsTooSmall.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestOnlyDrawerTeamCanGuess(t *testing.T) {
	_, all := newTeamGame(t, 2)
	r := startGame(t, all[0], all...)
	team := teamOf(all, r.drawer)
	if r.packets[all[0]].Team != team {
		t.Fatalf("got round %+v, drawer team %d", r.packets[all[0]], team)
	}

	var mate, rival *testClient
	for _, c := range all {
		switch {
		case c == r.drawer:
		case teamOf(all, c) == team:
			mate = c
		default:
			rival = c
		}
	}
	say(t, rival.Conn, r.word)
	if got := expect(t, rival.Conn, TypeError).Message; got != errNotYourTeam.Error() {
		t.Fatalf("got error %q", got)
	}

	say(t, mate.Conn, r.word)
	for _, c := range all {
		expect(t, c.Conn, TypeSystem)
		end := expect(t, c.Conn, TypeRoundEnd)
		scores := teamScores(t, expect(t, c.Conn, TypePlayers))
		if end.Winner != mate.id || scores[team-1] != end.Points || scores[2-team] != 0 {
			t.Fatalf("got round end %+v and scores %v", end, scores)
		}
	}

	next := expectRound(t, all...)
	if got := teamOf(all, next.drawer); got == team || next.packets[all[0]].Team != got {
		t.Fatalf("team %d drew twice", team)
	}
}

func TestTeamTooSmallStopsGame(t *testing.T) {
	_, all := newTeamGame(t, 1)
	r := startGame(t, all[0], all...)

	var leaving *testClient
	for _, c := range all[1:] {
		if c != r.drawer {
			leaving = c
			break
		}
	}
	leaving.Close()
	for _, c := range all {
		if c == leaving {
			continue
		}
		if got := expect(t, c.Conn, TypeSystem).Message; got != "A team doesn't have enough players left." {
			t.Fatalf("got system %q", got)
		}
		if got := expect(t, c.Conn, TypeGame).State; got != StateWaiting {
			t.Fatalf("got state %q", got)
		}
	}
}

func TestSettingsTeams(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)

	s := defaultSettings()
	s.Teams = true
	send(t, alice.Conn, SettingsPacket{Type: TypeSettings, Settings: s})
	if got := string(expect(t, alice.Conn, TypeSettings).Teams); got != "true" {
		t.Fatalf("got teams %s", got)
	}
	if got := expect(t, alice.Conn, TypePlayers).Players[0].Team; got != 1 {
		t.Fatalf("got team %d", got)
	}
}
