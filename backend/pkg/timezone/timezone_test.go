package timezone

import (
	"testing"
	_ "time/tzdata"
)

func TestTimezone_Resolve_AcceptsOnlyWhatPostgresAccepts(t *testing.T) {
	name := func(s string) *string { return &s }

	tests := []struct {
		name    string
		in      *string
		want    string
		wantErr bool
	}{
		{name: "absent means UTC", in: nil, want: "UTC"},
		{name: "empty means UTC", in: name(""), want: "UTC"},
		{name: "UTC stays UTC", in: name("UTC"), want: "UTC"},
		{name: "east of UTC", in: name("Asia/Riyadh"), want: "Asia/Riyadh"},
		{name: "west of UTC", in: name("America/New_York"), want: "America/New_York"},
		{name: "half-hour offset", in: name("Asia/Kolkata"), want: "Asia/Kolkata"},
		{name: "Local: Go resolves it, Postgres rejects it", in: name("Local"), wantErr: true},
		{name: "not a zone", in: name("Middle-earth/Shire"), wantErr: true},
		{name: "injection attempt", in: name("UTC'; DROP TABLE flows; --"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Resolve(%v) = %q, want an error", tt.in, got)
				}
				if got != "" {
					t.Errorf("Resolve(%v) returned %q alongside its error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%v) returned %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Resolve(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
