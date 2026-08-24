package controllers

import "testing"

func TestNormalizeHedocGameImageURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		base string
		want string
	}{
		{"https unchanged", "https://cdn.example.com/game.png", "https://api.example.com/api", "https://cdn.example.com/game.png"},
		{"http upgraded", "http://cdn.example.com/game.png", "https://api.example.com/api", "https://cdn.example.com/game.png"},
		{"protocol relative", "//cdn.example.com/game.png", "https://api.example.com/api", "https://cdn.example.com/game.png"},
		{"relative path", "/images/game.png", "https://api.example.com/api", "https://api.example.com/images/game.png"},
		{"cdn host embedded in api path", "https://api.example.com/cdn.example.com/games/game.png", "https://api.example.com/api", "https://cdn.example.com/games/game.png"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeHedocGameImageURL(test.raw, test.base); got != test.want {
				t.Fatalf("normalizeHedocGameImageURL() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeM7ImageURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		base string
		want string
	}{
		{"https unchanged", "https://cdn.example.com/game.png", "https://api.ddhh.work/", "https://cdn.example.com/game.png"},
		{"http upgraded", "http://cdn.example.com/game.png", "https://api.ddhh.work/", "https://cdn.example.com/game.png"},
		{"protocol relative", "//cdn.example.com/game.png", "https://api.ddhh.work/", "https://cdn.example.com/game.png"},
		{"relative path", "/images/game.png", "https://api.ddhh.work/", "https://api.ddhh.work/images/game.png"},
		{"bare CDN host", "cdn.example.com/games/game.png", "https://api.ddhh.work/", "https://cdn.example.com/games/game.png"},
		{"CDN host embedded in API path", "https://api.ddhh.work/cdn.example.com/games/game.png", "https://api.ddhh.work/", "https://cdn.example.com/games/game.png"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeM7ImageURL(test.base, test.raw); got != test.want {
				t.Fatalf("normalizeM7ImageURL() = %q, want %q", got, test.want)
			}
		})
	}
}
