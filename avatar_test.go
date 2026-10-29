package main

import "testing"

func TestRandomAvatarIsValid(t *testing.T) {
	for range 100 {
		if a := randomAvatar(); !a.valid() {
			t.Fatalf("invalid random avatar %+v", a)
		}
	}
}

func TestAvatarValid(t *testing.T) {
	invalid := []Avatar{
		{Color: -1},
		{Color: avatarColors},
		{Eyes: avatarEyes},
		{Mouth: -1},
		{Mouth: avatarMouths},
	}
	for _, a := range invalid {
		if a.valid() {
			t.Errorf("%+v should be invalid", a)
		}
	}
	if !(Avatar{Color: avatarColors - 1, Eyes: avatarEyes - 1, Mouth: avatarMouths - 1}).valid() {
		t.Error("max avatar should be valid")
	}
}

func TestSetAvatar(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	want := Avatar{Color: 3, Eyes: 2, Mouth: 5}
	send(t, alice.Conn, AvatarPacket{Type: TypeAvatar, Avatar: want})
	for {
		players := expect(t, bob.Conn, TypePlayers).Players
		if players[0].Avatar == want {
			break
		}
	}
}

func TestInvalidAvatarRejected(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)

	send(t, alice.Conn, AvatarPacket{Type: TypeAvatar, Avatar: Avatar{Color: 99}})
	if got := expect(t, alice.Conn, TypeError).Message; got != errInvalidAvatar.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestPlayersHaveValidDefaultAvatar(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	connect(t, srv)

	for _, p := range expect(t, alice.Conn, TypePlayers).Players {
		if !p.Avatar.valid() {
			t.Fatalf("got invalid avatar %+v", p.Avatar)
		}
	}
}
