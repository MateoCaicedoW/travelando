package reservations

import (
	"strconv"

	lucide "github.com/eduardolat/gomponents-lucide"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ReservationDetailsEl(inquiry models.Inquiry) Node {
	return reservationDetailsEl(inquiry, false)
}

func OwnerReservationDetailsEl(inquiry models.Inquiry) Node {
	return reservationDetailsEl(inquiry, true)
}

func reservationDetailsEl(inquiry models.Inquiry, owner bool) Node {
	id := strconv.FormatInt(inquiry.ID, 10)
	pending := inquiry.Status == models.InquiryPending || inquiry.Status == models.InquiryCancelled && inquiry.CancelledFromStatus == models.InquiryPending
	status, statusClass, title := "Reserva confirmada", "bg-forest text-white", "Detalles de reserva"
	if inquiry.Status == models.InquiryPending {
		status, statusClass, title = "Solicitud pendiente", "bg-cream text-brand-dark", "Detalles de solicitud"
	}
	if inquiry.Status == models.InquiryCancelled {
		status, statusClass = "Reserva cancelada", "bg-gray-100 text-muted"
		if pending {
			status, title = "Solicitud cancelada", "Detalles de solicitud"
		}
	}
	stayType := helpers.StayTypeLabel(inquiry.StayType)
	balanceLabel := "Al día"
	if inquiry.CommissionBalance > 0 {
		balanceLabel = "Pendiente $" + helpers.FormatCOP(inquiry.CommissionBalance)
	} else if inquiry.CommissionBalance < 0 {
		balanceLabel = "Saldo a favor $" + helpers.FormatCOP(-inquiry.CommissionBalance)
	}
	values := detailSection("Valores",
		detailLine("Total acordado", "$"+helpers.FormatCOP(inquiry.FinalTotal)+" COP"),
		detailLine("Depósito", "$"+helpers.FormatCOP(inquiry.DepositCOP)+" COP"),
		detailLine("Comisión", "$"+helpers.FormatCOP(inquiry.Commission)+" COP"),
		detailLine("Comisión pagada", "$"+helpers.FormatCOP(inquiry.CommissionPaidAmount)+" COP"),
		detailLine("Saldo", balanceLabel),
	)
	if pending {
		values = detailSection("Valores",
			detailLine("Total estimado", "$"+helpers.FormatCOP(inquiry.Estimated)+" COP"),
			detailLine("Depósito", "$"+helpers.FormatCOP(inquiry.DepositCOP)+" COP"),
			P(Class("text-xs leading-5 text-muted"), Text("El total final se define al confirmar la reserva.")),
		)
	}
	editURL := "/admin/inquiries/" + id + "/edit?return_to=reservations"
	cancelURL, listTarget := "/admin/reservations/"+id+"/cancel", "#reservations-content"
	if owner {
		editURL = "/owner/reservations/" + id + "/edit?return_to=reservations"
		cancelURL, listTarget = "/owner/reservations/"+id+"/cancel", "#owner-reservations-content"
	}
	editLabel, cancelLabel, cancelConfirm := "Editar reserva", "Cancelar reserva", "¿Cancelar esta reserva? Sus fechas quedarán disponibles."
	if pending {
		editLabel, cancelLabel, cancelConfirm = "Editar solicitud", "Cancelar solicitud", "¿Cancelar esta solicitud?"
	}
	return Div(Class("w-full p-5 sm:p-8"),
		Div(Class("mb-6 flex items-start justify-between gap-4"),
			Div(Class("min-w-0"),
				Div(Class("flex flex-wrap items-center gap-x-3 gap-y-2"),
					H1(Class("text-3xl font-extrabold tracking-tight text-ink sm:text-4xl"), Text(title)),
					Span(Class("inline-flex rounded-full px-3 py-1 text-xs font-bold "+statusClass), Text(status)),
				),
				P(Class("mt-2 text-base leading-7 text-muted"), Text(inquiry.CabinName+" · "+helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo))),
			),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		),
		Div(Class("grid gap-4 md:grid-cols-2"),
			detailSection("Huésped", detailLine("Nombre", inquiry.GuestName), detailLine("Teléfono", inquiry.GuestPhone), detailLine("Correo", inquiry.GuestEmail)),
			detailSection("Estadía", detailLine("Cabaña", inquiry.CabinName), detailLine("Dueño", inquiry.OwnerName), detailLine("Tipo", stayType), detailLine("Horario", helpers.ScheduleLabel(inquiry.StayType, inquiry.StartTime, inquiry.EndTime)), detailLine("Huéspedes", strconv.Itoa(inquiry.Guests))),
			detailSection("Fechas", detailLine("Estadía", helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo)), detailLine("Limpieza", strconv.Itoa(inquiry.CleaningHours)+" horas")),
			values,
			If(len(inquiry.CommissionAdjustments) > 0, detailSection("Ajustes de comisión", Map(inquiry.CommissionAdjustments, func(adjustment models.CommissionAdjustment) Node {
				actor := "Travelando"
				if adjustment.ActorRole == "owner" {
					actor = "Dueño"
				}
				sign := "+"
				if adjustment.Amount < 0 {
					sign = "−"
				}
				return Div(Class("border-b border-line/70 pb-2 last:border-0 last:pb-0"),
					P(Class("text-xs text-muted"), Text(adjustment.CreatedAt+" · "+actor)),
					P(Class("mt-1 text-sm font-bold text-ink"), Text("$"+helpers.FormatCOP(adjustment.PreviousAmount)+" → $"+helpers.FormatCOP(adjustment.NewAmount)+" ("+sign+"$"+helpers.FormatCOP(absReservationAmount(adjustment.Amount))+")")),
				)
			}))),
		),
		If(inquiry.Status == models.InquiryReserved || inquiry.Status == models.InquiryPending, Div(Class("mt-6 flex flex-wrap gap-3 border-t border-line pt-5"),
			A(Href(editURL), Attr("hx-get", editURL), Attr("hx-target", "#modal"), Attr("hx-swap", "outerHTML"), Attr("hx-disabled-elt", "this"), Class("inline-flex min-h-11 items-center justify-center rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white hover:brightness-90"), Text(editLabel)),
			Form(Method("post"), Action(cancelURL), Attr("hx-post", cancelURL), Attr("hx-target", listTarget), Attr("hx-swap", "outerHTML"), Attr("hx-confirm", cancelConfirm), Class("inline-flex"),
				Button(Type("submit"), Class("inline-flex min-h-11 items-center justify-center rounded-lg border border-red-200 px-5 py-2.5 font-bold text-red-800 hover:bg-red-50"), Text(cancelLabel)),
			),
		)),
	)
}

func detailSection(title string, rows ...Node) Node {
	return Section(Attr("aria-label", title), Class("rounded-xl border border-line bg-white p-4"),
		H2(Class("mb-3 text-sm font-extrabold uppercase tracking-wide text-brand-dark"), Text(title)),
		Div(Class("grid gap-2"), Group(rows)),
	)
}

func detailLine(label, value string) Node {
	return Div(Class("flex flex-wrap justify-between gap-x-4 gap-y-1 border-b border-line/70 pb-2 last:border-0 last:pb-0"),
		Span(Class("text-sm text-muted"), Text(label)),
		Span(Class("text-right text-sm font-bold text-ink"), Text(value)),
	)
}

func absReservationAmount(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
