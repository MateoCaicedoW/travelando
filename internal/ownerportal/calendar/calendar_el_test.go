package calendar

import (
	"strings"
	"testing"
	"time"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
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
	if !strings.Contains(html.String(), `border-spacing-1 sm:border-spacing-1.5`) {
		t.Error("owner calendar should use tighter cell spacing on mobile")
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

func TestOwnerCalendarShowsCompactMobileDaysWithEntryDetails(t *testing.T) {
	day := time.Date(2026, time.November, 5, 0, 0, 0, 0, time.UTC)
	view := ownerCalendarView{
		Year: 2026, Month: time.November, MonthName: "Noviembre", Today: day,
		Entries: map[string][]services.CalendarEntry{
			day.Format(time.DateOnly): {{CabinID: 1, CabinName: "Cabaña norte", IsBlock: true, BlockSource: "external", BlockID: 2, DateFrom: "2026-11-05", DateTo: "2026-11-07"}},
		},
	}
	var html strings.Builder
	if err := ownerDayEl(day, view).Render(&html); err != nil {
		t.Fatal(err)
	}

	output := html.String()
	for _, want := range []string{`h-12`, `sm:h-24`, `class="absolute inset-0`, `sm:hidden`, `gap-0.5`, `size-2 shrink-0 rounded-full`, `aria-haspopup="dialog"`, `popover="auto"`, `sm:block`, "Cabaña norte", "Externa", "Quitar reserva externa"} {
		if !strings.Contains(output, want) {
			t.Errorf("owner calendar day missing %q: %s", want, output)
		}
	}
}
