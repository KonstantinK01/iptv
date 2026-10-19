package dto

type Channel struct {
	ID         int64
	Number     int
	Name       string
	StreamSlug string
	Active     bool
}

type ChannelItem struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Name   string `json:"name"`
}

type ChannelsResponse struct {
	Channels []ChannelItem `json:"channels"`
}

type PlaybackResponse struct {
	ChannelID int64  `json:"channel_id"`
	URL       string `json:"url"`
}
