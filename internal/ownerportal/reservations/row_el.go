package reservations

import (
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func ReservationRowEl(inquiry models.Inquiry) Node {
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
		A(Href(detailsURL), Attr("role", "menuitem"), hx.Get(detailsURL), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-cream"), Text("Ver")),
	}
	if inquiry.Status == models.InquiryPending || inquiry.Status == models.InquiryReserved {
		editURL := "/owner/reservations/" + id + "/edit?return_to=reservations"
		if inquiry.Status == models.InquiryPending {
			confirmURL := detailsURL + "/confirm"
			actions = append(actions, A(Href(confirmURL), Attr("role", "menuitem"), hx.Get(confirmURL), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-cream"), Text("Confirmar reserva")))
		}
		confirm := "¿Cancelar esta reserva? Las fechas quedarán disponibles."
		if inquiry.Status == models.InquiryPending {
			confirm = "¿Cancelar esta solicitud?"
		}
		actions = append(actions,
			A(Href(editURL), Attr("role", "menuitem"), hx.Get(editURL), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-cream"), Text("Editar")),
			Form(Method("post"), Action("/owner/reservations/"+id+"/cancel"), hx.Post("/owner/reservations/"+id+"/cancel"), hx.Target("#owner-reservations-content"), hx.Swap("outerHTML"), hx.Confirm(confirm),
				Button(Type("submit"), Attr("role", "menuitem"), Class("flex min-h-10 w-full items-center rounded-md px-3 py-2 text-left text-sm font-semibold text-red-800 hover:bg-red-50"), Text("Cancelar")),
			),
		)
	}
	return Tr(ID("owner-reservation-row-"+id),
		Td(Attr("data-label", "Huésped"), Class("px-4 py-3 align-top"), P(Class("font-bold text-ink"), Text(inquiry.GuestName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.GuestPhone))),
		Td(Attr("data-label", "Cabaña"), Class("px-4 py-3 align-top font-bold text-ink"), Text(inquiry.CabinName)),
		Td(Attr("data-label", "Fechas y estadía"), Class("px-4 py-3 align-top text-muted"), P(Text(helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo))), P(Class("mt-1 text-xs"), Text(helpers.StayTypeLabel(inquiry.StayType)+" · "+strconv.Itoa(inquiry.Guests)+" huéspedes"))),
		Td(Attr("data-label", "Total / estimado"), Class("px-4 py-3 align-top"), P(Class("text-xs text-muted"), Text(totalLabel)), P(Class("font-bold text-ink"), Text("$"+helpers.FormatCOP(total)+" COP"))),
		Td(Attr("data-label", "Estado"), Class("px-4 py-3 align-top"), Span(Class("inline-flex rounded-full px-2.5 py-1 text-xs font-bold "+statusClass), Text(status))),
		Td(Attr("data-label", "Acciones"), Class("px-4 py-3 text-right align-top"), components.ActionDropdown("owner-reservation-"+id, actions...)),
	)
}
