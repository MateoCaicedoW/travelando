package reservations

import (
	"net/url"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/models"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	"go.leapkit.dev/core/form/validate"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func newReservationFormEl(state *reservationFormState) Node {
	fieldErrors := validate.Errors(nil)
	values := url.Values(nil)
	if state != nil {
		fieldErrors, values = state.Errors, state.Values
	}
	fieldValue := func(key, fallback string) string {
		if values != nil {
			return values.Get(key)
		}
		return fallback
	}
	cabinIDValue := fieldValue("cabin_id", "")
	guestNameValue := fieldValue("guest_name", "")
	guestCedulaValue := fieldValue("guest_cedula", "")
	guestPhoneValue := fieldValue("guest_phone", "")
	guestEmailValue := fieldValue("guest_email", "")
	stayTypeValue := fieldValue("stay_type", string(models.StayDayUse))
	dateFromValue := fieldValue("date_from", "")
	dateToValue := fieldValue("date_to", "")
	guestsValue := fieldValue("guests", "2")
	depositValue := fieldValue("deposit_cop", "")

	cabinOptions := make([]components.SearchableOption, 0, len(state.Cabins))
	for _, c := range state.Cabins {
		cabinOptions = append(cabinOptions, components.SearchableOption{
			Value: strconv.FormatInt(c.ID, 10),
			Label: c.Name + " · " + c.OwnerName,
		})
	}

	return Div(Class("p-6 sm:p-8 flex flex-col gap-6"),
		Div(Class("flex items-start justify-between gap-4"),
			components.PageHeading("Nueva reserva", "Crea una reserva confirmada con datos de huésped. Se bloquearán las fechas de la cabaña."),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		),
		Form(ID("new-reservation"), Method("post"), Action("/admin/reservations"), hx.Post("/admin/reservations"), hx.Target("#modal"), hx.Swap("outerHTML"),
			Div(Class("grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"),
				cabinSelect(fieldErrors, cabinIDValue, cabinOptions, "lg:col-span-3 sm:col-span-2"),
				Div(Class("col-span-1 sm:col-span-2 lg:col-span-3"),
					components.TextFieldWithError("guest_name", "guest_name", "Nombre del huésped", guestNameValue, "Ej. María García", "text", true, firstError(fieldErrors, "guest_name")),
				),
				Div(Class("col-span-1 sm:col-span-1 lg:col-span-1"),
					components.TextFieldWithError("guest_cedula", "guest_cedula", "Cédula", guestCedulaValue, "Ej. 12345678", "text", true, firstError(fieldErrors, "guest_cedula")),
				),
				Div(Class("col-span-1 sm:col-span-1 lg:col-span-1"),
					components.TextFieldWithError("guest_phone", "guest_phone", "Teléfono", guestPhoneValue, "Ej. 3001234567", "tel", true, firstError(fieldErrors, "guest_phone")),
				),
				Div(Class("col-span-1 sm:col-span-2 lg:col-span-1"),
					components.TextFieldWithError("guest_email", "guest_email", "Correo electrónico", guestEmailValue, "Ej. maria@correo.com", "email", true, firstError(fieldErrors, "guest_email")),
				),
				Div(Class("col-span-1 sm:col-span-2 lg:col-span-1"),
					Div(Class("min-w-0"),
						Label(For("stay_type"), Class("mb-1.5 block text-[0.8125rem] font-bold text-ink"), Text("Tipo de estadía")),
						Div(Class("flex items-center gap-3 border rounded-xl border-line bg-white px-3 py-2 focus-within:border-brand focus-within:ring-2 focus-within:ring-brand/20"),
							Div(Class("flex items-center gap-2"),
								Input(Type("radio"), Name("stay_type"), Value(string(models.StayDayUse)), If(stayTypeValue == string(models.StayDayUse), Checked()), Required(), Class("w-4 h-4 text-brand border-ink")),
								Label(Class("text-sm text-ink"), Text("Pasadía")),
							),
							Div(Class("flex items-center gap-2 ml-4"),
								Input(Type("radio"), Name("stay_type"), Value(string(models.StayOvernight)), If(stayTypeValue == string(models.StayOvernight), Checked()), Required(), Class("w-4 h-4 text-brand border-ink")),
								Label(Class("text-sm text-ink"), Text("Estadía")),
							),
						),
						components.FieldError("stay_type", firstError(fieldErrors, "stay_type")),
					),
				),
				Div(Class("col-span-1 sm:col-span-1 lg:col-span-1"),
					Div(Class("min-w-0"),
						Label(For("date_from"), Class("mb-1.5 block text-[0.8125rem] font-bold text-ink"), Text("Fecha inicial")),
						Input(ID("date_from"), Name("date_from"), Type("date"), Value(dateFromValue), Required(), If(firstError(fieldErrors, "date_from") != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "date_from-error")}), Class("w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20 border-line border")),
						components.FieldError("date_from", firstError(fieldErrors, "date_from")),
					),
				),
				Div(Class("col-span-1 sm:col-span-1 lg:col-span-1"),
					Div(Class("min-w-0"),
						Label(For("date_to"), Class("mb-1.5 block text-[0.8125rem] font-bold text-ink"), Text("Fecha final")),
						Input(ID("date_to"), Name("date_to"), Type("date"), Value(dateToValue), Required(), If(firstError(fieldErrors, "date_to") != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "date_to-error")}), Class("w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20 border-line border")),
						components.FieldError("date_to", firstError(fieldErrors, "date_to")),
					),
				),
				Div(Class("col-span-1 sm:col-span-1 lg:col-span-1"),
					components.NumberFieldWithError("guests", "guests", "Huéspedes", guestsValue, "1", "20", firstError(fieldErrors, "guests")),
				),
				Div(Class("col-span-1 sm:col-span-1 lg:col-span-1"),
					components.NumberFieldWithError("deposit_cop", "deposit_cop", "Depósito (COP)", depositValue, "0", "", firstError(fieldErrors, "deposit_cop")),
				),
			),
			components.FormActions(
				Button(Type("button"), Attr("data-action", "click->modal#close"), Class(components.FormSecondaryClass), Text("Cancelar")),
				Button(Type("submit"), Class(components.FormPrimaryClass), Text("Crear reserva")),
			),
		),
	)
}

func cabinSelect(fieldErrors validate.Errors, value string, options []components.SearchableOption, spanClass string) Node {
	return components.SearchableSelect("cabin_id", "cabin_id", "Cabaña", "Selecciona una cabaña", value, options, true, firstError(fieldErrors, "cabin_id"), spanClass)
}

func firstError(fieldErrors validate.Errors, field string) string {
	if msgs, ok := fieldErrors[field]; ok && len(msgs) > 0 {
		return msgs[0].Error()
	}
	return ""
}
