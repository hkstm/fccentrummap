package contentfetch

import "testing"

func TestExtractYouTubeVideoID(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
		ok   bool
	}{
		{
			name: "embed url",
			html: `<iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ"></iframe>`,
			want: "dQw4w9WgXcQ",
			ok:   true,
		},
		{
			name: "embed url nocookie",
			html: `<iframe src="https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ"></iframe>`,
			want: "dQw4w9WgXcQ",
			ok:   true,
		},
		{
			name: "watch url",
			html: `<a href="https://www.youtube.com/watch?v=dQw4w9WgXcQ&feature=youtu.be">video</a>`,
			want: "dQw4w9WgXcQ",
			ok:   true,
		},
		{
			name: "youtu.be short link",
			html: `<a href="https://youtu.be/dQw4w9WgXcQ">video</a>`,
			want: "dQw4w9WgXcQ",
			ok:   true,
		},
		{
			name: "youtube shorts",
			html: `<a href="https://www.youtube.com/shorts/dQw4w9WgXcQ">short</a>`,
			want: "dQw4w9WgXcQ",
			ok:   true,
		},
		{
			name: "no youtube",
			html: `<p>no video here</p>`,
			want: "",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ExtractYouTubeVideoID(tt.html)
			if ok != tt.ok {
				t.Fatalf("ok mismatch: got %v want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("video id mismatch: got %q want %q", got, tt.want)
			}
		})
	}
}
