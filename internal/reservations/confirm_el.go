package reservations

import (
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func confirmReservationForm(inquiry models.Inquiry, owner bool, total string, notice string) Node {
	id := strconv.FormatInt(inquiry.ID, 10)
	action := "/admin/reservations/" + id + "/confirm"
	if owner {
		action = "/owner/reservations/" + id + "/confirm"
	}
	return Form(
		ID("confirm-reservation-form"),
		Method("post"), Action(action), hx.Post(action), hx.Target("#confirm-reservation-form"), hx.Swap("outerHTML"),
		Class("w-full p-5 sm:p-8"),
		Div(Class("mb-6 flex items-start justify-between gap-4"),
			Div(Class("min-w-0"),
				H1(Class("text-3xl font-extrabold tracking-tight text-ink sm:text-4xl"), Text("Confirmar reserva")),
				P(Class("mt-2 text-base leading-7 text-muted"), Text(inquiry.CabinName+" · "+helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo))),
			),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		),
		Div(Class("grid gap-4 rounded-xl border border-line bg-cream p-4 sm:grid-cols-2"),
			confirmationDetail("Huésped", inquiry.GuestName),
			confirmationDetail("Depósito incluido", "$"+helpers.FormatCOP(inquiry.DepositCOP)+" COP"),
			confirmationDetail("Estimado actual", "$"+helpers.FormatCOP(inquiry.Estimated)+" COP"),
		),
		P(Class("mt-4 text-sm leading-6 text-muted"), Text("Ingresa el total acordado con el huésped, incluido el depósito. Al confirmar, se bloquearán las fechas y se calculará la comisión.")),
		If(notice != "", P(Attr("role", "alert"), Class("mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm font-semibold text-red-800"), Text(notice))),
		Div(Class("mt-4 max-w-sm"),
			components.NumberFieldWithError("final-total", "final_total", "Total acordado (COP)", total, strconv.FormatInt(inquiry.DepositCOP, 10), strconv.FormatInt(services.MaxAmountCOP, 10), ""),
		),
		components.FormActions(
			Button(Type("submit"), Class(components.FormPrimaryClass), Text("Confirmar reserva")),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Class(components.FormSecondaryClass), Text("Cancelar")),
		),
	)
}

func confirmationDetail(label, value string) Node {
	return Div(P(Class("text-label font-bold uppercase tracking-wide text-muted"), Text(label)), P(Class("mt-1 text-sm font-extrabold text-ink"), Text(value)))
}

func confirmReservationDialog(inquiry models.Inquiry, owner bool, total, notice string) Node {
	return components.CustomDialog(components.DialogProps{Label: "Confirmar reserva", Yield: confirmReservationForm(inquiry, owner, total, notice)})
}
