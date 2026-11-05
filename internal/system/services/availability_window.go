package services

import (
	"time"

	"travelando/internal/system/models"
)

func reservationWindow(cabin models.Cabin, stayType models.StayType, dateFrom, dateTo string) (time.Time, time.Time, error) {
	startTime, endTime := cabin.DayStart, cabin.DayEnd
	if stayType == models.StayOvernight {
		startTime, endTime = cabin.CheckIn, cabin.CheckOut
	} else if stayType != models.StayDayUse {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	return reservationWindowFor(dateFrom, dateTo, stayType, startTime, endTime, cabin.CleaningHours)
}

func reservationWindowFor(dateFrom, dateTo string, stayType models.StayType, startTime, endTime string, cleaningHours int) (time.Time, time.Time, error) {
	startDate, err := time.Parse(time.DateOnly, dateFrom)
	if err != nil || cleaningHours < 0 || cleaningHours > 24 {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	endDate := startDate
	if stayType == models.StayOvernight {
		endDate, err = time.Parse(time.DateOnly, dateTo)
		if err != nil || !endDate.After(startDate) {
			return time.Time{}, time.Time{}, ErrInvalidInput
		}
	} else if stayType != models.StayDayUse {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	startClock, err := time.Parse("15:04", startTime)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	endClock, err := time.Parse("15:04", endTime)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startClock.Hour(), startClock.Minute(), 0, 0, time.UTC)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), endClock.Hour(), endClock.Minute(), 0, 0, time.UTC)
	if !end.After(start) {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	return start, end.Add(time.Duration(cleaningHours) * time.Hour), nil
}
