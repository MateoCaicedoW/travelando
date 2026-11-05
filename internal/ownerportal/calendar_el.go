package ownerportal

import (
	"net/http"
	"strconv"
	"time"

	lucide "github.com/eduardolat/gomponents-lucide"
	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ownerCalendarPageEl(view ownerCalendarView, r *http.Request) Node {
	return system.Layout{Title: "Mis cabañas | Travelando", Portal: system.PortalOwner, Yield: ownerCalendarContentEl(view)}.New(r)
}

func ownerCalendarContentEl(view ownerCalendarView) Node {
	return Div(ID("owner-content"), Class("mx-auto w-full max-w-7xl px-5 py-8 sm:px-8 sm:py-12"),
		Div(Class("flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"),
			components.PageHeading("Calendario de mis cabañas", "Registra reservas externas para mantener las fechas actualizadas."),
			If(len(view.Cabins) > 0, A(Href(ownerBlockModalURL(view.Month, view.Year)), Attr("hx-get", ownerBlockModalURL(view.Month, view.Year)), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Class("inline-flex min-h-10 items-center rounded-lg bg-brand-dark px-4 py-2 text-sm font-extrabold text-white hover:brightness-90"), Text("Registrar reserva externa"))),
		),
		ownerCalendarEl(view),
	)
}

func ownerCalendarEl(view ownerCalendarView) Node {
	return Section(Class("mt-6 overflow-hidden rounded-2xl border border-line bg-white"),
		Div(Class("flex items-center justify-between gap-3 border-b border-line bg-cream/60 p-4 sm:p-5"),
			A(Href(ownerMonthURL(view.Previous)), Attr("hx-get", ownerMonthURL(view.Previous)), Attr("hx-target", "#owner-content"), Attr("hx-swap", "outerHTML"), Attr("hx-push-url", "true"), Class("grid size-10 place-items-center rounded-lg border border-line bg-white font-bold text-ink hover:border-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Attr("aria-label", "Mes anterior"), Text("←")),
			H2(Class("text-lg font-extrabold capitalize text-ink sm:text-xl"), Text(view.MonthName+" "+strconv.Itoa(view.Year))),
			A(Href(ownerMonthURL(view.Next)), Attr("hx-get", ownerMonthURL(view.Next)), Attr("hx-target", "#owner-content"), Attr("hx-swap", "outerHTML"), Attr("hx-push-url", "true"), Class("grid size-10 place-items-center rounded-lg border border-line bg-white font-bold text-ink hover:border-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Attr("aria-label", "Mes siguiente"), Text("→")),
		),
		If(len(view.Cabins) == 0, P(Class("m-4 rounded-xl bg-cream/60 p-5 text-sm text-muted"), Text("Todavía no tienes cabañas asignadas. Contacta a Travelando."))),
		If(len(view.Cabins) > 0, Div(Class("overflow-x-auto p-2 sm:p-4"),
			Table(Class("w-full table-fixed border-separate border-spacing-1.5"),
				Caption(Class("sr-only"), Textf("Calendario de tus cabañas, %s %d", view.MonthName, view.Year)),
				THead(Tr(Map(make([]int, len(ownerWeekdays)), func(i int) Node {
					return Th(Scope("col"), Class("pb-2 text-center text-xs font-extrabold uppercase tracking-wide text-muted"), Text(ownerWeekdays[i]))
				}))),
				TBody(Map(view.Weeks, func(week []time.Time) Node {
					return Tr(Group(Map(week, func(day time.Time) Node { return ownerDayEl(day, view) })))
				})),
			),
			Div(Class("mt-3 flex flex-wrap gap-x-5 gap-y-2 border-t border-line pt-3 text-xs font-semibold text-muted"),
				ownerLegend("bg-forest", "Reserva de Travelando"),
				ownerLegend("bg-brand-dark", "Reserva externa"),
				ownerLegend("bg-brand", "Otro bloqueo"),
			),
		),
		),
	)
}

func ownerDayEl(day time.Time, view ownerCalendarView) Node {
	if day.Month() != view.Month {
		return Td(Class("h-16 rounded-lg bg-cream/40 sm:h-24"))
	}
	key := day.Format(time.DateOnly)
	entries := view.Entries[key]
	classes := "h-16 rounded-lg border border-line p-1.5 align-top sm:h-24 sm:p-2"
	if len(entries) > 0 {
		classes += " bg-forest-soft/40"
	} else {
		classes += " bg-white"
	}
	if day.Year() == view.Today.Year() && day.Month() == view.Today.Month() && day.Day() == view.Today.Day() {
		classes += " ring-2 ring-brand ring-inset"
	}
	return Td(Class(classes),
		Span(Class("text-sm font-bold tabular-nums text-ink"), Text(strconv.Itoa(day.Day()))),
		Div(Class("mt-1 space-y-0.5 overflow-hidden"), Map(entries, func(entry services.CalendarEntry) Node {
			return ownerCalendarEntryEl(entry, key, view)
		})),
	)
}

func ownerCalendarEntryEl(entry services.CalendarEntry, dayKey string, view ownerCalendarView) Node {
	color, label := "bg-forest", "Reservada"
	if entry.IsBlock {
		color, label = "bg-brand", "No disponible"
		if entry.BlockSource == "external" {
			color, label = "bg-brand-dark", "Externa"
		}
	}
	content := []Node{
		Span(Attr("aria-hidden", "true"), Class("size-2 shrink-0 rounded-full "+color)),
		Span(Attr("title", entry.CabinName+" · "+ownerEntryLabel(entry)+ownerEntrySchedule(entry)), Attr("aria-label", entry.CabinName+" · "+ownerEntryLabel(entry)+ownerEntrySchedule(entry)), Class("min-w-0 flex-1 truncate text-[0.6rem] font-semibold text-ink sm:text-xs"), Text(entry.CabinName+" · "+label+ownerEntrySchedule(entry))),
	}
	if entry.IsBlock && entry.BlockSource == "external" && entry.DateFrom == dayKey {
		deleteAction := "/owner/blocks/" + strconv.FormatInt(entry.BlockID, 10) + "/delete?month=" + strconv.Itoa(int(view.Month)) + "&year=" + strconv.Itoa(view.Year)
		content = append(content, Form(Method("post"), Action(deleteAction), Attr("hx-post", deleteAction), Attr("hx-target", "#owner-content"), Attr("hx-swap", "outerHTML"), Attr("hx-confirm", "¿Quitar esta reserva externa?"), Class("shrink-0"),
			Input(Type("hidden"), Name("cabin_id"), Value(strconv.FormatInt(entry.CabinID, 10))),
			Input(Type("hidden"), Name("date_from"), Value(entry.DateFrom)),
			Button(Type("submit"), Attr("aria-label", "Quitar reserva externa de "+entry.CabinName), Attr("title", "Quitar reserva externa"), Class("grid size-6 shrink-0 place-items-center rounded text-brand-dark hover:bg-cream focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.Trash2(Class("size-3.5"), Attr("aria-hidden", "true"))),
		))
	}
	return Div(Class("flex items-center gap-1 overflow-hidden"), Group(content))
}

func ownerExternalBlockForm(cabins []models.Cabin, state *blockFormState, month time.Month, year int) Node {
	if len(cabins) == 0 {
		return Div()
	}
	fromValue, toValue, cabinValue := "", "", ""
	fromError, toError, cabinError := "", "", ""
	if state != nil {
		fromValue, toValue, cabinValue = state.Values.Get("date_from"), state.Values.Get("date_to"), state.Values.Get("cabin_id")
		fromError, toError, cabinError = state.Errors.ErrorStringFor("date_from"), state.Errors.ErrorStringFor("date_to"), state.Errors.ErrorStringFor("cabin_id")
	}
	cabinOptions := make([]components.SearchableOption, len(cabins))
	for i, cabin := range cabins {
		cabinOptions[i] = components.SearchableOption{Value: strconv.FormatInt(cabin.ID, 10), Label: cabin.Name}
	}
	formAction := "/owner/blocks?month=" + strconv.Itoa(int(month)) + "&year=" + strconv.Itoa(year)
	return Div(Class("mx-auto w-full max-w-4xl p-4 sm:p-7"),
		Div(Class("flex items-start justify-between gap-4"),
			components.PageHeading("Registrar reserva externa", "Incluye reservas hechas por otros medios o para familiares. No se cobra comisión."),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		),
		Form(Method("post"), Action(formAction), Attr("hx-post", formAction), Attr("hx-target", "#owner-content"), Attr("hx-swap", "outerHTML"), Class("mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4"),
			Div(Class("min-w-0"),
				components.SearchableSelect("cabin_id", "cabin_id", "Cabaña", "Busca una cabaña", cabinValue, cabinOptions, true, cabinError),
			),
			components.TextFieldWithError("external-from", "date_from", "Desde", fromValue, "", "date", true, fromError),
			components.TextFieldWithError("external-to", "date_to", "Hasta (sin incluir)", toValue, "", "date", true, toError),
			Div(Class("flex items-end gap-2"),
				Button(Type("submit"), Class("inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-brand-dark px-4 py-2.5 font-extrabold text-white hover:brightness-90"), Text("Bloquear fechas")),
			),
			P(Class("text-xs text-muted sm:col-span-2 lg:col-span-4"), Text("La fecha final no se incluye, igual que el checkout.")),
		),
	)
}

func ownerExternalBlockDialogEl(cabins []models.Cabin, state *blockFormState, month time.Month, year int, oob bool) Node {
	swapOOB := ""
	if oob {
		swapOOB = "outerHTML"
	}
	return components.CustomDialog(components.DialogProps{Label: "Registrar reserva externa", SwapOOB: swapOOB, Yield: ownerExternalBlockForm(cabins, state, month, year)})
}

func ownerBlockModalURL(month time.Month, year int) string {
	return "/owner/blocks/new?month=" + strconv.Itoa(int(month)) + "&year=" + strconv.Itoa(year)
}

func ownerLegend(color, label string) Node {
	return Span(Class("inline-flex items-center gap-2"), Span(Attr("aria-hidden", "true"), Class("size-2.5 rounded-full "+color)), Text(label))
}

func ownerEntryLabel(entry services.CalendarEntry) string {
	if !entry.IsBlock {
		return "Reserva de Travelando"
	}
	if entry.BlockSource == "external" {
		return "Reserva externa"
	}
	return "Bloqueo de Travelando"
}

func ownerEntrySchedule(entry services.CalendarEntry) string {
	if entry.IsBlock {
		return ""
	}
	return " · " + ownerCalendarClock(entry.StartAt) + "–" + ownerCalendarClock(entry.EndAt) + " · lista " + ownerCalendarDateTime(entry.ReadyAt)
}

func ownerCalendarDateTime(value string) string {
	if parsed, err := time.Parse(time.DateTime, value); err == nil {
		return parsed.Format("02/01 15:04")
	}
	return value
}

func ownerCalendarClock(value string) string {
	for _, layout := range []string{time.DateTime, "2006-01-02 15:04"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("15:04")
		}
	}
	return value
}
