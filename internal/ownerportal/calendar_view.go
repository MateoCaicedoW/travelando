package ownerportal

import (
	"fmt"
	"time"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

var ownerMonthNames = [...]string{"", "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
var ownerWeekdays = [...]string{"Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"}

type ownerCalendarView struct {
	Year      int
	Month     time.Month
	MonthName string
	Today     time.Time
	Weeks     [][]time.Time
	Entries   map[string][]services.CalendarEntry
	Cabins    []models.Cabin
	Previous  time.Time
	Next      time.Time
}

func newOwnerCalendarView(year int, month time.Month, data services.CalendarData, now time.Time) ownerCalendarView {
	view := ownerCalendarView{
		Year: year, Month: month, MonthName: ownerMonthNames[month], Today: now,
		Weeks: ownerWeeks(year, month), Entries: make(map[string][]services.CalendarEntry),
		Cabins:   data.OccupiedCabins,
		Previous: time.Date(year, month-1, 1, 0, 0, 0, 0, time.UTC),
		Next:     time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC),
	}
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	for _, entry := range data.Entries {
		from, to, ok := entry.OccupiedRange()
		if !ok {
			continue
		}
		firstDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		for day := maxOwnerDate(firstDay, monthStart); day.Before(monthEnd) && day.Before(to); day = day.AddDate(0, 0, 1) {
			key := day.Format(time.DateOnly)
			view.Entries[key] = append(view.Entries[key], entry)
		}
	}
	return view
}

func ownerWeeks(year int, month time.Month) [][]time.Time {
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

func ownerMonthURL(date time.Time) string {
	return fmt.Sprintf("/owner/?month=%d&year=%d", date.Month(), date.Year())
}

func maxOwnerDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
