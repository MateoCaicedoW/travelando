package calendar

import (
	"strconv"
	"time"

	"travelando/internal/system/services"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func DayEl(day time.Time, view ownerCalendarView) Node {
	if day.Month() != view.Month {
		return Td(Class("h-12 rounded-lg bg-cream/40 sm:h-24"))
	}
	key := day.Format(time.DateOnly)
	entries := view.Entries[key]
	classes := "relative h-12 rounded-lg border border-line p-1 align-top sm:h-24 sm:p-2"
	if len(entries) > 0 {
		classes += " bg-forest-soft/40"
	} else {
		classes += " bg-white"
	}
	if day.Year() == view.Today.Year() && day.Month() == view.Today.Month() && day.Day() == view.Today.Day() {
		classes += " ring-2 ring-brand ring-inset"
	}
	dayNumber := Span(Class("text-sm font-bold tabular-nums text-ink"), Text(strconv.Itoa(day.Day())))
	if len(entries) == 0 {
		return Td(Class(classes), dayNumber)
	}
	popoverID := "owner-occupancy-popover-" + key
	buttonLabel := strconv.Itoa(day.Day()) + " de " + view.MonthName + ": " + strconv.Itoa(len(entries)) + " ocupaciones"
	return Td(Class(classes),
		Button(Type("button"), Attr("aria-label", buttonLabel), Attr("aria-controls", popoverID), Attr("aria-haspopup", "dialog"), Attr("popovertarget", popoverID), Class("absolute inset-0 flex flex-col items-start rounded-lg p-1 text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark sm:hidden"),
			dayNumber,
			Div(Class("mt-1 flex flex-nowrap gap-0.5 overflow-hidden"), Map(entries, func(entry services.CalendarEntry) Node {
				color := "bg-forest"
				if entry.IsBlock {
					color = "bg-brand"
					if entry.BlockSource == "external" {
						color = "bg-brand-dark"
					}
				}
				return Span(Attr("aria-hidden", "true"), Class("size-2 shrink-0 rounded-full "+color))
			})),
		),
		Div(Class("mt-1 hidden space-y-0.5 overflow-hidden sm:block"), Map(entries, func(entry services.CalendarEntry) Node {
			return CalendarEntryEl(entry, key, view)
		})),
		Div(ID(popoverID), Attr("popover", "auto"), Attr("role", "dialog"), Attr("aria-label", buttonLabel), Class("fixed inset-x-4 bottom-4 top-auto m-auto max-h-[70vh] w-80 max-w-[calc(100vw-2rem)] overflow-y-auto rounded-2xl border border-line bg-white p-4 text-left shadow-xl"), Map(entries, func(entry services.CalendarEntry) Node {
			return CalendarEntryEl(entry, key, view)
		})),
	)
}

func CalendarEntryEl(entry services.CalendarEntry, dayKey string, view ownerCalendarView) Node {
	color, label := "bg-forest", "Reservada"
	if entry.IsBlock {
		color, label = "bg-brand", "No disponible"
		if entry.BlockSource == "external" {
			color, label = "bg-brand-dark", "Externa"
		}
	}
	entryText := entry.CabinName + " · " + label + EntrySchedule(entry)
	if !entry.IsBlock {
		entryText += " · " + entry.GuestName
	}
	content := []Node{
		Span(Attr("aria-hidden", "true"), Class("size-2 shrink-0 rounded-full "+color)),
		Span(Attr("title", entryText+GuestContact(entry)), Attr("aria-label", entryText+GuestContact(entry)), Class("min-w-0 flex-1 truncate text-[0.6rem] font-semibold text-ink sm:text-xs"), Text(entryText)),
	}
	if !entry.IsBlock {
		editURL := "/owner/reservations/" + strconv.FormatInt(entry.InquiryID, 10) + "/edit"
		content[1] = A(Href(editURL), hx.Get(editURL), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Attr("title", entryText+GuestContact(entry)), Attr("aria-label", "Editar reserva: "+entryText+GuestContact(entry)), Class("min-w-0 flex-1 truncate text-[0.6rem] font-semibold text-ink hover:text-brand-dark hover:underline sm:text-xs"), Text(entryText))
	}
	if entry.IsBlock && entry.BlockSource == "external" && entry.DateFrom == dayKey {
		deleteAction := "/owner/blocks/" + strconv.FormatInt(entry.BlockID, 10) + "/delete?month=" + strconv.Itoa(int(view.Month)) + "&year=" + strconv.Itoa(view.Year)
		content = append(content, Form(Method("post"), Action(deleteAction), hx.Post(deleteAction), hx.Target("#owner-content"), hx.Swap("outerHTML"), hx.Confirm("¿Quitar esta reserva externa?"), Class("shrink-0"),
			Input(Type("hidden"), Name("cabin_id"), Value(strconv.FormatInt(entry.CabinID, 10))),
			Input(Type("hidden"), Name("date_from"), Value(entry.DateFrom)),
			Button(Type("submit"), Attr("aria-label", "Quitar reserva externa de "+entry.CabinName), Attr("title", "Quitar reserva externa"), Class("grid size-6 shrink-0 place-items-center rounded text-brand-dark hover:bg-cream focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.Trash2(Class("size-3.5"), Attr("aria-hidden", "true"))),
		))
	}
	return Div(Class("flex items-center gap-1 overflow-hidden"), Group(content))
}

func GuestContact(entry services.CalendarEntry) string {
	if entry.IsBlock {
		return ""
	}
	return " · " + entry.GuestPhone + " · " + entry.GuestEmail
}
