package controller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/KonstantinK01/iptv/internal/dto"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	UserByLogin(ctx context.Context, login string) (dto.User, error)
}

type ChannelStore interface {
	ListActiveChannels(ctx context.Context) ([]dto.Channel, error)
	ActiveChannelByID(ctx context.Context, id int64) (dto.Channel, error)
}

type Deps struct {
	Users         UserStore
	Channels      ChannelStore
	StreamBaseURL string
	Log           *slog.Logger
}

type server struct {
	Deps
	dummyHash []byte
}

func NewHandler(d Deps) (http.Handler, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("generate dummy hash: %w", err)
	}
	s := &server{Deps: d, dummyHash: dummy}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/channels", onlyGET(http.HandlerFunc(s.listChannels)))
	mux.Handle("/api/v1/channels/{id}/playback", onlyGET(http.HandlerFunc(s.playback)))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	})

	return s.logging(s.recoverer(s.basicAuth(mux))), nil
}
