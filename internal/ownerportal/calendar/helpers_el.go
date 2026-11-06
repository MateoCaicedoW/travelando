package calendar

import (
	"fmt"
	"strconv"
	"time"

	"travelando/internal/system/services"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func MonthURL(date time.Time) string {
	return fmt.Sprintf("/owner/?month=%d&year=%d", date.Month(), date.Year())
}

func BlockModalURL(month time.Month, year int) string {
	return "/owner/blocks/new?month=" + strconv.Itoa(int(month)) + "&year=" + strconv.Itoa(year)
}

func EntryLegend(color, label string) Node {
	return Span(Class("inline-flex items-center gap-2"), Span(Attr("aria-hidden", "true"), Class("size-2.5 rounded-full "+color)), Text(label))
}

func EntrySchedule(entry services.CalendarEntry) string {
	if entry.IsBlock {
		return ""
	}
	return " · " + CalendarClock(entry.StartAt) + "–" + CalendarClock(entry.EndAt) + " · lista " + CalendarDateTime(entry.ReadyAt)
}

func CalendarDateTime(value string) string {
	if parsed, err := time.Parse(time.DateTime, value); err == nil {
		return parsed.Format("02/01 15:04")
	}
	return value
}

func CalendarClock(value string) string {
	for _, layout := range []string{time.DateTime, "2006-01-02 15:04"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("15:04")
		}
	}
	return value
}
