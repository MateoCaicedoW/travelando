package requests

import (
	"net/url"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func reportPage(inquiry models.Inquiry, token string, values url.Values, fieldErrors validate.Errors) Node {
	dateText := helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo)
	stay := helpers.StayTypeLabel(inquiry.StayType)
	statusLabel := "Pendiente de confirmar"
	if inquiry.Status == models.InquiryReserved {
		statusLabel = "Reserva confirmada"
	} else if inquiry.Status == models.InquiryLost {
		statusLabel = "No se concretó"
	} else if inquiry.Status == models.InquiryCancelled {
		statusLabel = "Reserva cancelada"
		if inquiry.CancelledFromStatus == models.InquiryPending {
			statusLabel = "Solicitud cancelada"
		}
	}
	statusValue := string(models.InquiryReserved)
	totalValue := strconv.FormatInt(inquiry.Estimated, 10)
	if values != nil {
		statusValue = values.Get("status")
		totalValue = values.Get("final_total")
	}
	statusClass := "w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"
	if fieldErrors.Has("status") {
		statusClass += " border-red-300 focus:ring-red-200"
	}
	totalClass := "w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"
	if fieldErrors.Has("final_total") {
		totalClass += " border-red-300 focus:ring-red-200"
	}
	return Div(ID("report-page"), Class("mx-auto w-full max-w-2xl px-5 py-10 sm:px-8 sm:py-16"),
		components.PageHeading("Resultado de la consulta", "Este enlace corresponde solo a esta solicitud."),
		Div(Class("rounded-2xl border border-line bg-white p-5 shadow-sm sm:p-7"),
			Div(Class("flex flex-wrap items-start justify-between gap-3 border-b border-line pb-5"),
				Div(
					P(Class("text-xs font-extrabold uppercase tracking-[0.16em] text-brand-dark"), Text(inquiry.OwnerName+" · "+inquiry.CabinName)),
					H2(Class("mt-1 text-xl font-extrabold text-ink"), Text(inquiry.GuestName)),
				),
				Span(Class("rounded-full bg-cream px-3 py-1.5 text-xs font-extrabold text-brand-dark"), Text(statusLabel)),
			),
			Div(Class("grid gap-4 py-5 sm:grid-cols-2"),
				SummaryItem("Plan y fechas", stay+" · "+dateText),
				SummaryItem("Horario", helpers.ScheduleLabel(inquiry.StayType, inquiry.StartTime, inquiry.EndTime)),
				SummaryItem("Huéspedes", strconv.Itoa(inquiry.Guests)),
				SummaryItem("Precio estimado", "$"+helpers.FormatCOP(inquiry.Estimated)+" COP"),
				If(inquiry.DepositCOP > 0, SummaryItem("Depósito incluido", "$"+helpers.FormatCOP(inquiry.DepositCOP)+" COP")),
				If(inquiry.Status == models.InquiryReserved, SummaryItem("Total acordado", "$"+helpers.FormatCOP(inquiry.FinalTotal)+" COP")),
				If(inquiry.Status == models.InquiryReserved, SummaryItem("Comisión Travelando", "$"+helpers.FormatCOP(inquiry.Commission)+" COP")),
			),
			If(inquiry.Status == models.InquiryPending, Form(Method("post"), Action("/report/"+token), Attr("hx-post", "/report/"+token), Attr("hx-target", "#report-page"), Attr("hx-swap", "outerHTML"), Class("border-t border-line pt-5"),
				P(Class("text-sm leading-6 text-muted"), Text("Cuando terminen de hablar, registra si se concretó. Si reservaron, ingresa el total acordado incluyendo depósito; la comisión se calcula solo sobre alojamiento.")),
				Div(Class("mt-4 max-w-xs"),
					components.FieldLabel("status", "Resultado"),
					Select(ID("status"), Name("status"), Required(), If(fieldErrors.Has("status"), Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "status-error")}), Class(statusClass),
						Option(Value(string(models.InquiryReserved)), If(statusValue == string(models.InquiryReserved), Selected()), Text("Se reservó")),
						Option(Value(string(models.InquiryLost)), If(statusValue == string(models.InquiryLost), Selected()), Text("No se concretó")),
					),
					components.FieldError("status", fieldErrors.ErrorStringFor("status")),
				),
				Div(Class("mt-4 max-w-xs"),
					components.FieldLabel("final_total", "Valor total de la reserva (COP)"),
					Input(ID("final_total"), Name("final_total"), Type("number"), Min(strconv.FormatInt(inquiry.DepositCOP, 10)), Max("1000000000000"), Step("1"), Value(totalValue), If(fieldErrors.Has("final_total"), Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "final_total-error")}), Class(totalClass)),
					components.FieldError("final_total", fieldErrors.ErrorStringFor("final_total")),
				),
				Button(Type("submit"), Class("mt-5 inline-flex min-h-11 items-center justify-center rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Text("Guardar resultado")),
			)),
		),
	)
}

func SummaryItem(label, value string) Node {
	return Div(Class("rounded-xl bg-cream p-4"),
		P(Class("text-xs font-bold uppercase tracking-wider text-muted"), Text(label)),
		P(Class("mt-1 text-sm font-extrabold text-ink"), Text(value)),
	)
}
