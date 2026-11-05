package calendar

import (
	"strings"
	"testing"
	"time"

	"travelando/internal/system/services"
)

func TestCalendarViewRendersEachWeekdayOnceInOrder(t *testing.T) {
	view := calendarView{
		Year: 2026, Month: time.January, MonthName: "Enero",
		Weeks: calendarWeeks(2026, time.January), Entries: map[string][]services.CalendarEntry{},
	}
	var html strings.Builder
	if err := calendarViewEl(view).Render(&html); err != nil {
		t.Fatal(err)
	}

	previous := -1
	for _, day := range []string{">Dom</th>", ">Lun</th>", ">Mar</th>", ">Mié</th>", ">Jue</th>", ">Vie</th>", ">Sáb</th>"} {
		position := strings.Index(html.String(), day)
		if position <= previous || strings.Count(html.String(), day) != 1 {
			t.Errorf("weekday %s rendered out of order or not exactly once", day)
		}
		previous = position
	}
}

func TestOccupiedDayUsesOnlyDetailPopover(t *testing.T) {
	day := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	view := calendarView{
		Year: 2026, Month: time.January, MonthName: "Enero", Today: day,
		Entries: map[string][]services.CalendarEntry{"2026-01-01": {{IsBlock: true}}},
	}
	var html strings.Builder
	if err := calendarDayEl(day, view).Render(&html); err != nil {
		t.Fatal(err)
	}
	output := html.String()
	if strings.Contains(output, `role="tooltip"`) || strings.Contains(output, "calendar-tooltip") || strings.Contains(output, "aria-describedby=") {
		t.Fatal("occupied day unexpectedly renders duplicate tooltip")
	}
	if !strings.Contains(output, `popovertarget="occupancy-popover-2026-01-01"`) {
		t.Fatal("occupied day is missing its detail popover")
	}
	if !strings.Contains(output, "absolute inset-0") {
		t.Fatal("occupied day trigger does not cover the full cell")
	}
}
