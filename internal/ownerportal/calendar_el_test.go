package ownerportal

import (
	"strings"
	"testing"
	"time"

	"travelando/internal/system/models"
)

func TestOwnerCalendarRendersEachWeekdayOnceInOrder(t *testing.T) {
	view := ownerCalendarView{
		Year: 2026, Month: time.January, MonthName: "Enero",
		Weeks: ownerWeeks(2026, time.January), Cabins: []models.Cabin{{ID: 1}},
	}
	var html strings.Builder
	if err := ownerCalendarEl(view).Render(&html); err != nil {
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
