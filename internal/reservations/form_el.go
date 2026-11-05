package reservations

import (
	"net/url"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func reservationForm(inquiry models.Inquiry, state *formState, owner bool, returnTo string) Node {
	values := url.Values{}
	if state == nil {
		values = inquiryFormValues(inquiry)
	} else {
		values = state.Values
	}
	isPending := inquiry.Status == models.InquiryPending
	id := strconv.FormatInt(inquiry.ID, 10)
	action := "/admin/inquiries/" + id
	if owner {
		action = "/owner/reservations/" + id
	}
	stayType := values.Get("stay_type")
	dateToLabel := "Salida"
	if stayType == string(models.StayDayUse) {
		dateToLabel = "Fecha de pasadía"
	}
	balanceLabel := "Saldo pendiente"
	balanceClass := "text-brand-dark"
	if inquiry.CommissionBalance < 0 {
		balanceLabel = "Saldo a favor"
		balanceClass = "text-forest"
	} else if inquiry.CommissionBalance == 0 {
		balanceLabel = "Saldo al día"
		balanceClass = "text-forest"
	}
	headingTitle := "Editar reserva"
	submitLabel := "Guardar cambios"
	if isPending {
		headingTitle = "Editar solicitud"
		submitLabel = "Guardar solicitud"
	}
	heading := Div(Class("flex items-start justify-between gap-4"), components.PageHeading(headingTitle, inquiry.CabinName+" · "+helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo)),
		Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), Text("×")),
	)
	return Div(ID("reservation-form-content"), Class("w-full px-4 py-5 sm:px-6 sm:py-6"),
		heading,
		noticeEl(state),
		Form(Method("post"), Action(action), Attr("hx-post", action), Attr("hx-target", "#reservation-form-content"), Attr("hx-swap", "outerHTML"), Class("grid gap-4 sm:grid-cols-2"),
			If(returnTo == "reservations", Input(Type("hidden"), Name("return_to"), Value(returnTo))),
			components.TextFieldWithError("guest-name", "guest_name", "Nombre del huésped", values.Get("guest_name"), "", "text", true, fieldError(state, "guest_name")),
			components.TextFieldWithError("guest-phone", "guest_phone", "Teléfono", values.Get("guest_phone"), "", "tel", true, fieldError(state, "guest_phone")),
			Div(Class("sm:col-span-2"), components.TextFieldWithError("guest-email", "guest_email", "Correo electrónico", values.Get("guest_email"), "", "email", true, fieldError(state, "guest_email"))),
			Div(Class("min-w-0"),
				components.FieldLabel("stay-type", "Tipo de reserva"),
				Select(ID("stay-type"), Name("stay_type"), Required(), If(fieldError(state, "stay_type") != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "stay-type-error")}), Class("w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"),
					Option(Value(string(models.StayDayUse)), If(stayType == string(models.StayDayUse), Selected()), Text("Pasadía")),
					Option(Value(string(models.StayOvernight)), If(stayType == string(models.StayOvernight), Selected()), Text("Estadía")),
				),
				components.FieldError("stay_type", fieldError(state, "stay_type")),
			),
			components.TextFieldWithError("date-from", "date_from", "Llegada", values.Get("date_from"), "", "date", true, fieldError(state, "date_from")),
			components.TextFieldWithError("date-to", "date_to", dateToLabel, values.Get("date_to"), "", "date", true, fieldError(state, "date_to")),
			components.NumberFieldWithError("guests", "guests", "Huéspedes", values.Get("guests"), "1", "", fieldError(state, "guests")),
			If(isPending, Div(Class("sm:col-span-2 rounded-xl bg-forest-soft p-4"),
				P(Class("text-sm font-extrabold text-forest"), Text("Estimado actual: $"+helpers.FormatCOP(inquiry.Estimated)+" COP")),
				P(Class("mt-1 text-sm leading-6 text-muted"), Text("Al guardar, recalcularemos el estimado con las tarifas actuales y el depósito guardado.")),
			)),
			If(!isPending, Div(Class("sm:col-span-2 rounded-xl bg-cream p-4"),
				components.NumberFieldWithError("final-total", "final_total", "Total acordado, incluido depósito (COP)", values.Get("final_total"), strconv.FormatInt(inquiry.DepositCOP, 10), "1000000000000", fieldError(state, "final_total")),
				Label(Class("mt-3 flex items-start gap-2 text-sm font-semibold text-ink"),
					Input(Type("checkbox"), Name("use_final_total"), Value("on"), If(values.Get("use_final_total") == "on", Checked()), Class("mt-1 size-4 accent-brand-dark")),
					Text("Usar este total acordado en vez del cálculo automático"),
				),
				P(Class("mt-2 text-xs leading-5 text-muted"), Text("El total incluye el depósito; la comisión se calcula solo sobre alojamiento. Sin usar este total, cambios en huéspedes, fechas o tipo recalculan con tarifas actuales y el depósito guardado. El cargo por huésped aplica solo por encima de los incluidos.")),
				P(Class("mt-2 text-sm font-bold text-ink"), Text("Comisión guardada: "+helpers.FormatCommissionRate(inquiry.CommissionBP)+"% · $"+helpers.FormatCOP(inquiry.Commission)+" COP")),
				P(Class("mt-1 text-sm font-bold "+balanceClass), Text(balanceLabel+": $"+helpers.FormatCOP(abs(inquiry.CommissionBalance))+" COP")),
				If(len(inquiry.CommissionAdjustments) > 0, Div(Class("mt-3 border-t border-line pt-3"),
					P(Class("text-xs font-extrabold uppercase tracking-wide text-muted"), Text("Ajustes de comisión")),
					Map(inquiry.CommissionAdjustments, func(adjustment models.CommissionAdjustment) Node {
						actor := "Travelando"
						if adjustment.ActorRole == "owner" {
							actor = "Dueño"
						}
						sign := "+"
						if adjustment.Amount < 0 {
							sign = "−"
						}
						return P(Class("mt-1 text-xs text-muted"), Text(adjustment.CreatedAt+" · "+actor+" · $"+helpers.FormatCOP(adjustment.PreviousAmount)+" → $"+helpers.FormatCOP(adjustment.NewAmount)+" ("+sign+"$"+helpers.FormatCOP(abs(adjustment.Amount))+")"))
					}),
				)),
			)),
			Div(Class("flex flex-wrap items-center gap-3 sm:col-span-2"),
				Button(Type("submit"), Class("inline-flex min-h-11 items-center justify-center rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Text(submitLabel)),
				Button(Type("button"), Attr("data-action", "click->modal#close"), Class("font-bold text-brand-dark hover:underline"), Text("Cancelar")),
			),
		),
	)
}

func noticeEl(state *formState) Node {
	if state == nil || state.Notice == "" {
		return Div()
	}
	return Div(Attr("role", "alert"), Class("mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm font-semibold text-red-800"), Text(state.Notice))
}

func inquiryFormValues(inquiry models.Inquiry) url.Values {
	dateTo := inquiry.DateTo
	if inquiry.StayType == models.StayDayUse {
		dateTo = inquiry.DateFrom
	}
	return url.Values{
		"guest_name": {inquiry.GuestName}, "guest_phone": {inquiry.GuestPhone}, "guest_email": {inquiry.GuestEmail},
		"stay_type": {string(inquiry.StayType)}, "date_from": {inquiry.DateFrom}, "date_to": {dateTo},
		"guests": {strconv.Itoa(inquiry.Guests)}, "final_total": {strconv.FormatInt(inquiry.FinalTotal, 10)},
	}
}

func fieldError(state *formState, field string) string {
	if state == nil {
		return ""
	}
	return state.Errors.ErrorStringFor(field)
}

func abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
