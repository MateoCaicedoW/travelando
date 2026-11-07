package reservations

import (
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func reservationsPageEl(inquiries []models.Inquiry) Node {
	return Div(Class("mx-auto w-full max-w-7xl px-4 py-4 sm:px-6 sm:py-6"),
		Div(Class("flex flex-col sm:flex-row sm:items-start sm:justify-between sm:gap-x-4"),
			components.PageHeading("Reservas", "Gestiona solicitudes pendientes, reservas confirmadas y cancelaciones."),
			Button(Type("button"), hx.Get("/admin/reservations/new"), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-brand-dark px-4 py-2 font-extrabold text-white transition hover:brightness-90"), lucide.Plus(Class("size-4"), Attr("aria-hidden", "true")), Text("Agregar reserva")),
		),
		reservationsContentEl(inquiries),
	)
}

func reservationsContentEl(inquiries []models.Inquiry) Node {
	return Div(ID("reservations-content"), Class("mt-2 md:mt-0 reservation-list-content overflow-hidden rounded-2xl border border-line bg-white"),
		If(len(inquiries) == 0, Div(Attr("role", "status"), Class("px-5 py-8 text-center"),
			P(Class("text-lg font-extrabold text-ink"), Text("Aún no hay reservas")),
			P(Class("mt-1 text-sm text-muted"), Text("Las solicitudes y reservas aparecerán aquí.")),
		)),
		If(len(inquiries) > 0, Div(Class("reservation-actions-scroll overflow-x-auto"),
			Table(Attr("role", "table"), Class("responsive-data-table w-full min-w-[1000px] text-left text-sm"),
				THead(Class("bg-cream/70 text-label uppercase tracking-wider text-muted"), Tr(
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Huésped")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Cabaña / dueño")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Fechas")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Total / estimado")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Estado")),
					Th(Scope("col"), Class("px-4 py-2.5 text-right font-bold"), Text("Acciones")),
				)),
				TBody(Class("divide-y divide-line"), Map(inquiries, reservationRow)),
			),
		)),
	)
}

func reservationRow(inquiry models.Inquiry) Node {
	id := strconv.FormatInt(inquiry.ID, 10)
	detailsURL := "/admin/reservations/" + id
	status, statusClass := "Reservada", "bg-forest text-white"
	if inquiry.Status == models.InquiryPending {
		status, statusClass = "Pendiente", "bg-cream text-brand-dark"
	} else if inquiry.Status == models.InquiryCancelled {
		status, statusClass = "Cancelada", "bg-gray-100 text-muted"
	}
	total, totalLabel := inquiry.FinalTotal, "Total"
	if inquiry.Status == models.InquiryPending || inquiry.Status == models.InquiryCancelled && inquiry.CancelledFromStatus == models.InquiryPending {
		total, totalLabel = inquiry.Estimated, "Estimado"
	}
	actions := []Node{
		A(Href(detailsURL), Attr("role", "menuitem"), hx.Get(detailsURL), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-cream"), Text("Ver")),
	}
	if inquiry.Status == models.InquiryPending || inquiry.Status == models.InquiryReserved {
		editURL := "/admin/inquiries/" + id + "/edit?return_to=reservations"
		if inquiry.Status == models.InquiryPending {
			confirmURL := detailsURL + "/confirm"
			actions = append(actions, A(Href(confirmURL), Attr("role", "menuitem"), hx.Get(confirmURL), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-cream"), Text("Confirmar reserva")))
		}
		actions = append(actions,
			A(Href(editURL), Attr("role", "menuitem"), hx.Get(editURL), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-cream"), Text("Editar")),
			Form(Method("post"), Action("/admin/reservations/"+id+"/cancel"), hx.Post("/admin/reservations/"+id+"/cancel"), hx.Target("#reservations-content"), hx.Swap("outerHTML"), hx.Confirm("¿Cancelar esta solicitud o reserva?"),
				Button(Type("submit"), Attr("role", "menuitem"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-red-800 hover:bg-red-50"), Text("Cancelar")),
			),
		)
	}
	return Tr(ID("reservation-row-"+id),
		Td(Attr("data-label", "Huésped"), Class("px-4 py-3 align-top"), P(Class("font-bold text-ink"), Text(inquiry.GuestName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.GuestPhone))),
		Td(Attr("data-label", "Cabaña / dueño"), Class("px-4 py-3 align-top"), P(Class("font-bold text-ink"), Text(inquiry.CabinName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.OwnerName))),
		Td(Attr("data-label", "Fechas"), Class("px-4 py-3 align-top text-muted"), P(Text(helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo))), P(Class("mt-1 text-xs"), Text(helpers.StayTypeLabel(inquiry.StayType)+" · "+strconv.Itoa(inquiry.Guests)+" huéspedes"))),
		Td(Attr("data-label", "Total / estimado"), Class("px-4 py-3 align-top"), P(Class("text-xs text-muted"), Text(totalLabel)), P(Class("font-bold text-ink"), Text("$"+helpers.FormatCOP(total)+" COP"))),
		Td(Attr("data-label", "Estado"), Class("px-4 py-3 align-top"), Span(Class("inline-flex rounded-full px-2.5 py-1 text-xs font-bold "+statusClass), Text(status))),
		Td(Attr("data-label", "Acciones"), Class("px-4 py-3 text-right align-top"), components.ActionDropdown("admin-reservation-"+id, actions...)),
	)
}
