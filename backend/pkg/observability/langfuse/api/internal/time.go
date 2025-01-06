package internal

import (
	"encoding/json"
	"time"
)

const dateFormat = "2006-01-02"

// DateTime wraps time.Time and adapts its JSON representation
// to conform to a RFC3339 date (e.g. 2006-01-02).
//
// Ref: https://ijmacd.github.io/rfc3339-iso8601
type Date struct {
	t *time.Time
}

// NewDate returns a new *Date. If the given time.Time
// is nil, nil will be returned.
func NewDate(t time.Time) *Date {
	return &Date{t: &t}
}

// NewOptionalDate returns a new *Date. If the given time.Time
// is nil, nil will be returned.
func NewOptionalDate(t *time.Time) *Date {
	if t == nil {
		return nil
	}
	return &Date{t: t}
}

// Time returns the Date's underlying time, if any. If the
// date is nil, the zero value is returned.
func (d *Date) Time() time.Time {
	if d == nil || d.t == nil {
		return time.Time{}
	}
	return *d.t
}

// TimePtr returns a pointer to the Date's underlying time.Time, if any.
func (d *Date) TimePtr() *time.Time {
	if d == nil || d.t == nil {
		return nil
	}
	if d.t.IsZero() {
		return nil
	}
	return d.t
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d == nil || d.t == nil {
		return nil, nil
	}
	return json.Marshal(d.t.Format(dateFormat))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsedTime, err := time.Parse(dateFormat, raw)
	if err != nil {
		return err
	}

	*d = Date{t: &parsedTime}
	return nil
}

// DateTime wraps time.Time and adapts its JSON representation
// to conform to a RFC3339 date-time (e.g. 2017-07-21T17:32:28Z).
//
// Ref: https://ijmacd.github.io/rfc3339-iso8601
type DateTime struct {
	t *time.Time
}

// NewDateTime returns a new *DateTime.
func NewDateTime(t time.Time) *DateTime {
	return &DateTime{t: &t}
}

// NewOptionalDateTime returns a new *DateTime. If the given time.Time
// is nil, nil will be returned.
func NewOptionalDateTime(t *time.Time) *DateTime {
	if t == nil {
		return nil
	}
	return &DateTime{t: t}
}

// Time returns the DateTime's underlying time, if any. If the
// date-time is nil, the zero value is returned.
func (d *DateTime) Time() time.Time {
	if d == nil || d.t == nil {
		return time.Time{}
	}
	return *d.t
}

// TimePtr returns a pointer to the DateTime's underlying time.Time, if any.
func (d *DateTime) TimePtr() *time.Time {
	if d == nil || d.t == nil {
		return nil
	}
	if d.t.IsZero() {
		return nil
	}
	return d.t
}

func (d *DateTime) MarshalJSON() ([]byte, error) {
	if d == nil || d.t == nil {
		return nil, nil
	}
	return json.Marshal(d.t.Format(time.RFC3339))
}

func (d *DateTime) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return err
	}

	*d = DateTime{t: &parsedTime}
	return nil
}
