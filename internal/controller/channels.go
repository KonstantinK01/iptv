package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/KonstantinK01/iptv/internal/dto"
)

func (s *server) listChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := s.Channels.ListActiveChannels(r.Context())
	if err != nil {
		s.Log.Error("list channels", "err", err)
		internalError(w)
		return
	}
	resp := dto.ChannelsResponse{Channels: make([]dto.ChannelItem, 0, len(channels))}
	for _, c := range channels {
		resp.Channels = append(resp.Channels, dto.ChannelItem{ID: c.ID, Number: c.Number, Name: c.Name})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) playback(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_id", "channel id must be a positive integer")
		return
	}

	channel, err := s.Channels.ActiveChannelByID(r.Context(), id)
	if errors.Is(err, dto.ErrNotFound) {
		writeError(w, http.StatusNotFound, "channel_not_found", "channel not found")
		return
	}
	if err != nil {
		s.Log.Error("load channel", "id", id, "err", err)
		internalError(w)
		return
	}

	writeJSON(w, http.StatusOK, dto.PlaybackResponse{
		ChannelID: channel.ID,
		URL:       playlistURL(s.StreamBaseURL, channel.StreamSlug),
	})
}
