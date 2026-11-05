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
