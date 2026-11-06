package calendar

import (
	"fmt"
	"strconv"
	"time"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"
	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type weekday struct{ short, name string }

var weekdays = [...]weekday{{"Dom", "domingo"}, {"Lun", "lunes"}, {"Mar", "martes"}, {"Mié", "miércoles"}, {"Jue", "jueves"}, {"Vie", "viernes"}, {"Sáb", "sábado"}}

func calendarPageEl(view calendarView) Node {
	return Div(Class("mx-auto w-full max-w-7xl px-4 py-4 sm:px-6 sm:py-6"),
		components.PageHeading("Calendario", "Consulta la ocupación mensual de todas las cabañas."),
		calendarViewEl(view),
	)
}

func calendarViewEl(view calendarView) Node {
	return Div(ID("calendar-view"),
		Div(Class("grid gap-4 lg:grid-cols-[minmax(15rem,0.8fr)_minmax(0,2fr)] lg:items-start"),
			cabinSummaryEl(view.Cabins, view.MonthName),
			Div(Class("overflow-visible rounded-2xl border border-line bg-white"),
				Div(Class("flex flex-col gap-4 border-b border-line bg-cream/60 p-3 sm:flex-row sm:items-center sm:justify-between sm:p-4"),
					H2(Class("text-xl font-extrabold capitalize text-ink sm:text-2xl"), Text(view.MonthName+" "), Span(Class("font-semibold text-muted"), Text(strconv.Itoa(view.Year)))),
					Div(Class("flex flex-wrap items-center gap-2"),
						monthControlsEl(view),
						Nav(Attr("aria-label", "Cambiar mes"), Class("flex gap-2"),
							monthLink(view.Previous, lucide.ChevronLeft(Class("size-4"), Attr("aria-hidden", "true")), "Mes anterior"),
							monthLink(view.Next, lucide.ChevronRight(Class("size-4"), Attr("aria-hidden", "true")), "Mes siguiente"),
						),
					),
				),
				Div(Class("overflow-visible p-2 sm:p-3"),
					Table(Class("w-full table-fixed border-separate border-spacing-1.5"),
						Caption(Class("sr-only"), Textf("Calendario de ocupación de %s %d", view.MonthName, view.Year)),
						THead(Tr(Map(weekdays[:], func(day weekday) Node {
							return Th(Scope("col"), Attr("aria-label", day.name), Class("pb-2 text-center text-label font-extrabold uppercase tracking-wide text-muted"), Text(day.short))
						}))),
						TBody(Map(view.Weeks, func(week []time.Time) Node {
							return Tr(Group(Map(week, func(day time.Time) Node { return calendarDayEl(day, view) })))
						})),
					),
					Div(Class("mt-3 flex flex-wrap gap-x-5 gap-y-2 border-t border-line pt-3 text-xs font-semibold text-muted"),
						legendItem("border border-line bg-white", "Disponible"),
						legendItem("bg-forest", "Reserva confirmada"),
						legendItem("bg-brand", "Bloqueo manual"),
						legendItem("bg-brand-dark", "Reserva externa"),
					),
				),
			),
		),
	)
}

func cabinSummaryEl(cabins []cabinSummary, month string) Node {
	return Section(Attr("aria-labelledby", "occupancy-heading"), Class("rounded-2xl border border-line bg-white p-3 sm:p-4"),
		Div(Class("flex items-start justify-between gap-3 border-b border-line pb-3"),
			Div(H2(ID("occupancy-heading"), Class("text-lg font-extrabold text-ink"), Text("Cabañas ocupadas")), P(Class("mt-1 text-sm capitalize text-muted"), Text(month))),
			Span(Class("rounded-full bg-cream px-2.5 py-1 text-xs font-bold text-brand-dark"), Textf("%d", len(cabins))),
		),
		If(len(cabins) == 0, P(Class("py-8 text-center text-sm text-muted"), Text("Ninguna cabaña ocupada este mes."))),
		Ul(Class("divide-y divide-line"), Map(cabins, func(cabin cabinSummary) Node {
			status := fmt.Sprintf("%d días ocupados", cabin.OccupiedDays)
			statusClass := "bg-forest-soft text-forest"
			if cabin.OccupiedDays == 1 {
				status = "1 día ocupado"
			}
			return Li(Class("flex items-center justify-between gap-3 py-3 first:pt-4 last:pb-1"),
				Div(Class("min-w-0"), P(Class("truncate text-sm font-bold text-ink"), Text(cabin.Cabin.Name)), If(!cabin.Cabin.Active, P(Class("mt-0.5 text-xs text-muted"), Text("Oculta del catálogo")))),
				Span(Class("shrink-0 rounded-full px-2.5 py-1 text-xs font-bold "+statusClass), Text(status)),
			)
		})),
		P(Class("mt-3 border-t border-line pt-3 text-xs leading-5 text-muted"), Text("Las reservas pendientes no bloquean fechas.")),
	)
}

func monthControlsEl(view calendarView) Node {
	return Form(Method("get"), Action("/admin/calendar"), hx.Get("/admin/calendar"), hx.Target("#calendar-view"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("flex items-center gap-2"),
		Label(Class("sr-only"), For("calendar-month"), Text("Mes")),
		Select(ID("calendar-month"), Name("month"), Attr("aria-label", "Mes"), hx.Get("/admin/calendar"), hx.Trigger("change"), hx.Include("#calendar-year"), hx.Target("#calendar-view"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("min-h-10 w-32 rounded-lg border border-line bg-white px-2 text-sm font-bold text-ink"), Map(monthNames[1:], func(name string) Node {
			month := monthNumber(name)
			return Option(Value(strconv.Itoa(int(month))), If(month == view.Month, Selected()), Text(monthNames[month]))
		})),
		Label(Class("sr-only"), For("calendar-year"), Text("Año")),
		Select(ID("calendar-year"), Name("year"), Attr("aria-label", "Año"), hx.Get("/admin/calendar"), hx.Trigger("change"), hx.Include("#calendar-month"), hx.Target("#calendar-view"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("min-h-10 w-24 rounded-lg border border-line bg-white px-2 text-sm font-bold text-ink"), Map(view.YearOptions, func(year int) Node {
			return Option(Value(strconv.Itoa(year)), If(year == view.Year, Selected()), Text(strconv.Itoa(year)))
		})),
	)
}

func monthNumber(name string) time.Month {
	for month, monthName := range monthNames {
		if monthName == name {
			return time.Month(month)
		}
	}
	return time.January
}

func monthLink(date time.Time, icon Node, ariaLabel string) Node {
	url := monthURL(date)
	return A(Href(url), Attr("aria-label", ariaLabel), hx.Get(url), hx.Target("#calendar-view"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("grid size-10 place-items-center rounded-lg border border-line bg-white text-lg font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), icon)
}

func calendarDayEl(day time.Time, view calendarView) Node {
	if day.Month() != view.Month {
		return Td(Class("h-[4.25rem] rounded-lg bg-cream/40 sm:h-24"))
	}
	key := day.Format("2006-01-02")
	entries := view.Entries[key]
	classes := "relative h-[4.25rem] rounded-lg border border-line sm:h-24"
	if len(entries) > 0 {
		classes += " bg-forest-soft/40"
	} else {
		classes += " bg-white p-1.5 sm:p-2"
	}
	if day.Year() == view.Today.Year() && day.Month() == view.Today.Month() && day.Day() == view.Today.Day() {
		classes += " ring-2 ring-brand ring-inset"
	}
	content := []Node{Span(Class("text-sm font-bold tabular-nums text-ink sm:text-base"), Text(strconv.Itoa(day.Day())))}
	if len(entries) > 0 {
		content = append(content, Div(Class("mt-2 flex flex-wrap gap-1"), Map(entries, func(entry services.CalendarEntry) Node {
			color := "bg-forest"
			if entry.IsBlock {
				color = "bg-brand"
				if entry.BlockSource == "external" {
					color = "bg-brand-dark"
				}
			}
			return Span(Attr("aria-hidden", "true"), Class("size-2.5 rounded-full "+color))
		})))
	}
	if len(entries) > 0 {
		buttonLabel := fmt.Sprintf("%d de %s: %d ocupaciones", day.Day(), view.MonthName, len(entries))
		popoverID := "occupancy-popover-" + key
		return Td(Class(classes),
			Button(Type("button"), Attr("aria-label", buttonLabel), Attr("aria-controls", popoverID), Attr("aria-haspopup", "dialog"), Attr("popovertarget", popoverID), Class("absolute inset-0 flex flex-col items-start rounded-lg p-1.5 text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark sm:p-2"), Group(content)),
			Div(ID(popoverID), Attr("popover", "auto"), Attr("role", "dialog"), Attr("aria-label", buttonLabel), Class("fixed inset-x-4 bottom-4 top-auto m-auto max-h-[70vh] w-80 max-w-[calc(100vw-2rem)] overflow-y-auto rounded-2xl border border-line bg-white p-4 text-left shadow-xl"), Map(entries, calendarEntryEl)),
		)
	}
	return Td(Class(classes), Group(content))
}

func calendarEntryEl(entry services.CalendarEntry) Node {
	dateLabel := "Desde " + displayDate(entry.DateFrom) + " · hasta " + displayDate(entry.DateTo) + " (sin incluir)"
	if entry.IsBlock {
		if entry.BlockSource == "external" {
			return Div(Class("border-b border-line py-2 last:border-0 last:pb-0"),
				P(Class("text-sm font-extrabold text-ink"), Text(entry.CabinName)),
				P(Class("text-xs font-bold text-brand-dark"), Text("Reserva externa")),
				P(Class("mt-1 text-xs text-muted"), Text(dateLabel)),
			)
		}
		reason := entry.BlockReason
		if reason == "" {
			reason = "Bloqueo manual"
		}
		return Div(Class("border-b border-line py-2 last:border-0 last:pb-0"),
			P(Class("text-sm font-extrabold text-ink"), Text(entry.CabinName)),
			P(Class("text-xs font-bold text-brand-dark"), Text("Bloqueo manual")),
			P(Class("mt-1 text-xs text-muted"), Text(reason)),
			P(Class("mt-1 text-xs text-muted"), Text(dateLabel)),
		)
	}
	stay := "Pasadía"
	if entry.StayType == models.StayOvernight {
		stay = "Estadía"
	}
	return Div(Class("border-b border-line py-2 last:border-0 last:pb-0"),
		P(Class("text-sm font-extrabold text-ink"), Text(entry.CabinName)),
		P(Class("text-xs font-bold text-forest"), Text("Reserva confirmada · "+stay)),
		P(Class("mt-1 text-xs text-muted"), Textf("%s · %d personas", entry.GuestName, entry.Guests)),
		P(Class("mt-1 text-xs text-muted"), Text(dateLabel)),
		P(Class("mt-1 text-xs text-muted"), Textf("%s–%s · lista desde %s", displayClock(entry.StartAt), displayClock(entry.EndAt), displayDateTime(entry.ReadyAt))),
	)
}

func displayDate(value string) string {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return date.Format("02/01/2006")
}

func displayClock(value string) string {
	for _, layout := range []string{time.DateTime, "2006-01-02 15:04"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("15:04")
		}
	}
	return value
}

func displayDateTime(value string) string {
	for _, layout := range []string{time.DateTime, "2006-01-02 15:04"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("02/01 15:04")
		}
	}
	return value
}

func legendItem(color, label string) Node {
	return Span(Class("inline-flex items-center gap-2"), Span(Attr("aria-hidden", "true"), Class("size-2.5 rounded-full "+color)), Text(label))
}
