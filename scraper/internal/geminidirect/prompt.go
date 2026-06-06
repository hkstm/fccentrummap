package geminidirect

import (
	"fmt"
	"strings"
)

type PromptInput struct {
	ArticleURL string
	YouTubeURL string
}

func BuildPrompt(input PromptInput) (string, error) {
	articleURL := strings.TrimSpace(input.ArticleURL)
	youtubeURL := strings.TrimSpace(input.YouTubeURL)
	if articleURL == "" {
		return "", fmt.Errorf("cannot build Gemini-direct prompt without article URL")
	}
	if youtubeURL == "" {
		return "", fmt.Errorf("cannot build Gemini-direct prompt without YouTube URL")
	}

	var b strings.Builder
	b.WriteString("You are an assistant extracting 'De spots van' location recommendations from public URL context.\n")
	b.WriteString("Use the article_url and youtube_url below as the primary source inputs.\n")
	b.WriteString("Goal: find the primary presenter and the Amsterdam spots that are actually recommended in the article/video.\n")
	b.WriteString("Return only places that you can support with evidence or notes. Do not invent places, addresses, or timestamps.\n")
	b.WriteString("Presenter guidance: use the article title and video title to determine the canonical presenter name and spelling; prefer the title form when it appears to represent the name the person is known by.\n")
	b.WriteString("Address guidance: include an address for a spot only when reasonably confident. Many videos show transition cards with the spot name, and the subtitle often contains the address, but not always. Do not guess; omit address or return null if unsure.\n")
	b.WriteString("Timestamps are YouTube timestamps in seconds from the start of the video.\n")
	b.WriteString("\n[source_inputs]\n")
	b.WriteString("article_url: ")
	b.WriteString(articleURL)
	b.WriteString("\n")
	b.WriteString("youtube_url: ")
	b.WriteString(youtubeURL)
	b.WriteString("\n\n")
	b.WriteString("Return only JSON matching exactly this response shape:\n")
	b.WriteString("{\n")
	b.WriteString("  \"article_url\": \"")
	b.WriteString(articleURL)
	b.WriteString("\",\n")
	b.WriteString("  \"youtube_url\": \"")
	b.WriteString(youtubeURL)
	b.WriteString("\",\n")
	b.WriteString("  \"model\": \"<configured model identifier if known>\",\n")
	b.WriteString("  \"presenter_name\": \"<optional primary presenter name>\",\n")
	b.WriteString("  \"spots\": [\n")
	b.WriteString("    {\"place\": \"<place name>\", \"address\": \"<optional detected address or null>\", \"youtubeTimestampSeconds\": 123, \"evidence\": \"<source fragment or short rationale>\", \"confidence\": 0.0}\n")
	b.WriteString("  ]\n")
	b.WriteString("}\n")
	b.WriteString("If the presenter is unknown, omit presenter_name or use an empty string. If an address, timestamp, or confidence is unknown, omit that field or use null instead of guessing.\n")
	return b.String(), nil
}
