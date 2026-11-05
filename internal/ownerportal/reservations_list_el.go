package ownerportal

import (
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ownerReservationsPageEl(inquiries []models.Inquiry) Node {
	return Div(Class("mx-auto w-full max-w-7xl px-5 py-8 sm:px-8 sm:py-12"),
		components.PageHeading("Reservas de mis cabañas", "Consulta solicitudes pendientes, reservas confirmadas y cancelaciones."),
		ownerReservationsContentEl(inquiries),
	)
}

func ownerReservationsContentEl(inquiries []models.Inquiry) Node {
	return Div(ID("owner-reservations-content"), Class("overflow-hidden rounded-2xl border border-line bg-white"),
		Div(Class("flex flex-wrap items-center justify-between gap-3 border-b border-line p-5 sm:px-6"),
			H2(Class("text-xl font-extrabold text-ink"), Text("Todas las reservas")),
			P(Class("text-sm text-muted"), Textf("%d registros", len(inquiries))),
		),
		If(len(inquiries) == 0, Div(Attr("role", "status"), Class("px-6 py-12 text-center"),
			P(Class("text-lg font-extrabold text-ink"), Text("Aún no tienes reservas")),
			P(Class("mt-1 text-sm text-muted"), Text("Las solicitudes y reservas aparecerán aquí.")),
		)),
		If(len(inquiries) > 0, Div(Class("overflow-x-auto"),
			Table(Class("w-full min-w-[900px] text-left text-sm"),
				THead(Class("bg-cream/70 text-xs uppercase tracking-wider text-muted"), Tr(
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Huésped")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Cabaña")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Fechas y estadía")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Total / estimado")),
					Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Estado")),
					Th(Scope("col"), Class("px-5 py-3 text-right font-bold"), Text("Acciones")),
				)),
				TBody(Class("divide-y divide-line"), Map(inquiries, ownerReservationRow)),
			),
		)),
	)
}

func ownerReservationRow(inquiry models.Inquiry) Node {
	id := strconv.FormatInt(inquiry.ID, 10)
	status, statusClass, totalLabel, total := "Confirmada", "bg-forest text-white", "Total", inquiry.FinalTotal
	if inquiry.Status == models.InquiryPending {
		status, statusClass, totalLabel, total = "Pendiente", "bg-cream text-brand-dark", "Estimado", inquiry.Estimated
	} else if inquiry.Status == models.InquiryCancelled {
		status, statusClass = "Cancelada", "bg-gray-100 text-muted"
		if inquiry.CancelledFromStatus == models.InquiryPending {
			totalLabel, total = "Estimado", inquiry.Estimated
		}
	}
	detailsURL := "/owner/reservations/" + id
	actions := []Node{
		A(Href(detailsURL), Attr("hx-get", detailsURL), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("inline-flex min-h-9 items-center rounded-lg border border-line px-3 text-xs font-bold text-brand-dark hover:border-brand hover:bg-cream"), Text("Ver")),
	}
	if inquiry.Status == models.InquiryPending || inquiry.Status == models.InquiryReserved {
		editURL := "/owner/reservations/" + id + "/edit?return_to=reservations"
		confirm := "¿Cancelar esta reserva? Las fechas quedarán disponibles."
		if inquiry.Status == models.InquiryPending {
			confirm = "¿Cancelar esta solicitud?"
		}
		actions = append(actions,
			A(Href(editURL), Attr("hx-get", editURL), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("inline-flex min-h-9 items-center rounded-lg border border-line px-3 text-xs font-bold text-ink hover:border-brand hover:bg-cream"), Text("Editar")),
			Form(Method("post"), Action("/owner/reservations/"+id+"/cancel"), Attr("hx-post", "/owner/reservations/"+id+"/cancel"), Attr("hx-target", "#owner-reservations-content"), Attr("hx-swap", "outerHTML"), Attr("hx-confirm", confirm),
				Button(Type("submit"), Class("inline-flex min-h-9 items-center rounded-lg border border-red-200 px-3 text-xs font-bold text-red-800 hover:bg-red-50"), Text("Cancelar")),
			),
		)
	}
	return Tr(ID("owner-reservation-row-"+id),
		Td(Class("px-5 py-4 align-top"), P(Class("font-bold text-ink"), Text(inquiry.GuestName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.GuestPhone))),
		Td(Class("px-5 py-4 align-top font-bold text-ink"), Text(inquiry.CabinName)),
		Td(Class("px-5 py-4 align-top text-muted"), P(Text(helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo))), P(Class("mt-1 text-xs"), Text(helpers.StayTypeLabel(inquiry.StayType)+" · "+strconv.Itoa(inquiry.Guests)+" huéspedes"))),
		Td(Class("px-5 py-4 align-top"), P(Class("text-xs text-muted"), Text(totalLabel)), P(Class("font-bold text-ink"), Text("$"+helpers.FormatCOP(total)+" COP"))),
		Td(Class("px-5 py-4 align-top"), Span(Class("inline-flex rounded-full px-2.5 py-1 text-xs font-bold "+statusClass), Text(status))),
		Td(Class("px-5 py-4 text-right align-top"), Div(Class("flex flex-wrap justify-end gap-2"), Group(actions))),
	)
}
