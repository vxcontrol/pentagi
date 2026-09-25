package graph

import (
	"testing"

	"pentagi/pkg/graph/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsPeriod_PeriodDays_MapsEachPeriodToItsDays(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		period  model.UsageStatsPeriod
		want    int32
		wantErr string
	}{
		{name: "a week is seven days", period: model.UsageStatsPeriodWeek, want: 7},
		{name: "a month is thirty days", period: model.UsageStatsPeriodMonth, want: 30},
		{name: "a quarter is ninety days", period: model.UsageStatsPeriodQuarter, want: 90},
		{name: "a period outside the enum", period: model.UsageStatsPeriod("decade"), wantErr: `invalid period: "decade"`},
		{name: "an empty period", period: model.UsageStatsPeriod(""), wantErr: `invalid period: ""`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := periodDays(tt.period)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
