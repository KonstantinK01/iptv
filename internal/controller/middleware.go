package controller

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/KonstantinK01/iptv/internal/dto"
	"golang.org/x/crypto/bcrypt"
)

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="iptv", charset="UTF-8"`)
	writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
}

func (s *server) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login, password, ok := r.BasicAuth()
		if !ok || !validLogin(login) {
			unauthorized(w)
			return
		}

		hash := s.dummyHash
		var user dto.User
		found := false

		u, err := s.Users.UserByLogin(r.Context(), login)
		switch {
		case err == nil:
			user, hash, found = u, []byte(u.PasswordHash), true
		case errors.Is(err, dto.ErrNotFound):
		default:
			s.Log.Error("load user", "err", err)
			internalError(w)
			return
		}

		passwordOK := bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
		if !found || !user.Active || !passwordOK {
			unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validLogin(login string) bool {
	return login != "" && len(login) <= 128 && utf8.ValidString(login) && !strings.ContainsRune(login, 0)
}


type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (s *server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.Log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
		)
	})
}

func (s *server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.Log.Error("panic in handler", "panic", rec)
				internalError(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func onlyGET(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}
