package usagelimits

import "time"

// CalculatePeriodBounds calculates period start/end for a reset period.
// For fixed periods (daily, weekly, monthly), returns the current period boundaries.
// For rolling periods, returns (now - period, now).
func CalculatePeriodBounds(period ResetPeriod, now time.Time) (start, end time.Time) {
	// Use UTC for consistency
	now = now.UTC()

	switch period {
	case ResetPeriodDaily:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 1)

	case ResetPeriodWeekly:
		// Week starts on Monday
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday is 7
		}
		daysToMonday := weekday - 1
		start = time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 7)

	case ResetPeriodMonthly:
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)

	case ResetPeriodRolling24h:
		end = now
		start = now.Add(-24 * time.Hour)

	case ResetPeriodRolling7d:
		end = now
		start = now.AddDate(0, 0, -7)

	case ResetPeriodRolling30d:
		end = now
		start = now.AddDate(0, 0, -30)

	default:
		// Default to monthly
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	}

	return start, end
}

// IsRollingPeriod returns true if the period is a rolling window.
func IsRollingPeriod(period ResetPeriod) bool {
	switch period {
	case ResetPeriodRolling24h, ResetPeriodRolling7d, ResetPeriodRolling30d:
		return true
	}
	return false
}
