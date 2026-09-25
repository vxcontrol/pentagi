package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenumberedMigrations_RenumberPlan_PlansAMoveOnlyForTheEarlierNumbering(t *testing.T) {
	for _, tc := range []struct {
		name     string
		versions []int64
		want     map[int64]int64
	}{
		{
			name:     "every earlier version is moved",
			versions: []int64{0, 20260621, 20260716, 20260823, 20260910, 20260911, 20260915},
			want:     map[int64]int64{20260823: 20260717, 20260910: 20260718, 20260911: 20260719, 20260915: 20260720},
		},
		{
			name:     "a database that stopped partway is moved as far as it got",
			versions: []int64{20260716, 20260823},
			want:     map[int64]int64{20260823: 20260717},
		},
		{
			name:     "the current numbering is left alone",
			versions: []int64{20260716, 20260717, 20260718, 20260719, 20260720, 20260721},
		},
		{
			name:     "a database of the line that reuses the numbers is left alone",
			versions: []int64{20260716, 20260801, 20260823, 20260910},
		},
		{
			name:     "a fresh database is left alone",
			versions: []int64{0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, renumberPlan(tc.versions))
		})
	}
}
