// Package timezone resolves caller-supplied IANA zone names for use in SQL.
package timezone

import (
	"fmt"
	"time"
)

// Resolve turns a caller-supplied name into one Postgres accepts in
// AT TIME ZONE. An absent or empty name means UTC.
func Resolve(name *string) (string, error) {
	if name == nil || *name == "" {
		return "UTC", nil
	}

	loc, err := time.LoadLocation(*name)
	if err != nil {
		return "", fmt.Errorf("unknown timezone: %q", *name)
	}

	if loc == time.Local || loc.String() == "" {
		return "", fmt.Errorf("unknown timezone: %q", *name)
	}

	return loc.String(), nil
}
