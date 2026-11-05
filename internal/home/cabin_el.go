package home

import (
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func CabinEl(cabin models.Cabin, data PageData) Node {
	stayType := "Pasadía"
	selectedStayType := data.StayType
	if selectedStayType == models.StayDayUse && !cabin.AcceptsDayUse && cabin.AcceptsOvernight || selectedStayType == models.StayOvernight && !cabin.AcceptsOvernight && cabin.AcceptsDayUse {
		if cabin.AcceptsOvernight {
			selectedStayType = models.StayOvernight
		} else {
			selectedStayType = models.StayDayUse
		}
	}
	if selectedStayType == models.StayOvernight {
		stayType = "Estadía"
	}
	estimated, estimateValid := priceFor(cabin, selectedStayType, data.DateFrom, data.DateTo, data.Guests, data.DepositCOP)
	estimatedLabel := "$" + helpers.FormatCOP(estimated)
	if !estimateValid {
		estimatedLabel = "—"
	}
	priceUnit := "COP · pasadía"
	if selectedStayType == models.StayOvernight {
		priceUnit = "COP / noche"
		if data.DateFrom != "" && data.DateTo != "" {
			priceUnit = "COP · estadía completa"
		}
	}
	requestValue := func(key, fallback string) string {
		if data.RequestValues != nil {
			return data.RequestValues.Get(key)
		}
		return fallback
	}
	requestErrors := validate.Errors(nil)
	if data.RequestErrors != nil {
		requestErrors = data.RequestErrors
	}
	requestStayType := requestValue("stay_type", string(data.StayType))
	if requestStayType == string(models.StayOvernight) && !cabin.AcceptsOvernight || requestStayType == string(models.StayDayUse) && !cabin.AcceptsDayUse {
		requestStayType = string(selectedStayType)
	}
	requestFrom := requestValue("date_from", data.DateFrom)
	requestTo := requestValue("date_to", data.DateTo)
	requestGuests := requestValue("guests", strconv.Itoa(data.Guests))
	depositValue, depositDetail := "Sin depósito", "No se requiere depósito."
	if data.DepositCOP > 0 {
		depositValue, depositDetail = "$"+helpers.FormatCOP(data.DepositCOP)+" COP", "Monto fijo por reserva; incluido en el estimado."
	}
	return Div(ID("cabin-detail"), Class("mx-auto w-full max-w-7xl px-5 pt-8 pb-32 sm:px-8 sm:py-12"),
		A(Href("/"), Class("text-sm font-bold text-brand-dark hover:underline"), Text("← Volver a buscar")),
		Div(Class("mt-6 grid items-start gap-8 lg:grid-cols-[1.2fr_0.8fr] lg:gap-12"),
			Div(
				components.CabinPhotoCarousel(cabin.Name, cabin.Photos),
				Div(Class("mt-7 flex flex-wrap items-center justify-between gap-3"),
					Div(
						P(Class("text-xs font-extrabold uppercase tracking-[0.17em] text-brand-dark"), Text(cabin.Location)),
						H1(Class("mt-2 text-3xl font-black tracking-tight text-ink sm:text-4xl"), Text(cabin.Name)),
					),
					Span(Class("rounded-full bg-cream px-4 py-2 text-sm font-bold text-brand-dark"), Textf("%d personas incluidas", cabin.IncludedGuests)),
				),
				P(Class("mt-5 whitespace-pre-line text-base leading-7 text-muted"), Text(cabin.Description)),
				If(len(cabin.Amenities) > 0, Div(Class("mt-7"),
					H2(Class("text-xl font-extrabold text-ink"), Text("Amenidades")),
					Div(Class("mt-3 grid gap-3 sm:grid-cols-2"), Map(cabin.Amenities, func(amenity models.CabinAmenity) Node {
						value := "No"
						if amenity.ValueType == models.AmenityBoolean {
							if amenity.BooleanValue {
								value = "Sí"
							}
						} else {
							value = strconv.Itoa(amenity.QuantityValue)
						}
						return Div(Class("rounded-xl border border-line bg-white px-4 py-3"),
							P(Class("text-sm font-bold text-ink"), Text(amenity.Name)),
							P(Class("mt-1 text-sm text-muted"), Text(value)),
						)
					})),
				)),
				Div(Class("mt-7 grid gap-3 sm:grid-cols-2 xl:grid-cols-3"),
					If(cabin.AcceptsDayUse, DetailItem("Pasadía", "$"+helpers.FormatCOP(cabin.DayPrice)+" COP", cabin.DayStart+" a "+cabin.DayEnd)),
					If(cabin.AcceptsOvernight, DetailItem("Estadía", "$"+helpers.FormatCOP(cabin.NightPrice)+" COP", "Por noche · entrada "+cabin.CheckIn+" · salida "+cabin.CheckOut)),
					DetailItem("Huésped adicional", "$"+helpers.FormatCOP(cabin.ExtraGuestPrice)+" COP", "Por reserva"),
					DetailItem("Depósito", depositValue, depositDetail),
				),
				If(cabin.LocationURL != "", Div(Class("mt-6"),
					A(Href(cabin.LocationURL), Target("_blank"), Rel("noopener noreferrer"), Class("inline-flex min-h-11 items-center gap-2 rounded-lg border border-line px-4 py-2 font-bold text-ink transition hover:border-brand hover:text-brand-dark"), Text("Ver ubicación en el mapa ↗")),
				)),
			),
			Aside(Class("rounded-2xl border border-line bg-white p-5 sm:sticky sm:top-6 sm:p-7"),
				Attr("data-controller", "price-estimate"),
				Attr("data-price-estimate-included-guests-value", strconv.Itoa(cabin.IncludedGuests)),
				Attr("data-price-estimate-extra-guest-price-value", strconv.FormatInt(cabin.ExtraGuestPrice, 10)),
				Attr("data-price-estimate-deposit-cop-value", strconv.FormatInt(data.DepositCOP, 10)),
				Attr("data-price-estimate-day-price-value", strconv.FormatInt(cabin.DayPrice, 10)),
				Attr("data-price-estimate-night-price-value", strconv.FormatInt(cabin.NightPrice, 10)),
				P(Class("text-xs font-extrabold uppercase tracking-[0.16em] text-brand-dark"), Text("Consulta disponibilidad")),
				H2(Class("mt-2 text-2xl font-extrabold text-ink"), Text("Habla con "+cabin.OwnerName)),
				If(!cabin.AcceptsDayUse && !cabin.AcceptsOvernight,
					P(Class("mt-4 rounded-xl border border-line bg-cream p-4 text-sm leading-6 text-muted"), Text("Esta cabaña no está aceptando pasadías ni estadías por ahora.")),
				),
				If(cabin.AcceptsDayUse || cabin.AcceptsOvernight, Group{
					P(Class("mt-2 text-sm leading-6 text-muted"), Text("Enviaremos tu solicitud y abriremos WhatsApp con los detalles listos para enviar.")),
					Form(Method("post"), Action("/requests/"+strconv.FormatInt(cabin.ID, 10)), Attr("hx-post", "/requests/"+strconv.FormatInt(cabin.ID, 10)), Attr("hx-target", "#cabin-detail"), Attr("hx-swap", "outerHTML"), Class("mt-6 space-y-4 pb-[calc(10rem+env(safe-area-inset-bottom))] sm:pb-0"),
						Div(Class("space-y-2"),
							components.FieldLabel("request_stay_type", "Tipo de plan"),
							Select(ID("request_stay_type"), Name("stay_type"), Required(), If(requestErrors.Has("stay_type"), Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "request_stay_type-error")}), Class("w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"),
								If(cabin.AcceptsDayUse, Option(Value(string(models.StayDayUse)), If(requestStayType == string(models.StayDayUse), Selected()), Text("Pasadía"))),
								If(cabin.AcceptsOvernight, Option(Value(string(models.StayOvernight)), If(requestStayType == string(models.StayOvernight), Selected()), Text("Estadía"))),
							),
							components.FieldError("request_stay_type", requestErrors.ErrorStringFor("stay_type")),
						),
						components.TextFieldWithError("request_date_from", "date_from", "Llegada o fecha de uso", requestFrom, "", "date", true, requestErrors.ErrorStringFor("date_from")),
						components.TextFieldWithError("request_date_to", "date_to", "Salida (solo estadía)", requestTo, "", "date", false, requestErrors.ErrorStringFor("date_to")),
						components.NumberFieldWithError("request_guests", "guests", "Huéspedes", requestGuests, "1", "", requestErrors.ErrorStringFor("guests")),
						P(Class("-mt-2 text-xs leading-5 text-muted"), Textf("%d incluidas · $%s COP por huésped adicional, una sola vez por reserva.", cabin.IncludedGuests, helpers.FormatCOP(cabin.ExtraGuestPrice))),
						components.TextFieldWithError("guest_name", "guest_name", "Tu nombre", requestValue("guest_name", ""), "Nombre y apellido", "text", true, requestErrors.ErrorStringFor("guest_name")),
						components.TextFieldWithError("guest_phone", "guest_phone", "Tu teléfono", requestValue("guest_phone", ""), "+57 300 000 0000", "tel", true, requestErrors.ErrorStringFor("guest_phone")),
						components.TextFieldWithError("guest_email", "guest_email", "Correo para la confirmación", requestValue("guest_email", ""), "tu@correo.com", "email", true, requestErrors.ErrorStringFor("guest_email")),
						Div(Class("fixed inset-x-0 bottom-0 z-40 border-t border-line bg-white p-3 pb-[calc(env(safe-area-inset-bottom)+0.25rem)] sm:static sm:mt-6 sm:border-0 sm:bg-transparent sm:p-0 sm:pb-0"),
							Div(Class("mx-auto grid max-w-7xl grid-cols-[minmax(0,1fr)_auto] items-center gap-3 sm:grid-cols-1 sm:rounded-xl sm:bg-cream sm:p-4"),
								Div(
									P(Class("text-sm font-bold text-ink"), Text("Precio estimado · "), Span(Attr("data-price-estimate-target", "type"), Text(stayType))),
									P(Class("mt-0.5 text-xl font-black text-ink sm:mt-1 sm:text-2xl"), Attr("aria-live", "polite"), Attr("aria-atomic", "true"),
										Span(Attr("data-price-estimate-target", "amount"), Text(estimatedLabel)),
										Text(" "),
										Span(Class("text-xs font-bold sm:text-sm"), Attr("data-price-estimate-target", "unit"), Text(priceUnit)),
									),
									P(Class("mt-1 text-xs leading-5 text-muted"), Text("El contacto confirma el precio final por WhatsApp.")),
								),
								Button(Type("submit"), Class("inline-flex min-h-12 w-full max-w-44 items-center justify-center rounded-lg bg-brand-dark px-3 py-2 text-center text-sm font-extrabold text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark sm:max-w-none sm:px-5 sm:py-3 sm:text-base"),
									Span(Class("sm:hidden"), Text("Solicitar por WhatsApp")),
									Span(Class("hidden sm:inline"), Text("Enviar solicitud y abrir WhatsApp")),
								),
							),
						),
						P(Class("text-center text-xs leading-5 text-muted"), Text("Aún no se realiza ningún pago. El pago se acuerda directamente con el contacto de la cabaña.")),
					),
				}),
			),
		),
	)
}

func DetailItem(title, value, detail string) Node {
	return Div(Class("rounded-xl border border-line bg-cream p-4"),
		P(Class("text-xs font-extrabold uppercase tracking-wider text-muted"), Text(title)),
		P(Class("mt-1 text-lg font-extrabold text-ink"), Text(value)),
		P(Class("mt-1 text-sm text-muted"), Text(detail)),
	)
}
