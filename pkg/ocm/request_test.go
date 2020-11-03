package ocm

import (
	"testing"
	"time"
)

func Test_ageDate(t *testing.T) {
	tests := []struct {
		name       string
		currentDay time.Time
		daysAgo    int
		want       string
	}{
		{
			name:       "Last day",
			currentDay: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			daysAgo:    1,
			want:       "2020-11-02",
		},
		{
			name:       "Last week",
			currentDay: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			daysAgo:    7,
			want:       "2020-10-27",
		},
		{
			name:       "Today",
			currentDay: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			daysAgo:    0,
			want:       "2020-11-03",
		},
		{
			name:       "Negative days ago?",
			currentDay: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			daysAgo:    -1,
			want:       "2020-11-03",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ageDate(tt.currentDay, tt.daysAgo); got != tt.want {
				t.Errorf("ageDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
