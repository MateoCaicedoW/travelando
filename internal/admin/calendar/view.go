package calendar

import (
	"fmt"
	"time"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

var monthNames = [...]string{"", "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}

type cabinSummary struct {
	Cabin        models.Cabin
	OccupiedDays int
}

type calendarView struct {
	Year        int
	Month       time.Month
	MonthName   string
	Today       time.Time
	Weeks       [][]time.Time
	Entries     map[string][]services.CalendarEntry
	Cabins      []cabinSummary
	Previous    time.Time
	Next        time.Time
	YearOptions []int
}

func newCalendarView(year int, month time.Month, data services.CalendarData, now time.Time) calendarView {
	view := calendarView{
		Year: year, Month: month, MonthName: monthNames[month], Today: now,
		Weeks: calendarWeeks(year, month), Entries: make(map[string][]services.CalendarEntry),
		Previous: time.Date(year, month-1, 1, 0, 0, 0, 0, time.UTC),
		Next:     time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC),
	}
	for y := year - 5; y <= year+5; y++ {
		if y >= 1 && y <= 9998 {
			view.YearOptions = append(view.YearOptions, y)
		}
	}
	counts := make(map[int64]map[string]struct{}, len(data.OccupiedCabins))
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	for _, entry := range data.Entries {
		from, to, ok := entry.OccupiedRange()
		if !ok {
			continue
		}
		firstDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		for day := maxDate(firstDay, monthStart); day.Before(monthEnd) && day.Before(to); day = day.AddDate(0, 0, 1) {
			key := day.Format("2006-01-02")
			view.Entries[key] = append(view.Entries[key], entry)
			if counts[entry.CabinID] == nil {
				counts[entry.CabinID] = make(map[string]struct{})
			}
			counts[entry.CabinID][key] = struct{}{}
		}
	}
	for _, cabin := range data.OccupiedCabins {
		if occupiedDays := len(counts[cabin.ID]); occupiedDays > 0 {
			view.Cabins = append(view.Cabins, cabinSummary{Cabin: cabin, OccupiedDays: occupiedDays})
		}
	}
	return view
}

func calendarWeeks(year int, month time.Month) [][]time.Time {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	start := first.AddDate(0, 0, -int(first.Weekday()))
	end := first.AddDate(0, 1, 0)
	weeks := make([][]time.Time, 0, 6)
	for week := start; week.Before(end); week = week.AddDate(0, 0, 7) {
		days := make([]time.Time, 7)
		for i := range days {
			days[i] = week.AddDate(0, 0, i)
		}
		weeks = append(weeks, days)
	}
	return weeks
}

func monthURL(date time.Time) string {
	return fmt.Sprintf("/admin/calendar?month=%d&year=%d", date.Month(), date.Year())
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
