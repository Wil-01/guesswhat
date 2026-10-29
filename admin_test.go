package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
)

const testPassword = "s3cret"

func newAdminServer(t *testing.T, password string) (*WordStore, *httptest.Server) {
	t.Helper()
	words, err := newMemoryWordStore()
	if err != nil {
		t.Fatal(err)
	}
	handler, err := newServer(newLobby(words), &Admin{words: words, password: password})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return words, srv
}

func adminRequest(t *testing.T, srv *httptest.Server, password, method, path string, body any) *http.Response {
	t.Helper()
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	req, err := http.NewRequest(method, srv.URL+path, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if password != "" {
		req.Header.Set("Authorization", "Bearer "+password)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func expectStatus(t *testing.T, res *http.Response, want int) {
	t.Helper()
	if res.StatusCode != want {
		t.Fatalf("%s %s: got status %d, want %d", res.Request.Method, res.Request.URL.Path, res.StatusCode, want)
	}
}

func TestAdminDisabledWithoutPassword(t *testing.T) {
	_, srv := newAdminServer(t, "")
	expectStatus(t, adminRequest(t, srv, "anything", "GET", "/api/admin/themes", nil), http.StatusServiceUnavailable)
}

func TestAdminRequiresPassword(t *testing.T) {
	_, srv := newAdminServer(t, testPassword)
	expectStatus(t, adminRequest(t, srv, "", "GET", "/api/admin/themes", nil), http.StatusUnauthorized)
	expectStatus(t, adminRequest(t, srv, "wrong", "GET", "/api/admin/themes", nil), http.StatusUnauthorized)

	res := adminRequest(t, srv, testPassword, "GET", "/api/admin/themes", nil)
	expectStatus(t, res, http.StatusOK)
	var all map[string][]string
	if err := json.NewDecoder(res.Body).Decode(&all); err != nil {
		t.Fatal(err)
	}
	if len(all["animaux"]) == 0 {
		t.Fatalf("got %v", all)
	}
}

func TestAdminManagesWords(t *testing.T) {
	words, srv := newAdminServer(t, testPassword)
	do := func(method, path string, body any, want int) {
		t.Helper()
		expectStatus(t, adminRequest(t, srv, testPassword, method, path, body), want)
	}

	do("POST", "/api/admin/themes", map[string]any{"name": "sport", "words": []string{"ballon"}}, http.StatusCreated)
	do("POST", "/api/admin/themes", map[string]any{"name": "sport", "words": []string{"filet"}}, http.StatusConflict)
	do("POST", "/api/admin/themes", map[string]any{"name": "Bad Name", "words": []string{"filet"}}, http.StatusBadRequest)
	do("POST", "/api/admin/themes/sport/words", map[string]string{"word": "ski de fond"}, http.StatusCreated)
	do("POST", "/api/admin/themes/sport/words", map[string]string{"word": "ballon"}, http.StatusConflict)
	do("POST", "/api/admin/themes/sport/words", map[string]string{"word": "42"}, http.StatusBadRequest)
	do("POST", "/api/admin/themes/nope/words", map[string]string{"word": "filet"}, http.StatusNotFound)
	do("POST", "/api/admin/themes/sport/words", "not an object", http.StatusBadRequest)
	if got := words.Words("sport"); !slices.Equal(got, []string{"ballon", "ski de fond"}) {
		t.Fatalf("got %v", got)
	}

	do("DELETE", "/api/admin/themes/sport/words/"+url.PathEscape("ski de fond"), nil, http.StatusNoContent)
	do("DELETE", "/api/admin/themes/sport/words/filet", nil, http.StatusNotFound)
	do("DELETE", "/api/admin/themes/sport/words/ballon", nil, http.StatusConflict)
	do("DELETE", "/api/admin/themes/sport", nil, http.StatusNoContent)
	do("DELETE", "/api/admin/themes/sport", nil, http.StatusNotFound)

	res := adminRequest(t, srv, "", "GET", "/api/themes", nil)
	var themes []string
	if err := json.NewDecoder(res.Body).Decode(&themes); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(themes, "sport") {
		t.Fatalf("got themes %v", themes)
	}
}

func TestNewThemeIsPlayable(t *testing.T) {
	words, srv := newAdminServer(t, testPassword)
	expectStatus(t, adminRequest(t, srv, testPassword, "POST", "/api/admin/themes",
		map[string]any{"name": "sport", "words": []string{"ballon"}}), http.StatusCreated)

	room := newRoom(words)
	if err := (Settings{Theme: "sport", Rounds: 1, Duration: 60}).validate(room.words); err != nil {
		t.Fatal(err)
	}
	if word, theme := words.Random("sport"); word != "ballon" || theme != "sport" {
		t.Fatalf("got %q, %q", word, theme)
	}
}
