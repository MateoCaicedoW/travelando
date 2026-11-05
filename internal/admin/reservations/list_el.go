package reservations

import (
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func reservationsPageEl(inquiries []models.Inquiry) Node {
	return Div(Class("mx-auto w-full max-w-7xl px-5 py-8 sm:px-8 sm:py-12"),
		components.PageHeading("Reservas", "Gestiona solicitudes pendientes, reservas confirmadas y cancelaciones."),
		reservationsContentEl(inquiries),
	)
}

func reservationsContentEl(inquiries []models.Inquiry) Node {
	return Div(ID("reservations-content"), Class("overflow-hidden rounded-2xl border border-line bg-white"),
		Div(Class("flex flex-wrap items-center justify-between gap-3 border-b border-line p-5 sm:px-6"),
			H2(Class("text-xl font-extrabold text-ink"), Text("Todas las reservas")),
			P(Class("text-sm text-muted"), Textf("%d registros", len(inquiries))),
		),
		If(len(inquiries) == 0, Div(Attr("role", "status"), Class("px-6 py-12 text-center"),
			P(Class("text-lg font-extrabold text-ink"), Text("Aún no hay reservas")),
			P(Class("mt-1 text-sm text-muted"), Text("Las solicitudes y reservas aparecerán aquí.")),
		)),
		If(len(inquiries) > 0, Div(Class("overflow-x-auto"),
			Table(Class("w-full min-w-[1000px] text-left text-sm"),
				THead(Class("bg-cream/70 text-xs uppercase tracking-wider text-muted"), Tr(
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Huésped")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Cabaña / dueño")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Fechas")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Total / estimado")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Estado")),
					Th(Scope("col"), Class("px-5 py-3 text-right font-bold"), Text("Acciones")),
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
		A(Href(detailsURL), Attr("hx-get", detailsURL), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("inline-flex min-h-9 items-center rounded-lg border border-line px-3 text-xs font-bold text-brand-dark hover:border-brand hover:bg-cream"), Text("Ver")),
	}
	if inquiry.Status == models.InquiryPending || inquiry.Status == models.InquiryReserved {
		editURL := "/admin/inquiries/" + id + "/edit?return_to=reservations"
		actions = append(actions,
			A(Href(editURL), Attr("hx-get", editURL), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("inline-flex min-h-9 items-center rounded-lg border border-line px-3 text-xs font-bold text-ink hover:border-brand hover:bg-cream"), Text("Editar")),
			Form(Method("post"), Action("/admin/reservations/"+id+"/cancel"), Attr("hx-post", "/admin/reservations/"+id+"/cancel"), Attr("hx-target", "#reservations-content"), Attr("hx-swap", "outerHTML"), Attr("hx-confirm", "¿Cancelar esta solicitud o reserva?"),
				Button(Type("submit"), Class("inline-flex min-h-9 items-center rounded-lg border border-red-200 px-3 text-xs font-bold text-red-800 hover:bg-red-50"), Text("Cancelar")),
			),
		)
	}
	return Tr(ID("reservation-row-"+id),
		Td(Class("px-5 py-4 align-top"), P(Class("font-bold text-ink"), Text(inquiry.GuestName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.GuestPhone))),
		Td(Class("px-5 py-4 align-top"), P(Class("font-bold text-ink"), Text(inquiry.CabinName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.OwnerName))),
		Td(Class("px-5 py-4 align-top text-muted"), P(Text(helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo))), P(Class("mt-1 text-xs"), Text(helpers.StayTypeLabel(inquiry.StayType)+" · "+strconv.Itoa(inquiry.Guests)+" huéspedes"))),
		Td(Class("px-5 py-4 align-top"), P(Class("text-xs text-muted"), Text(totalLabel)), P(Class("font-bold text-ink"), Text("$"+helpers.FormatCOP(total)+" COP"))),
		Td(Class("px-5 py-4 align-top"), Span(Class("inline-flex rounded-full px-2.5 py-1 text-xs font-bold "+statusClass), Text(status))),
		Td(Class("px-5 py-4 text-right align-top"), Div(Class("flex flex-wrap justify-end gap-2"), Group(actions))),
	)
}
