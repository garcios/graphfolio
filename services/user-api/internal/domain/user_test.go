package domain_test

import (
	"testing"
	"user-api/internal/domain"
)

func TestIsValidFinancialYearEnd(t *testing.T) {
	tests := []struct {
		name  string
		month int
		day   int
		want  bool
	}{
		// Valid major EOFY presets
		{name: "Australia June 30", month: 6, day: 30, want: true},
		{name: "UK April 5", month: 4, day: 5, want: true},
		{name: "US / Canada Dec 31", month: 12, day: 31, want: true},
		{name: "NZ / India March 31", month: 3, day: 31, want: true},
		{name: "South Africa Feb 28", month: 2, day: 28, want: true},
		{name: "Leap year Feb 29", month: 2, day: 29, want: true},
		{name: "Jan 1 (First day of year)", month: 1, day: 1, want: true},
		{name: "Jan 31", month: 1, day: 31, want: true},
		{name: "Jul 31", month: 7, day: 31, want: true},
		{name: "Aug 31", month: 8, day: 31, want: true},
		{name: "Sep 30", month: 9, day: 30, want: true},
		{name: "Oct 31", month: 10, day: 31, want: true},
		{name: "Nov 30", month: 11, day: 30, want: true},

		// Invalid bounds
		{name: "Month 0", month: 0, day: 15, want: false},
		{name: "Month 13", month: 13, day: 15, want: false},
		{name: "Day 0", month: 6, day: 0, want: false},
		{name: "Day 32", month: 12, day: 32, want: false},
		{name: "Negative month", month: -1, day: 15, want: false},
		{name: "Negative day", month: 6, day: -5, want: false},

		// Days exceeding specific month limits
		{name: "Feb 30", month: 2, day: 30, want: false},
		{name: "Feb 31", month: 2, day: 31, want: false},
		{name: "Apr 31 (30 days hath April)", month: 4, day: 31, want: false},
		{name: "Jun 31", month: 6, day: 31, want: false},
		{name: "Sep 31", month: 9, day: 31, want: false},
		{name: "Nov 31", month: 11, day: 31, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.IsValidFinancialYearEnd(tt.month, tt.day)
			if got != tt.want {
				t.Errorf("IsValidFinancialYearEnd(%d, %d) = %v; want %v", tt.month, tt.day, got, tt.want)
			}
		})
	}
}
