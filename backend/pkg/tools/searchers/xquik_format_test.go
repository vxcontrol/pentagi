package searchers

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestFormatXquikResultsBoundsUntrustedFields(t *testing.T) {
	longText := strings.Repeat("a", xquikMaxTextBytes-1) + "é" + strings.Repeat("b", 20)
	result := formatXquikResults("query", 1, xquikSearchResponse{Tweets: []xquikTweet{{
		ID:   "456",
		Text: longText,
		URL:  "https://x.com.evil.example/injected",
		Author: xquikAuthor{
			Username: "safe_user\n## injected",
			Name:     "Safe name\n## injected",
		},
	}}})

	if strings.Contains(result, "x.com.evil.example") {
		t.Errorf("result trusted an invalid X host: %s", result)
	}
	if strings.Contains(result, "https://x.com/safe_user/status/456") {
		t.Errorf("result built a URL from an invalid username: %s", result)
	}
	if !strings.Contains(result, "[truncated]") {
		t.Errorf("result did not truncate oversized post text: %s", result)
	}
	if !strings.Contains(result, "Safe name ## injected (@safe_user ## injected)") {
		t.Errorf("result did not flatten untrusted author fields: %s", result)
	}
	if !utf8.ValidString(result) {
		t.Errorf("result contains invalid UTF-8 after truncation")
	}
}

func TestFormatXquikResultsClampsUnexpectedRows(t *testing.T) {
	tweets := make([]xquikTweet, xquikMaxLimit+1)
	for index := range tweets {
		tweets[index] = xquikTweet{ID: "1", Text: "post", Author: xquikAuthor{Username: "user"}}
	}

	result := formatXquikResults("query", xquikMaxLimit, xquikSearchResponse{Tweets: tweets})
	if got := strings.Count(result, "<x_post_json>"); got != xquikMaxLimit {
		t.Errorf("rendered posts = %d, want %d", got, xquikMaxLimit)
	}
	if !strings.Contains(result, "did not follow the cursor") {
		t.Errorf("result missing bounded-result notice: %s", result)
	}
}

func TestFormatXquikResultsHonorsRequestedLimit(t *testing.T) {
	tweets := []xquikTweet{
		{ID: "1", Text: "first", Author: xquikAuthor{Username: "user"}},
		{ID: "2", Text: "second", Author: xquikAuthor{Username: "user"}},
	}

	result := formatXquikResults("query", 1, xquikSearchResponse{Tweets: tweets})
	if got := strings.Count(result, "<x_post_json>"); got != 1 {
		t.Errorf("rendered posts = %d, want 1", got)
	}
	if strings.Contains(result, `"second"`) {
		t.Errorf("result exceeded the requested limit: %s", result)
	}
	if !strings.Contains(result, "did not follow the cursor") {
		t.Errorf("result missing bounded-result notice: %s", result)
	}
}

func TestClampXquikLimit(t *testing.T) {
	tests := map[int]int{-1: xquikDefaultLimit, 0: xquikDefaultLimit, 1: 1, 25: 25, 26: xquikMaxLimit}
	for input, want := range tests {
		if got := clampXquikLimit(input); got != want {
			t.Errorf("clampXquikLimit(%d) = %d, want %d", input, got, want)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := map[string]time.Duration{
		"":                    0,
		"0":                   0,
		"nope":                0,
		"3":                   3 * time.Second,
		"3600":                xquikMaxRetryAfter,
		"9223372036854775807": xquikMaxRetryAfter,
	}
	for input, want := range tests {
		if got := parseRetryAfter(input); got != want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", input, got, want)
		}
	}
}
