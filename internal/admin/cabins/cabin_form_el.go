package cabins

import (
	"net/url"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	lucide "github.com/eduardolat/gomponents-lucide"

	"go.leapkit.dev/core/form/validate"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func cabinSharedFormEl(cabin models.Cabin, blocks []models.CabinBlock, bookings []models.Inquiry, owners []models.Owner, state *cabinFormState, editing, ownerMode bool, ownerPage int) Node {
	title := "Nueva cabaña"
	if editing {
		title = "Editar cabaña"
	}
	formAction := "/admin/cabins"
	formTarget := "#cabin-page"
	if ownerMode {
		formAction = "/owner/cabins"
		formTarget = "#owner-content"
	}
	if editing {
		formAction += "/" + strconv.FormatInt(cabin.ID, 10)
	}
	if ownerMode {
		formAction += "?page=" + strconv.Itoa(max(1, ownerPage))
	}
	fieldErrors := validate.Errors(nil)
	values := url.Values(nil)
	blockErrors := validate.Errors(nil)
	blockValues := url.Values(nil)
	if state != nil {
		fieldErrors, values = state.Errors, state.Values
		blockErrors, blockValues = state.BlockErrors, state.BlockValues
	}
	fieldValue := func(key, fallback string) string {
		if values != nil {
			return values.Get(key)
		}
		return fallback
	}
	ownerIDValue := fieldValue("owner_id", strconv.FormatInt(cabin.OwnerID, 10))
	blockValue := func(key string) string {
		if blockValues != nil {
			return blockValues.Get(key)
		}
		return ""
	}
	descriptionClass := "w-full bg-white px-4 py-3 text-base leading-6 text-ink focus:ring-2 focus:ring-brand/20"
	if fieldErrors.Has("description") {
		descriptionClass += " border-red-300 focus:ring-red-200"
	}
	return Div(ID("cabin-page"),
		If(editing, Class("mx-auto w-full max-w-5xl px-5 py-8 sm:px-8 sm:py-12")),
		If(!editing, Class("mx-auto w-full max-w-5xl p-4 sm:p-7")),
		If(editing && !ownerMode, A(Href("/admin/cabins"), Class("inline-flex items-center gap-2 text-sm font-bold text-brand-dark hover:underline"), lucide.ArrowLeft(Class("size-4"), Attr("aria-hidden", "true")), Text("Volver a cabañas"))),
		If(!editing && !ownerMode,
			Div(Class("flex items-start justify-between gap-4"),
				components.PageHeading(title, "Datos y tarifas de esta cabaña. Precios en pesos colombianos (COP)."),
				Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
			),
		),
		If(ownerMode, Div(Class("mb-4 flex items-start justify-between gap-4"),
			components.PageHeading(title, "Datos de la cabaña y tarifas en pesos colombianos (COP)."),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		)),
		If(editing && !ownerMode, Div(Class("mt-6"), components.PageHeading(title, "Datos de la cabaña y tarifas en pesos colombianos (COP)."))),
		Form(Method("post"), Action(formAction), Attr("hx-post", formAction), Attr("hx-target", formTarget), Attr("hx-swap", "outerHTML"), If(editing && !ownerMode, Class("rounded-2xl border border-line bg-white p-5 sm:p-7")),
			Div(Class("space-y-7"),
				If(!ownerMode, Div(Class("grid gap-x-6 gap-y-4 rounded-2xl border border-line p-4 sm:grid-cols-2 sm:p-6"),
					Div(Class("sm:col-span-2"), H2(Class("text-base font-extrabold text-ink"), Text("Dueño de la cabaña"))),
					Div(Class("min-w-0 sm:col-span-2"),
						components.FieldLabel("owner_id", "Dueño"),
						Select(ID("owner_id"), Name("owner_id"), If(fieldErrors.Has("owner_id"), Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "owner_id-error")}), Class("min-h-11 w-full rounded-xl border border-line bg-white px-4 py-2.5 text-base text-ink focus:border-brand focus:ring-2 focus:ring-brand/20"),
							Option(Value(""), If(ownerIDValue == "", Selected()), Text("Elige un dueño")),
							Map(owners, func(owner models.Owner) Node {
								id := strconv.FormatInt(owner.ID, 10)
								label := owner.Name + " · " + owner.Cedula + " · " + owner.WhatsApp
								if !owner.Active {
									label += " · Inactivo"
								}
								return Option(Value(id), If(ownerIDValue == id, Selected()), Text(label))
							}),
						),
						P(Class("mt-1 text-xs text-muted"), Text("El nombre y WhatsApp de contacto vienen del dueño.")),
						components.FieldError("owner_id", fieldErrors.ErrorStringFor("owner_id")),
					),
				)),
				Div(Class("grid gap-x-6 gap-y-4 rounded-2xl border border-line p-4 sm:grid-cols-2 sm:p-6"),
					Div(Class("sm:col-span-2"), H2(Class("text-base font-extrabold text-ink"), Text("Datos de la cabaña"))),
					components.TextFieldWithError("name", "name", "Nombre de la cabaña", fieldValue("name", cabin.Name), "Ej. Refugio del río", "text", true, fieldErrors.ErrorStringFor("name")),
					components.NumberFieldWithError("included_guests", "included_guests", "Personas incluidas en el precio", fieldValue("included_guests", strconv.Itoa(cabin.IncludedGuests)), "1", "", fieldErrors.ErrorStringFor("included_guests")),
					Div(Class("min-w-0 sm:col-span-2"),
						components.FieldLabel("description", "Descripción"),
						Textarea(ID("description"), Name("description"), Rows("4"), MaxLength("2000"), Placeholder("Describe lo que hace especial a este lugar"), If(fieldErrors.Has("description"), Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "description-error")}), Class(descriptionClass), Text(fieldValue("description", cabin.Description))),
						components.FieldError("description", fieldErrors.ErrorStringFor("description")),
					),
					components.TextFieldWithError("location", "location", "Ubicación", fieldValue("location", cabin.Location), "Vereda, municipio", "text", true, fieldErrors.ErrorStringFor("location")),
					components.TextFieldWithError("location_url", "location_url", "Enlace del mapa (opcional)", fieldValue("location_url", cabin.LocationURL), "https://maps.google.com/...", "url", false, fieldErrors.ErrorStringFor("location_url")),
					Div(Class("min-w-0 sm:col-span-2"), components.TextFieldWithError("cover_image", "cover_image", "URL de foto de portada (opcional)", fieldValue("cover_image", cabin.CoverImage), "https://...", "url", false, fieldErrors.ErrorStringFor("cover_image"))),
					If(editing, Label(Class("flex min-h-11 items-center gap-2 text-sm font-semibold text-ink"),
						Input(Type("checkbox"), Name("active"), Value("on"), If(cabin.Active, Checked()), Class("size-4 accent-brand")),
						Text("Publicada en el buscador"),
					)),
				),
				Div(Class("rounded-2xl border border-line p-4 sm:p-6"),
					H2(Class("text-base font-extrabold text-ink"), Text("Tarifas y horarios")),
					P(Class("mt-1 text-sm text-muted"), Text("El cargo por huésped adicional se cobra una sola vez por reserva.")),
					Div(Class("mt-4 max-w-sm"), components.NumberFieldWithError("extra_guest_price", "extra_guest_price", "Cargo por huésped adicional (COP)", fieldValue("extra_guest_price", strconv.FormatInt(cabin.ExtraGuestPrice, 10)), "0", "1000000000000", fieldErrors.ErrorStringFor("extra_guest_price"))),
					Div(Class("mt-4 grid gap-6 lg:grid-cols-2"),
						Div(Class("grid gap-x-4 gap-y-4 sm:grid-cols-2"),
							H3(Class("sm:col-span-2 text-sm font-bold text-muted"), Text("Pasadía")),
							components.NumberFieldWithError("day_price", "day_price", "Precio de pasadía (COP)", fieldValue("day_price", strconv.FormatInt(cabin.DayPrice, 10)), "0", "1000000000000", fieldErrors.ErrorStringFor("day_price")),
							components.TimeFieldWithError("day_start", "day_start", "Hora de inicio", fieldValue("day_start", cabin.DayStart), fieldErrors.ErrorStringFor("day_start")),
							components.TimeFieldWithError("day_end", "day_end", "Hora de fin", fieldValue("day_end", cabin.DayEnd), fieldErrors.ErrorStringFor("day_end")),
						),
						Div(Class("grid gap-x-4 gap-y-4 sm:grid-cols-2"),
							H3(Class("sm:col-span-2 text-sm font-bold text-muted"), Text("Estadía")),
							components.NumberFieldWithError("night_price", "night_price", "Precio por noche (COP)", fieldValue("night_price", strconv.FormatInt(cabin.NightPrice, 10)), "0", "1000000000000", fieldErrors.ErrorStringFor("night_price")),
							components.TimeFieldWithError("check_in", "check_in", "Hora de entrada", fieldValue("check_in", cabin.CheckIn), fieldErrors.ErrorStringFor("check_in")),
							components.TimeFieldWithError("check_out", "check_out", "Hora de salida", fieldValue("check_out", cabin.CheckOut), fieldErrors.ErrorStringFor("check_out")),
						),
					),
				),
			),
			Div(Class("mt-6 flex flex-col-reverse gap-3 border-t border-line pt-5 sm:flex-row sm:flex-wrap"),
				Button(Type("submit"), Class("inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white transition hover:brightness-90 sm:w-auto"), lucide.Save(Class("size-4"), Attr("aria-hidden", "true")), Text("Guardar cabaña")),
				If(!editing && !ownerMode, Button(Type("button"), Attr("data-action", "click->modal#close"), Class("inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg border border-line px-3 py-2 font-bold text-muted hover:text-ink sm:w-auto"), Text("Cancelar"))),
				If(ownerMode, Button(Type("button"), Attr("data-action", "click->modal#close"), Class("inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg border border-line px-4 py-2 font-bold text-muted hover:text-ink sm:w-auto"), Text("Cancelar"))),
				If(editing && !ownerMode, A(Href("/admin/cabins"), Class("inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg border border-line px-4 py-2 font-bold text-muted hover:text-ink sm:w-auto"), Text("Cancelar"))),
			),
		),
		If(editing && !ownerMode, Div(Class("mt-6 rounded-2xl border border-line bg-white p-5 sm:p-7"),
			H2(Class("text-lg font-extrabold text-ink"), Text("Calendario de disponibilidad")),
			If(len(bookings) == 0 && len(blocks) == 0, P(Class("mt-2 text-sm text-muted"), Text("No hay fechas ocupadas."))),
			If(len(bookings) > 0, Div(Class("mt-4"),
				H3(Class("text-sm font-extrabold text-ink"), Text("Reservas confirmadas")),
				Div(Class("mt-2 divide-y divide-line"), Map(bookings, func(inquiry models.Inquiry) Node {
					return Div(Class("flex flex-wrap items-center justify-between gap-2 py-3 text-sm"),
						P(Class("font-bold text-ink"), Text(helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo)), Span(Class("ml-2 font-normal text-muted"), Text(inquiry.GuestName))),
						Span(Class("text-muted"), Text(helpers.StayTypeLabel(inquiry.StayType)+" · "+strconv.Itoa(inquiry.Guests)+" huéspedes")),
					)
				})),
			)),
			H3(Class("mt-5 text-sm font-extrabold text-ink"), Text("Bloquear fechas manualmente")),
			P(Class("mt-1 text-sm text-muted"), Text("Usa la salida como fecha final exclusiva, igual que en una estadía.")),
			Form(Method("post"), Action("/admin/cabins/"+strconv.FormatInt(cabin.ID, 10)+"/blocks"), Attr("hx-post", "/admin/cabins/"+strconv.FormatInt(cabin.ID, 10)+"/blocks"), Attr("hx-target", "#cabin-page"), Attr("hx-swap", "outerHTML"), Class("mt-4 grid gap-4 sm:grid-cols-[1fr_1fr_1.2fr_auto] sm:items-end"),
				components.TextFieldWithError("block-from", "date_from", "Desde", blockValue("date_from"), "", "date", true, blockErrors.ErrorStringFor("date_from")),
				components.TextFieldWithError("block-to", "date_to", "Hasta (sin incluir)", blockValue("date_to"), "", "date", true, blockErrors.ErrorStringFor("date_to")),
				components.TextFieldWithError("block-reason", "reason", "Motivo", blockValue("reason"), "Uso propio", "text", false, blockErrors.ErrorStringFor("reason")),
				Button(Type("submit"), Class("inline-flex min-h-11 items-center gap-2 rounded-lg border border-line px-4 py-2.5 font-bold text-ink hover:border-brand"), lucide.CalendarPlus(Class("size-4"), Attr("aria-hidden", "true")), Text("Bloquear")),
			),
			If(len(blocks) == 0, P(Class("mt-5 text-sm text-muted"), Text("No hay bloqueos manuales."))),
			Div(Class("mt-3 divide-y divide-line"), Map(blocks, func(block models.CabinBlock) Node {
				return Div(Class("flex flex-wrap items-center justify-between gap-3 py-3"),
					P(Class("text-sm font-semibold text-ink"), Text(block.DateFrom+" – "+block.DateTo), If(block.Source == "external", Text(" · Reserva externa")), If(block.Source != "external" && block.Reason != "", Text(" · "+block.Reason))),
					Form(Method("post"), Action("/admin/cabins/"+strconv.FormatInt(cabin.ID, 10)+"/blocks/"+strconv.FormatInt(block.ID, 10)+"/delete"), Attr("hx-post", "/admin/cabins/"+strconv.FormatInt(cabin.ID, 10)+"/blocks/"+strconv.FormatInt(block.ID, 10)+"/delete"), Attr("hx-target", "#cabin-page"), Attr("hx-swap", "outerHTML"),
						Button(Type("submit"), Class("inline-flex items-center gap-2 text-sm font-bold text-brand-dark hover:underline"), lucide.Trash2(Class("size-4"), Attr("aria-hidden", "true")), Text("Quitar bloqueo")),
					),
				)
			})),
		)),
	)
}
