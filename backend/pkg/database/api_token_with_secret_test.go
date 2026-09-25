package database

import (
	"testing"
	"time"
)

func TestAPITokenWithSecret_IsAPITokenExpired_ExpiresOnceTheTTLHasPassed(t *testing.T) {
	tests := []struct {
		name string
		age  time.Duration
		ttl  int64
		want bool
	}{
		{name: "well inside the ttl", age: time.Minute, ttl: 3600, want: false},
		{name: "a second before the ttl runs out", age: 3599 * time.Second, ttl: 3600, want: false},
		{name: "a second past the ttl", age: 3601 * time.Second, ttl: 3600, want: true},
		{name: "long past the ttl", age: 24 * time.Hour, ttl: 3600, want: true},
		{name: "a zero ttl expires at once", age: time.Nanosecond, ttl: 0, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got := IsAPITokenExpired(start.Add(-tt.age), tt.ttl)
			// The function reads its own clock, so a runner stalled past the token's remaining life cannot judge the row.
			if elapsed := time.Since(start); !tt.want && elapsed >= time.Duration(tt.ttl)*time.Second-tt.age {
				t.Skipf("the call took %v, longer than the token had left", elapsed)
			}
			if got != tt.want {
				t.Errorf("IsAPITokenExpired(%v ago, %d) = %v, want %v", tt.age, tt.ttl, got, tt.want)
			}
		})
	}
}
