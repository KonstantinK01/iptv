package controller

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonstantinK01/iptv/internal/dto"
	"golang.org/x/crypto/bcrypt"
)

type fakeUsers map[string]dto.User

func (f fakeUsers) UserByLogin(_ context.Context, login string) (dto.User, error) {
	u, ok := f[login]
	if !ok {
		return dto.User{}, dto.ErrNotFound
	}
	return u, nil
}

type fakeChannels []dto.Channel

func (f fakeChannels) ListActiveChannels(context.Context) ([]dto.Channel, error) {
	var out []dto.Channel
	for _, c := range f {
		if c.Active {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f fakeChannels) ActiveChannelByID(_ context.Context, id int64) (dto.Channel, error) {
	for _, c := range f {
		if c.ID == id && c.Active {
			return c, nil
		}
	}
	return dto.Channel{}, dto.ErrNotFound
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	hash := func(p string) string {
		b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	h, err := NewHandler(Deps{
		Users: fakeUsers{
			"ivan":   {ID: 1, Login: "ivan", PasswordHash: hash("secret"), Active: true},
			"dmitry": {ID: 2, Login: "dmitry", PasswordHash: hash("secret"), Active: false},
		},
		Channels: fakeChannels{
			{ID: 1, Number: 1, Name: "One", StreamSlug: "ch1", Active: true},
			{ID: 2, Number: 2, Name: "Two", StreamSlug: "ch2", Active: true},
			{ID: 3, Number: 3, Name: "Off", StreamSlug: "ch3", Active: false},
		},
		StreamBaseURL: "http://stream.test/",
		Log:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func do(h http.Handler, method, path, login, password string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if login != "" {
		req.SetBasicAuth(login, password)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestStatusCodes(t *testing.T) {
	h := newTestHandler(t)

	tests := []struct {
		name           string
		method, path   string
		login, pass    string
		wantStatus     int
		wantAuthHeader bool
	}{
		{"no credentials", "GET", "/api/v1/channels", "", "", 401, true},
		{"wrong password", "GET", "/api/v1/channels", "ivan", "nope", 401, true},
		{"unknown user", "GET", "/api/v1/channels", "alexander", "secret", 401, true},
		{"inactive user", "GET", "/api/v1/channels", "dmitry", "secret", 401, true},
		{"list ok", "GET", "/api/v1/channels", "ivan", "secret", 200, false},
		{"playback ok", "GET", "/api/v1/channels/1/playback", "ivan", "secret", 200, false},
		{"playback without auth", "GET", "/api/v1/channels/1/playback", "", "", 401, true},
		{"id not a number", "GET", "/api/v1/channels/abc/playback", "ivan", "secret", 400, false},
		{"id zero", "GET", "/api/v1/channels/0/playback", "ivan", "secret", 400, false},
		{"id negative", "GET", "/api/v1/channels/-5/playback", "ivan", "secret", 400, false},
		{"id overflow", "GET", "/api/v1/channels/99999999999999999999/playback", "ivan", "secret", 400, false},
		{"channel missing", "GET", "/api/v1/channels/999/playback", "ivan", "secret", 404, false},
		{"channel inactive", "GET", "/api/v1/channels/3/playback", "ivan", "secret", 404, false},
		{"unknown route", "GET", "/api/v1/nothing", "ivan", "secret", 404, false},
		{"wrong method", "POST", "/api/v1/channels", "ivan", "secret", 405, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(h, tc.method, tc.path, tc.login, tc.pass)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", rec.Code, tc.wantStatus, rec.Body)
			}
			if got := rec.Header().Get("WWW-Authenticate") != ""; got != tc.wantAuthHeader {
				t.Errorf("WWW-Authenticate present = %v, want %v", got, tc.wantAuthHeader)
			}
			if tc.wantStatus >= 400 {
				var body dto.ErrorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error.Code == "" {
					t.Errorf("error response is not in the common format: %q", rec.Body)
				}
			}
		})
	}
}

func TestListChannelsHidesInactive(t *testing.T) {
	rec := do(newTestHandler(t), "GET", "/api/v1/channels", "ivan", "secret")

	var resp dto.ChannelsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Channels) != 2 {
		t.Fatalf("got %d channels, want 2: %+v", len(resp.Channels), resp.Channels)
	}
	if resp.Channels[0].Name != "One" || resp.Channels[1].Name != "Two" {
		t.Errorf("unexpected channels: %+v", resp.Channels)
	}
}

func TestPlaybackURL(t *testing.T) {
	rec := do(newTestHandler(t), "GET", "/api/v1/channels/1/playback", "ivan", "secret")

	var resp dto.PlaybackResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ChannelID != 1 {
		t.Errorf("channel_id = %d, want 1", resp.ChannelID)
	}
	if want := "http://stream.test/live/ch1/index.m3u8"; resp.URL != want {
		t.Errorf("url = %q, want %q", resp.URL, want)
	}
}
