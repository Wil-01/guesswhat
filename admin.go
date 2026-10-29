package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const maxBodySize = 64 << 10

var errAdminDisabled = errors.New("admin is disabled, set ADMIN_PASSWORD to enable it")

type Admin struct {
	words    *WordStore
	password string
}

func (a *Admin) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/themes", a.auth(a.list))
	mux.HandleFunc("POST /api/admin/themes", a.auth(a.createTheme))
	mux.HandleFunc("DELETE /api/admin/themes/{theme}", a.auth(a.deleteTheme))
	mux.HandleFunc("POST /api/admin/themes/{theme}/words", a.auth(a.addWord))
	mux.HandleFunc("DELETE /api/admin/themes/{theme}/words/{word}", a.auth(a.deleteWord))
}

func (a *Admin) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if a.password == "" {
			http.Error(w, errAdminDisabled.Error(), http.StatusServiceUnavailable)
			return
		}
		token, ok := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(token), []byte(a.password)) != 1 {
			http.Error(w, "wrong password", http.StatusUnauthorized)
			return
		}
		next(w, req)
	}
}

func (a *Admin) list(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, a.words.All())
}

func (a *Admin) createTheme(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Name  string   `json:"name"`
		Words []string `json:"words"`
	}
	if !readJSON(w, req, &body) {
		return
	}
	if err := a.words.AddTheme(strings.TrimSpace(body.Name), body.Words); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a.words.Words(body.Name))
}

func (a *Admin) deleteTheme(w http.ResponseWriter, req *http.Request) {
	if err := a.words.DeleteTheme(req.PathValue("theme")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *Admin) addWord(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Word string `json:"word"`
	}
	if !readJSON(w, req, &body) {
		return
	}
	theme := req.PathValue("theme")
	if err := a.words.AddWord(theme, body.Word); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a.words.Words(theme))
}

func (a *Admin) deleteWord(w http.ResponseWriter, req *http.Request) {
	if err := a.words.DeleteWord(req.PathValue("theme"), req.PathValue("word")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readJSON(w http.ResponseWriter, req *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, maxBodySize)).Decode(v); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeStoreError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, errUnknownTheme), errors.Is(err, errWordNotFound):
		status = http.StatusNotFound
	case errors.Is(err, errThemeExists), errors.Is(err, errWordExists),
		errors.Is(err, errLastTheme), errors.Is(err, errLastWord):
		status = http.StatusConflict
	case errors.Is(err, errInvalidTheme), errors.Is(err, errInvalidWord), errors.Is(err, errEmptyTheme):
		status = http.StatusBadRequest
	default:
		fmt.Fprintln(os.Stderr, "word store error:", err)
	}
	http.Error(w, err.Error(), status)
}
