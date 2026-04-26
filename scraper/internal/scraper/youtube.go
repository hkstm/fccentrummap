package scraper

import "regexp"

var youtubePatterns = []*regexp.Regexp{
	regexp.MustCompile(`youtube(?:-nocookie)?\.com/embed/([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`youtube\.com/watch\?[^\s"'<>]*v=([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`youtu\.be/([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`youtube\.com/shorts/([A-Za-z0-9_-]{11})`),
}

func ExtractYouTubeVideoID(html string) (string, bool) {
	for _, pattern := range youtubePatterns {
		match := pattern.FindStringSubmatch(html)
		if len(match) == 2 {
			return match[1], true
		}
	}
	return "", false
}
