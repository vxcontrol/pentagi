package api

import (
	"time"

	"github.com/google/uuid"
)

// Bool returns a pointer to the given bool value.
func Bool(b bool) *bool {
	return &b
}

// Byte returns a pointer to the given byte value.
func Byte(b byte) *byte {
	return &b
}

// Complex64 returns a pointer to the given complex64 value.
func Complex64(c complex64) *complex64 {
	return &c
}

// Complex128 returns a pointer to the given complex128 value.
func Complex128(c complex128) *complex128 {
	return &c
}

// Float32 returns a pointer to the given float32 value.
func Float32(f float32) *float32 {
	return &f
}

// Float64 returns a pointer to the given float64 value.
func Float64(f float64) *float64 {
	return &f
}

// Int returns a pointer to the given int value.
func Int(i int) *int {
	return &i
}

// Int8 returns a pointer to the given int8 value.
func Int8(i int8) *int8 {
	return &i
}

// Int16 returns a pointer to the given int16 value.
func Int16(i int16) *int16 {
	return &i
}

// Int32 returns a pointer to the given int32 value.
func Int32(i int32) *int32 {
	return &i
}

// Int64 returns a pointer to the given int64 value.
func Int64(i int64) *int64 {
	return &i
}

// Rune returns a pointer to the given rune value.
func Rune(r rune) *rune {
	return &r
}

// String returns a pointer to the given string value.
func String(s string) *string {
	return &s
}

// Uint returns a pointer to the given uint value.
func Uint(u uint) *uint {
	return &u
}

// Uint8 returns a pointer to the given uint8 value.
func Uint8(u uint8) *uint8 {
	return &u
}

// Uint16 returns a pointer to the given uint16 value.
func Uint16(u uint16) *uint16 {
	return &u
}

// Uint32 returns a pointer to the given uint32 value.
func Uint32(u uint32) *uint32 {
	return &u
}

// Uint64 returns a pointer to the given uint64 value.
func Uint64(u uint64) *uint64 {
	return &u
}

// Uintptr returns a pointer to the given uintptr value.
func Uintptr(u uintptr) *uintptr {
	return &u
}

// UUID returns a pointer to the given uuid.UUID value.
func UUID(u uuid.UUID) *uuid.UUID {
	return &u
}

// Time returns a pointer to the given time.Time value.
func Time(t time.Time) *time.Time {
	return &t
}

// MustParseDate attempts to parse the given string as a
// date time.Time, and panics upon failure.
func MustParseDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return t
}

// MustParseDateTime attempts to parse the given string as a
// datetime time.Time, and panics upon failure.
func MustParseDateTime(datetime string) time.Time {
	t, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		panic(err)
	}
	return t
}
