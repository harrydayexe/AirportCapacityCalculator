package simulation

import "time"

const (
	// Time constants for calculations

	// HoursPerDay represents the number of hours in a day
	HoursPerDay = 24

	// DaysPerYear represents the number of days in a standard year
	DaysPerYear = 365

	// HoursPerYear represents the total operating hours in a full year (365 days * 24 hours)
	// This is used as the default for 24/7 airport operations
	HoursPerYear = DaysPerYear * HoursPerDay // 8760 hours

	// SecondsPerHour represents the number of seconds in an hour
	SecondsPerHour = 3600
)

// YearDuration represents the duration of a standard year
const YearDuration = DaysPerYear * HoursPerDay * time.Hour
