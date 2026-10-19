package controller

import (
	"net/url"
	"strings"
)

// {baseURL}/live/{slug}/index.m3u8.
func playlistURL(baseURL, slug string) string {
	return strings.TrimRight(baseURL, "/") + "/live/" + url.PathEscape(slug) + "/index.m3u8"
}
