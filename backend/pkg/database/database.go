package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	obs "pentagi/pkg/observability"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

func StringToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func Int64ToNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func NullInt64ToInt64(i sql.NullInt64) *int64 {
	if i.Valid {
		return &i.Int64
	}
	return nil
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
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(50)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.SetLogger(&GormLogger{})
	db.LogMode(true)
	return db, nil
}
