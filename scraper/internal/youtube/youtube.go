package youtube

import (
	"fmt"
	"regexp"
)

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`youtube(?:-nocookie)?\.com/embed/([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`youtube\.com/watch\?[^\s"'<>]*v=([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`youtu\.be/([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`youtube\.com/shorts/([A-Za-z0-9_-]{11})`),
}

func ExtractVideoID(html string) (string, bool) {
	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(html)
		if len(match) == 2 {
			return match[1], true
		}
	}
	return "", false
}

func WatchURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}
