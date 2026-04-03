package appcore

import (
	"testing"
	"time"
)

func TestCalculateStreak(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	tests := []struct {
		name   string
		dates  []time.Time
		expect int
	}{
		{
			name:   "no reviews",
			dates:  nil,
			expect: 0,
		},
		{
			name:   "reviewed today only",
			dates:  []time.Time{today.Add(10 * time.Hour)},
			expect: 1,
		},
		{
			name: "3 consecutive days",
			dates: []time.Time{
				today.Add(8 * time.Hour),
				today.AddDate(0, 0, -1).Add(14 * time.Hour),
				today.AddDate(0, 0, -2).Add(9 * time.Hour),
			},
			expect: 3,
		},
		{
			name: "gap breaks streak",
			dates: []time.Time{
				today.Add(8 * time.Hour),
				today.AddDate(0, 0, -2).Add(14 * time.Hour),
			},
			expect: 1,
		},
		{
			name: "no review today",
			dates: []time.Time{
				today.AddDate(0, 0, -1).Add(14 * time.Hour),
				today.AddDate(0, 0, -2).Add(9 * time.Hour),
			},
			expect: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStreak(tt.dates)
			if got != tt.expect {
				t.Errorf("calculateStreak() = %d, want %d", got, tt.expect)
			}
		})
	}
}
