package contentfetch

import "github.com/hkstm/fccentrummap/internal/youtube"

func ExtractYouTubeVideoID(html string) (string, bool) {
	return youtube.ExtractVideoID(html)
}
