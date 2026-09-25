//go:build postgres

package database_test

import (
	"context"
	"testing"
	"time"
	_ "time/tzdata"

	"pentagi/pkg/database"
)

func TestFlows_GetFlowsStatsByDay_BucketsInTheGivenZoneAndZeroFillsEmptyDays(t *testing.T) {
	db := openSchema(t)

	const zone = "Asia/Riyadh"
	loc, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatalf("load %s: %v", zone, err)
	}

	uid := seedUser(t, db)

	// 00:30 local, two days back: 21:30 UTC on the day before that.
	local := time.Now().In(loc)
	target := time.Date(local.Year(), local.Month(), local.Day(), 0, 30, 0, 0, loc).AddDate(0, 0, -2)
	seedFlow(t, db, uid, target)

	q := database.New(db)
	const days = 6

	localRows, err := q.GetFlowsStatsByDay(context.Background(), database.GetFlowsStatsByDayParams{
		UserID: uid,
		Tz:     zone,
		Days:   days,
	})
	if err != nil {
		t.Fatalf("query in %s: %v", zone, err)
	}

	if len(localRows) != days+1 {
		t.Fatalf("series in %s has %d rows, want %d — a day with no flows must come back as a zero, not be missing",
			zone, len(localRows), days+1)
	}

	wantDay := target.Format("2006-01-02")
	var placed string
	var zeros int
	for _, row := range localRows {
		day := row.Date.In(loc).Format("2006-01-02")
		switch {
		case row.TotalFlowsCount == 1:
			placed = day
		case row.TotalFlowsCount == 0:
			zeros++
		default:
			t.Fatalf("day %s counted %d flows, want 0 or 1", day, row.TotalFlowsCount)
		}
	}

	if placed != wantDay {
		t.Errorf("flow created %s sits in the %s column, want %s", target.Format(time.RFC3339), placed, wantDay)
	}
	if zeros != days {
		t.Errorf("%d empty days came back, want %d", zeros, days)
	}

	utcRows, err := q.GetFlowsStatsByDay(context.Background(), database.GetFlowsStatsByDayParams{
		UserID: uid,
		Tz:     "UTC",
		Days:   days,
	})
	if err != nil {
		t.Fatalf("query in UTC: %v", err)
	}

	var placedUTC string
	for _, row := range utcRows {
		if row.TotalFlowsCount == 1 {
			placedUTC = row.Date.UTC().Format("2006-01-02")
		}
	}

	wantUTC := target.UTC().Format("2006-01-02")
	if placedUTC != wantUTC {
		t.Errorf("read in UTC the flow sits in %s, want %s", placedUTC, wantUTC)
	}
	if placedUTC == placed {
		t.Errorf("both zones put the flow in %s; the timezone argument is not reaching the grouping", placed)
	}
}
