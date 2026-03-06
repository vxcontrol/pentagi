package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	obs "pentagi/pkg/observability"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

func StringToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func PtrStringToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func NullStringToPtrString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func Int64ToNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func Uint64ToNullInt64(i *uint64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Int64: 0, Valid: false}
	}
	return sql.NullInt64{Int64: int64(*i), Valid: true}
}

func NullInt64ToInt64(i sql.NullInt64) *int64 {
	if i.Valid {
		return &i.Int64
	}
	return nil
}

func TimeToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

func PtrTimeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func SanitizeUTF8(msg string) string {
	if msg == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(msg)) // Pre-allocate for efficiency

	for i := 0; i < len(msg); {
		// Explicitly skip null bytes
		if msg[i] == '\x00' {
			i++
			continue
		}
		// Decode rune and check for errors
		r, size := utf8.DecodeRuneInString(msg[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 byte, replace with Unicode replacement character
			builder.WriteRune(utf8.RuneError)
			i += size
		} else {
			builder.WriteRune(r)
			i += size
		}
	}

	return builder.String()
}

type GormLogger struct{}

func (*GormLogger) Print(v ...interface{}) {
	ctx, span := obs.Observer.NewSpan(context.TODO(), obs.SpanKindInternal, "gorm.print")
	defer span.End()

	switch v[0] {
	case "sql":
		query := fmt.Sprintf("%v", v[3])
		values := v[4].([]interface{})
		for i, val := range values {
			query = strings.Replace(query, fmt.Sprintf("$%d", i+1), fmt.Sprintf("'%v'", val), 1)
		}
		logrus.WithContext(ctx).WithFields(
			logrus.Fields{
				"component":     "pentagi-gorm",
				"type":          "sql",
				"rows_returned": v[5],
				"src":           v[1],
				"values":        v[4],
				"duration":      v[2],
			},
		).Info(query)
	case "log":
		logrus.WithContext(ctx).WithFields(logrus.Fields{"component": "pentagi-gorm"}).Info(v[2])
	case "info":
		// do not log validators
	}
}

func NewGorm(dsn, dbType string) (*gorm.DB, error) {
	db, err := gorm.Open(dbType, dsn)
	if err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(5)
	db.DB().SetMaxOpenConns(20)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.SetLogger(&GormLogger{})
	db.LogMode(true)
	return db, nil
}
