package cabins

import (
	"net/url"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/models"

	lucide "github.com/eduardolat/gomponents-lucide"

	"go.leapkit.dev/core/form/validate"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func cabinSharedFormEl(cabin models.Cabin, blocks []models.CabinBlock, bookings []models.Inquiry, owners []models.Owner, amenityTypes []models.AmenityType, state *cabinFormState, editing, ownerMode bool, ownerPage int) Node {
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
	if state != nil {
		fieldErrors, values = state.Errors, state.Values
		if len(state.AmenityTypes) > 0 {
			amenityTypes = state.AmenityTypes
		}
	}
	fieldValue := func(key, fallback string) string {
		if values != nil {
			return values.Get(key)
		}
		return fallback
	}
	fieldChecked := func(key string, fallback bool) bool {
		if values != nil {
			return values.Get(key) == "on"
		}
		return fallback
	}
	ownerIDValue := fieldValue("owner_id", strconv.FormatInt(cabin.OwnerID, 10))
	ownerOptions := make([]components.SearchableOption, len(owners))
	for i, owner := range owners {
		label := owner.Name + " · " + owner.Cedula + " · " + owner.WhatsApp
		if !owner.Active {
			label += " · Inactivo"
		}
		ownerOptions[i] = components.SearchableOption{Value: strconv.FormatInt(owner.ID, 10), Label: label}
	}
	descriptionClass := "w-full bg-white px-4 py-3 text-base leading-6 text-ink focus:ring-2 focus:ring-brand/20"
	if fieldErrors.Has("description") {
		descriptionClass += " border-red-300 focus:ring-red-200"
	}
	return Div(ID("cabin-page"),
		Class("w-full p-4 sm:p-6"),
		Div(Class("flex items-start justify-between gap-4"),
			components.PageHeading(title, "Datos y tarifas de esta cabaña. Precios en pesos colombianos (COP)."),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		),
		Form(ID("cabin-data"), Method("post"), Action(formAction), Attr("hx-post", formAction), Attr("hx-target", formTarget), Attr("hx-swap", "outerHTML"),
			Div(Class("space-y-7"),
				If(!ownerMode, Div(Class("grid gap-x-6 gap-y-4 rounded-2xl border border-line p-4 sm:grid-cols-2 sm:p-6"),
					Div(Class("sm:col-span-2"), H2(Class("text-base font-extrabold text-ink"), Text("Dueño de la cabaña"))),
					Div(Class("min-w-0 sm:col-span-2"),
						components.SearchableSelect("owner_id", "owner_id", "Dueño", "Busca un dueño", ownerIDValue, ownerOptions, false, fieldErrors.ErrorStringFor("owner_id")),
						P(Class("mt-1 text-xs text-muted"), Text("El nombre y WhatsApp de contacto vienen del dueño.")),
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
					If(editing, Label(Class("flex min-h-11 items-center gap-2 text-sm font-semibold text-ink"),
						Input(Type("checkbox"), Name("active"), Value("on"), If(cabin.Active, Checked()), Class("size-4 accent-brand")),
						Text("Publicada en el buscador"),
					)),
				),
				amenityFieldsEl(cabin, amenityTypes, values, fieldErrors, ownerMode, false, nil),
				Div(Class("rounded-2xl border border-line p-4 sm:p-6"),
					H2(Class("text-base font-extrabold text-ink"), Text("Tarifas y horarios")),
					P(Class("mt-1 text-sm text-muted"), Text("El cargo por huésped adicional se cobra una sola vez por reserva.")),
					Div(Class("mt-5 max-w-sm"),
						components.NumberFieldWithError("cleaning_hours", "cleaning_hours", "Horas de aseo después de cada reserva", fieldValue("cleaning_hours", strconv.Itoa(cabin.CleaningHours)), "0", "24", fieldErrors.ErrorStringFor("cleaning_hours")),
						P(Class("mt-1 text-xs text-muted"), Text("La cabaña vuelve a estar lista al terminar este tiempo.")),
					),
					Div(Class("mt-5 grid items-start gap-5 lg:grid-cols-[minmax(15rem,1fr)_minmax(0,2fr)]"),
						components.NumberFieldWithError("extra_guest_price", "extra_guest_price", "Cargo por huésped adicional (COP)", fieldValue("extra_guest_price", strconv.FormatInt(cabin.ExtraGuestPrice, 10)), "0", "1000000000000", fieldErrors.ErrorStringFor("extra_guest_price")),
						Div(Class("space-y-1"),
							P(Class("text-sm font-bold text-ink"), Text("Modalidades que acepta esta cabaña")),
							Div(Class("flex flex-wrap gap-x-5 gap-y-1"),
								Label(Class("inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-ink"),
									Input(Type("checkbox"), Name("accepts_day_use"), Value("on"), If(fieldChecked("accepts_day_use", cabin.AcceptsDayUse), Checked()), Class("size-4 accent-brand")),
									Text("Acepta pasadías"),
								),
								Label(Class("inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-ink"),
									Input(Type("checkbox"), Name("accepts_overnight"), Value("on"), If(fieldChecked("accepts_overnight", cabin.AcceptsOvernight), Checked()), Class("size-4 accent-brand")),
									Text("Acepta estadías"),
								),
							),
							P(Class("text-xs text-muted"), Text("Si desmarcas ambas, la cabaña no aparecerá en búsquedas ni recibirá solicitudes nuevas.")),
						),
					),
					Div(Class("mt-5 grid gap-4 lg:grid-cols-2"),
						Div(Class("rounded-xl bg-cream/50 p-4 sm:p-5"),
							H3(Class("mb-4 text-sm font-extrabold text-ink"), Text("Pasadía")),
							Div(Class("grid gap-x-4 gap-y-4 sm:grid-cols-2"),
								Div(Class("sm:col-span-2"), components.NumberFieldWithError("day_price", "day_price", "Precio de pasadía (COP)", fieldValue("day_price", strconv.FormatInt(cabin.DayPrice, 10)), "0", "1000000000000", fieldErrors.ErrorStringFor("day_price"))),
								components.TimeFieldWithError("day_start", "day_start", "Hora de inicio", fieldValue("day_start", cabin.DayStart), fieldErrors.ErrorStringFor("day_start")),
								components.TimeFieldWithError("day_end", "day_end", "Hora de fin", fieldValue("day_end", cabin.DayEnd), fieldErrors.ErrorStringFor("day_end")),
							),
						),
						Div(Class("rounded-xl bg-cream/50 p-4 sm:p-5"),
							H3(Class("mb-4 text-sm font-extrabold text-ink"), Text("Estadía")),
							Div(Class("grid gap-x-4 gap-y-4 sm:grid-cols-2"),
								Div(Class("sm:col-span-2"), components.NumberFieldWithError("night_price", "night_price", "Precio por noche (COP)", fieldValue("night_price", strconv.FormatInt(cabin.NightPrice, 10)), "0", "1000000000000", fieldErrors.ErrorStringFor("night_price"))),
								components.TimeFieldWithError("check_in", "check_in", "Hora de entrada", fieldValue("check_in", cabin.CheckIn), fieldErrors.ErrorStringFor("check_in")),
								components.TimeFieldWithError("check_out", "check_out", "Hora de salida", fieldValue("check_out", cabin.CheckOut), fieldErrors.ErrorStringFor("check_out")),
							),
						),
					),
				),
			),
			Div(Class("mt-6 flex flex-col-reverse gap-3 border-t border-line pt-5 sm:flex-row sm:flex-wrap"),
				Button(Type("submit"), Class("inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white transition hover:brightness-90 sm:w-auto"), lucide.Save(Class("size-4"), Attr("aria-hidden", "true")), Text("Guardar cabaña")),
				Button(Type("button"), Attr("data-action", "click->modal#close"), Class("inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg border border-line px-4 py-2 font-bold text-muted hover:text-ink sm:w-auto"), Text("Cancelar")),
			),
		),
	)
}

func amenityFieldsEl(cabin models.Cabin, types []models.AmenityType, values url.Values, fieldErrors validate.Errors, ownerMode, oob bool, createState *components.AmenityManagerState) Node {
	manageURL := "/admin/amenities"
	createURL := "/admin/cabin-amenities/create"
	if ownerMode {
		manageURL = "/owner/amenities"
		createURL = "/owner/cabin-amenities/create"
	}
	newName, newValueType, message := "", models.AmenityBoolean, ""
	var createErrors validate.Errors
	if createState != nil {
		newName, newValueType, message, createErrors = createState.Type.Name, createState.Type.ValueType, createState.Message, createState.Errors
	}
	return Div(ID("cabin-amenity-fields"), If(oob, Attr("hx-swap-oob", "outerHTML")), Class("rounded-2xl border border-line p-4 sm:p-6"),
		Div(Class("flex flex-wrap items-start justify-between gap-3"),
			Div(H2(Class("text-base font-extrabold text-ink"), Text("Amenidades")),
				P(Class("mt-1 text-sm text-muted"), Text("Indica cuáles tiene esta cabaña y sus cantidades.")),
			),
			A(Href(manageURL), Class("inline-flex min-h-10 items-center rounded-lg border border-line px-3 py-2 text-sm font-bold text-brand-dark hover:border-brand"), Text("Administrar catálogo")),
		),
		If(len(types) == 0, P(Class("mt-4 text-sm text-muted"), Text("No hay amenidades disponibles."))),
		If(len(types) > 0, Div(Class("mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3"), Map(types, func(amenityType models.AmenityType) Node {
			key := "amenity_" + strconv.FormatInt(amenityType.ID, 10)
			value, checked := cabinAmenityFormValue(cabin, amenityType, values)
			if amenityType.ValueType == models.AmenityBoolean {
				return Div(Class("rounded-xl border border-line p-3"),
					Label(Class("flex min-h-10 items-center gap-2 text-sm font-semibold text-ink"),
						Input(ID(key), Type("checkbox"), Name(key), Value("on"), If(checked, Checked()), If(fieldErrors.Has(key), Attr("aria-invalid", "true")), Class("size-4 accent-brand")),
						Text(amenityType.Name),
					),
					components.FieldError(key, fieldErrors.ErrorStringFor(key)),
				)
			}
			return Div(Class("rounded-xl border border-line p-3"),
				components.NumberFieldWithError(key, key, amenityType.Name, value, "0", "", fieldErrors.ErrorStringFor(key)),
			)
		}))),
		Div(Class("mt-5 grid gap-3 border-t border-line pt-4 sm:grid-cols-[minmax(0,1fr)_14rem_auto] sm:items-end"),
			Div(Class("min-w-0"),
				components.FieldLabel("new_amenity_name", "Nombre"),
				Input(ID("new_amenity_name"), Name("new_amenity_name"), Type("text"), Value(newName), MaxLength("80"), Placeholder("Ej. Piscina"), If(createErrors.Has("new_amenity_name"), Attr("aria-invalid", "true")), Class("mt-1 h-10 w-full bg-white px-3 text-sm text-ink")),
				components.FieldError("new_amenity_name", createErrors.ErrorStringFor("new_amenity_name")),
			),
			Div(Class("min-w-0"),
				components.FieldLabel("new_amenity_value_type", "Tipo de valor"),
				Select(ID("new_amenity_value_type"), Name("new_amenity_value_type"), If(createErrors.Has("new_amenity_value_type"), Attr("aria-invalid", "true")), Class("mt-1 h-10 w-full bg-white px-3 text-sm text-ink"),
					Option(Value(string(models.AmenityBoolean)), If(newValueType == models.AmenityBoolean, Selected()), Text("Sí / no")),
					Option(Value(string(models.AmenityQuantity)), If(newValueType == models.AmenityQuantity, Selected()), Text("Cantidad")),
				),
				components.FieldError("new_amenity_value_type", createErrors.ErrorStringFor("new_amenity_value_type")),
			),
			Button(Type("button"), Attr("hx-post", createURL), Attr("hx-target", "#cabin-amenity-fields"), Attr("hx-swap", "outerHTML"), Attr("hx-include", "#cabin-data"), Class("inline-flex min-h-10 items-center justify-center rounded-lg bg-brand-dark px-3 text-sm font-bold text-white hover:brightness-90"), Text("Agregar")),
			If(message != "", P(Attr("role", "status"), Class("text-sm font-semibold text-forest sm:col-span-3"), Text(message))),
		),
	)
}

func cabinAmenityFormValue(cabin models.Cabin, amenityType models.AmenityType, values url.Values) (string, bool) {
	key := "amenity_" + strconv.FormatInt(amenityType.ID, 10)
	if values != nil {
		submitted, exists := values[key]
		if amenityType.ValueType == models.AmenityBoolean {
			return "", exists && len(submitted) == 1 && submitted[0] == "on"
		}
		if exists && len(submitted) == 1 {
			return submitted[0], false
		}
		return "0", false
	}
	for _, amenity := range cabin.Amenities {
		if amenity.TypeID == amenityType.ID {
			if amenityType.ValueType == models.AmenityBoolean {
				return "", amenity.BooleanValue
			}
			return strconv.Itoa(amenity.QuantityValue), false
		}
	}
	return "0", false
}
