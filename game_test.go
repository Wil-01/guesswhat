package main

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func TestFirstClientIsAdmin(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	players := expect(t, alice.Conn, TypePlayers).Players
	if len(players) != 2 || !players[0].Admin || players[1].Admin || players[0].ID != alice.id || players[1].ID != bob.id {
		t.Fatalf("got players %+v", players)
	}
}

func TestAdminTransferredOnLeave(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	alice.Close()

	players := expect(t, bob.Conn, TypePlayers).Players
	if len(players) != 1 || players[0].ID != bob.id || !players[0].Admin {
		t.Fatalf("got players %+v", players)
	}
}

func TestStartRequiresAdmin(t *testing.T) {
	_, srv := newTestServer(t)
	connect(t, srv)
	bob := connect(t, srv)

	send(t, bob.Conn, Envelope{Type: TypeStart})
	if got := expect(t, bob.Conn, TypeError).Message; got != errNotAdmin.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestStartRequiresPlayers(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)

	send(t, alice.Conn, Envelope{Type: TypeStart})
	if got := expect(t, alice.Conn, TypeError).Message; got != errNotEnoughPlayers.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestStartTwiceRejected(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	startGame(t, alice, alice, bob)

	send(t, alice.Conn, Envelope{Type: TypeStart})
	if got := expect(t, alice.Conn, TypeError).Message; got != errAlreadyStarted.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestWordOnlySentToDrawer(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	r := startGame(t, alice, alice, bob)

	guess := r.packets[r.guesser()]
	if r.word == "" || guess.Word != "" || guess.Drawer != r.drawer.id {
		t.Fatalf("got drawer %+v and guesser %+v", r.packets[r.drawer], guess)
	}
	if guess.Hint != hint(r.word) {
		t.Fatalf("got hint %q for %q", guess.Hint, r.word)
	}
}

func TestOnlyDrawerCanDraw(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	r := startGame(t, alice, alice, bob)
	guesser := r.guesser()

	send(t, guesser.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, guesser.Conn, TypeError)
	send(t, guesser.Conn, Envelope{Type: TypeClear})
	expect(t, guesser.Conn, TypeError)

	send(t, r.drawer.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, guesser.Conn, TypeDraw)
}

func TestDrawerCannotRevealWord(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	r := startGame(t, alice, alice, bob)

	say(t, r.drawer.Conn, "it is a "+strings.ToUpper(r.word))
	expect(t, r.drawer.Conn, TypeError)
	say(t, r.drawer.Conn, "hint: it is big")
	expect(t, r.guesser().Conn, TypeMessage)
}

func TestLateJoinerGetsRound(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	r := startGame(t, alice, alice, bob)
	send(t, r.drawer.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, r.guesser().Conn, TypeDraw)

	carol := connect(t, srv)
	if len(carol.drawing) != 1 {
		t.Fatalf("got drawing %+v", carol.drawing)
	}
	if p := carol.round; p.Word != "" || p.Hint != hint(r.word) || p.Drawer != r.drawer.id {
		t.Fatalf("got round %+v", p)
	}
}

func TestCorrectGuessWins(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	r := startGame(t, alice, alice, bob)
	guesser := r.guesser()

	say(t, guesser.Conn, "  "+strings.ToUpper(normalize(r.word))+" ")
	for _, c := range []*testClient{alice, bob} {
		expect(t, c.Conn, TypeSystem)
		end := expect(t, c.Conn, TypeRoundEnd)
		if end.Word != r.word || end.Winner != guesser.id || end.Points < minPoints {
			t.Fatalf("got round end %+v", end)
		}
		ranking := expect(t, c.Conn, TypeGameEnd).Players
		if ranking[0].ID != guesser.id || ranking[0].Score != end.Points || ranking[1].Score != 0 {
			t.Fatalf("got ranking %+v", ranking)
		}
		expectClosed(t, c.Conn)
	}
}

func TestWrongGuessIsChat(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	r := startGame(t, alice, alice, bob)

	say(t, r.guesser().Conn, "definitely not the word 42")
	if got := expect(t, r.drawer.Conn, TypeMessage).Message; got != "definitely not the word 42" {
		t.Fatalf("got message %q", got)
	}
}

func TestTimeUp(t *testing.T) {
	room, srv := newTestServer(t)
	setRoom(room, func(r *Room) { r.roundDuration = 100 * time.Millisecond })
	alice := connect(t, srv)
	bob := connect(t, srv)
	startGame(t, alice, alice, bob)

	if got := expect(t, bob.Conn, TypeSystem).Message; got != "Time's up!" {
		t.Fatalf("got system %q", got)
	}
	if end := expect(t, bob.Conn, TypeRoundEnd); end.Winner != 0 || end.Points != 0 {
		t.Fatalf("got round end %+v", end)
	}
	expect(t, bob.Conn, TypeGameEnd)
}

func TestRoundsRotateDrawer(t *testing.T) {
	room, srv := newTestServer(t)
	setRoom(room, func(r *Room) { r.rounds = 3 })
	alice := connect(t, srv)
	bob := connect(t, srv)
	all := []*testClient{alice, bob}

	r := startGame(t, alice, all...)
	var previous *testClient
	for i := 1; i <= 3; i++ {
		p := r.packets[alice]
		if p.Round != i || p.Rounds != 3 {
			t.Fatalf("got round %d/%d, want %d/3", p.Round, p.Rounds, i)
		}
		if r.drawer == previous {
			t.Fatalf("round %d: same drawer twice in a row", i)
		}
		previous = r.drawer
		say(t, r.guesser().Conn, r.word)
		for _, c := range all {
			expect(t, c.Conn, TypeSystem)
			expect(t, c.Conn, TypeRoundEnd)
		}
		if i < 3 {
			r = expectRound(t, all...)
		}
	}
	ranking := expect(t, alice.Conn, TypeGameEnd).Players
	if ranking[0].Score+ranking[1].Score < 3*minPoints {
		t.Fatalf("got ranking %+v", ranking)
	}
}

func TestDrawerLeavingEndsRound(t *testing.T) {
	room, srv := newTestServer(t)
	setRoom(room, func(r *Room) { r.rounds = 2 })
	alice := connect(t, srv)
	bob := connect(t, srv)
	carol := connect(t, srv)
	r := startGame(t, alice, alice, bob, carol)

	var rest []*testClient
	for _, c := range []*testClient{alice, bob, carol} {
		if c != r.drawer {
			rest = append(rest, c)
		}
	}
	r.drawer.Close()
	for _, c := range rest {
		if got := expect(t, c.Conn, TypeSystem).Message; got != "The drawer has left." {
			t.Fatalf("got system %q", got)
		}
		expect(t, c.Conn, TypeRoundEnd)
	}
	expectRound(t, rest...)
}

func TestNotEnoughPlayersStopsGame(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	startGame(t, alice, alice, bob)

	bob.Close()
	expect(t, alice.Conn, TypeSystem)
	if got := expect(t, alice.Conn, TypeGame).State; got != StateWaiting {
		t.Fatalf("got state %q", got)
	}
}

func TestPointsFor(t *testing.T) {
	cases := []struct {
		remaining time.Duration
		want      int
	}{
		{80 * time.Second, maxPoints},
		{40 * time.Second, maxPoints / 2},
		{time.Second, minPoints},
		{0, minPoints},
	}
	for _, c := range cases {
		if got := pointsFor(c.remaining, 80*time.Second); got != c.want {
			t.Errorf("pointsFor(%v) = %d, want %d", c.remaining, got, c.want)
		}
	}
}

func TestThemesSentOnJoin(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	if alice.theme.Theme != "" || !slices.Equal(alice.theme.Themes, themeNames()) {
		t.Fatalf("got theme packet %+v", alice.theme)
	}
}

func TestSetTheme(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	send(t, alice.Conn, ThemePacket{Type: TypeTheme, Theme: "animaux"})
	for _, c := range []*testClient{alice, bob} {
		if got := expect(t, c.Conn, TypeTheme).Theme; got != "animaux" {
			t.Fatalf("got theme %q", got)
		}
	}
	r := startGame(t, alice, alice, bob)
	for c, p := range r.packets {
		if p.Theme != "animaux" {
			t.Fatalf("client %d got round theme %q", c.id, p.Theme)
		}
	}
	if !slices.Contains(themes["animaux"], r.word) {
		t.Fatalf("%q is not an animal", r.word)
	}
}

func TestMixedThemeTellsWordTheme(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	r := startGame(t, alice, alice, bob)

	theme := r.packets[r.guesser()].Theme
	if !slices.Contains(themes[theme], r.word) {
		t.Fatalf("%q is not in theme %q", r.word, theme)
	}
}

func TestSetThemeRejected(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	send(t, bob.Conn, ThemePacket{Type: TypeTheme, Theme: "animaux"})
	if got := expect(t, bob.Conn, TypeError).Message; got != errNotAdmin.Error() {
		t.Fatalf("got error %q", got)
	}
	send(t, alice.Conn, ThemePacket{Type: TypeTheme, Theme: "nope"})
	if got := expect(t, alice.Conn, TypeError).Message; !strings.HasPrefix(got, errUnknownTheme.Error()) {
		t.Fatalf("got error %q", got)
	}
	startGame(t, alice, alice, bob)
	send(t, alice.Conn, ThemePacket{Type: TypeTheme, Theme: "animaux"})
	if got := expect(t, alice.Conn, TypeError).Message; got != errAlreadyStarted.Error() {
		t.Fatalf("got error %q", got)
	}
}
