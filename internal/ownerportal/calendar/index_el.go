package calendar

import (
	"net/http"
	"strconv"
	"time"

	"travelando/internal/system"
	"travelando/internal/system/components"

	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func IndexPageEl(view ownerCalendarView, r *http.Request) Node {
	return system.Layout{Title: "Mis cabañas | Travelando", Portal: system.PortalOwner, Yield: IndexContentEl(view)}.New(r)
}

func IndexContentEl(view ownerCalendarView) Node {
	return Div(ID("owner-content"), Class("mx-auto w-full max-w-7xl px-4 py-4 sm:px-6 sm:py-6"),
		Div(Class("flex flex-col sm:flex-row sm:items-start sm:justify-between sm:gap-x-4"),
			components.PageHeading("Calendario de mis cabañas", "Consulta los datos de tus huéspedes, edita reservas y registra reservas externas."),
			If(len(view.Cabins) > 0, A(Href(BlockModalURL(view.Month, view.Year)), hx.Get(BlockModalURL(view.Month, view.Year)), hx.Target("body"), hx.Swap("beforeend"), Class("inline-flex min-h-10 items-center rounded-lg bg-brand-dark px-4 py-2 text-sm font-extrabold text-white hover:brightness-90"), Text("Registrar reserva externa"))),
		),
		CalendarEl(view),
	)
}

func CalendarEl(view ownerCalendarView) Node {
	return Section(Class("overflow-hidden rounded-2xl border border-line bg-white"),
		Div(Class("flex items-center justify-between gap-3 border-b border-line bg-cream/60 p-3 sm:p-4"),
			A(Href(MonthURL(view.Previous)), hx.Get(MonthURL(view.Previous)), hx.Target("#owner-content"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("grid size-10 place-items-center rounded-lg border border-line bg-white font-bold text-ink hover:border-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Attr("aria-label", "Mes anterior"), Text("←")),
			H2(Class("text-lg font-extrabold capitalize text-ink sm:text-xl"), Text(view.MonthName+" "+strconv.Itoa(view.Year))),
			A(Href(MonthURL(view.Next)), hx.Get(MonthURL(view.Next)), hx.Target("#owner-content"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("grid size-10 place-items-center rounded-lg border border-line bg-white font-bold text-ink hover:border-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Attr("aria-label", "Mes siguiente"), Text("→")),
		),
		If(len(view.Cabins) == 0, P(Class("m-4 rounded-xl bg-cream/60 p-4 text-sm text-muted"), Text("Todavía no tienes cabañas asignadas. Contacta a Travelando."))),
		If(len(view.Cabins) > 0, Div(Class("overflow-x-auto p-2 sm:p-3"),
			Table(Class("w-full table-fixed border-separate border-spacing-1 sm:border-spacing-1.5"),
				Caption(Class("sr-only"), Textf("Calendario de tus cabañas, %s %d", view.MonthName, view.Year)),
				THead(Tr(Map(ownerWeekdays[:], func(day string) Node {
					return Th(Scope("col"), Class("pb-2 text-center text-label font-extrabold uppercase tracking-wide text-muted"), Text(day))
				}))),
				TBody(Map(view.Weeks, func(week []time.Time) Node {
					return Tr(Group(Map(week, func(day time.Time) Node { return DayEl(day, view) })))
				})),
			),
			Div(Class("mt-3 flex flex-wrap gap-x-5 gap-y-2 border-t border-line pt-3 text-xs font-semibold text-muted"),
				EntryLegend("bg-forest", "Reserva de Travelando"),
				EntryLegend("bg-brand-dark", "Reserva externa"),
				EntryLegend("bg-brand", "Otro bloqueo"),
			),
		)),
	)
}
