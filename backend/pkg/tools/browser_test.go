package tools

import (
	"testing"
)

func TestBrowserResolveUrl(t *testing.T) {
	tests := []struct {
		name      string
		scPrvURL  string
		scPubURL  string
		targetURL string
		wantURL   string
		wantErr   bool
	}{
		{
			name:      "both URLs set, private target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://192.168.1.1/test",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "both URLs set, public target uses public",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "https://google.com",
			wantURL:   "http://scraper-pub:8080",
			wantErr:   false,
		},
		{
			name:      "only private URL set, private target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "",
			targetURL: "http://localhost:3000",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "only private URL set, public target falls back to private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "",
			targetURL: "https://example.com",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "only public URL set, public target uses public",
			scPrvURL:  "",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "https://google.com",
			wantURL:   "http://scraper-pub:8080",
			wantErr:   false,
		},
		{
			name:      "only public URL set, private target falls back to public",
			scPrvURL:  "",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://10.0.0.1",
			wantURL:   "http://scraper-pub:8080",
			wantErr:   false,
		},
		{
			name:      "no URLs set, returns error",
			scPrvURL:  "",
			scPubURL:  "",
			targetURL: "https://example.com",
			wantURL:   "",
			wantErr:   true,
		},
		{
			name:      "localhost target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://localhost:8080",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "local zone target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://myapp.local",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "10.x.x.x private IP uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://10.1.2.3:8000",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "172.16.x.x private IP uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://172.16.0.1",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &browser{
				scPrvURL: tt.scPrvURL,
				scPubURL: tt.scPubURL,
			}

			gotURL, err := b.resolveUrl(tt.targetURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				got := gotURL.Scheme + "://" + gotURL.Host
				if got != tt.wantURL {
					t.Errorf("resolveUrl() = %v, want %v", got, tt.wantURL)
				}
			}
		})
	}
}

func TestBrowserIsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		scPrvURL string
		scPubURL string
		want     bool
	}{
		{
			name:     "both URLs set",
			scPrvURL: "http://scraper-prv:8080",
			scPubURL: "http://scraper-pub:8080",
			want:     true,
		},
		{
			name:     "only private URL set",
			scPrvURL: "http://scraper-prv:8080",
			scPubURL: "",
			want:     true,
		},
		{
			name:     "only public URL set",
			scPrvURL: "",
			scPubURL: "http://scraper-pub:8080",
			want:     true,
		},
		{
			name:     "no URLs set",
			scPrvURL: "",
			scPubURL: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &browser{
				scPrvURL: tt.scPrvURL,
				scPubURL: tt.scPubURL,
			}

			if got := b.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
