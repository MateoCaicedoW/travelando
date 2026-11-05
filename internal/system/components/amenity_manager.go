package components

import (
	"strconv"

	"go.leapkit.dev/core/form/validate"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type AmenityManagerState struct {
	Type    models.AmenityType
	Errors  validate.Errors
	Message string
}

func AmenityManager(title, description, basePath string, types []models.AmenityType, state *AmenityManagerState) Node {
	newType := models.AmenityType{ValueType: models.AmenityBoolean}
	message := ""
	if state != nil && state.Type.ID == 0 {
		newType = state.Type
	}
	if state != nil {
		message = state.Message
	}
	return Div(Class("mx-auto w-full max-w-5xl px-5 py-8 sm:px-8 sm:py-12"),
		PageHeading(title, description),
		If(message != "", P(Attr("role", "alert"), Class("mt-4 rounded-xl border border-red-200 bg-red-50 p-4 text-sm font-semibold text-red-800"), Text(message))),
		Form(Method("post"), Action(basePath+"/create"), Class("mt-6 grid gap-3 rounded-xl border border-line bg-white p-4 sm:grid-cols-[minmax(0,1fr)_14rem_auto] sm:items-end"),
			Div(Class("sm:col-span-3"), H2(Class("text-sm font-extrabold text-ink"), Text("Agregar amenidad"))),
			amenityTypeFields("new", newType, managerErrors(state, 0)),
			Button(Type("submit"), Class("inline-flex min-h-11 items-center justify-center rounded-lg bg-brand-dark px-4 py-2 text-sm font-extrabold text-white hover:brightness-90"), Text("Agregar")),
		),
		P(Class("mt-3 text-xs leading-5 text-muted"),
			Text("Cambiar entre sí/no y cantidad reinicia valores existentes. Eliminar una amenidad también eliminará los valores asignados en todas las cabañas."),
		),
		If(len(types) == 0, P(Class("mt-6 border-t border-line py-5 text-sm text-muted"), Text("Aún no hay amenidades personalizadas."))),
		If(len(types) > 0, Div(Attr("role", "region"), Attr("aria-label", "Tabla de amenidades"), TabIndex("0"), Class("mt-5 overflow-x-auto rounded-xl border border-line bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"),
			Table(Class("w-full min-w-[42rem] border-collapse text-left"),
				THead(Class("bg-cream text-xs font-bold text-muted"), Tr(
					Th(Attr("scope", "col"), Class("px-4 py-3"), Text("Amenidad")),
					Th(Attr("scope", "col"), Class("px-4 py-3"), Text("Tipo de valor")),
					Th(Attr("scope", "col"), Class("w-48 px-4 py-3 text-right"), Text("Acciones")),
				)),
				TBody(Map(types, func(amenityType models.AmenityType) Node {
					return amenityTypeRow(basePath, amenityType, state)
				})),
			),
		)),
	)
}

func amenityTypeFields(id string, amenityType models.AmenityType, errors validate.Errors) Node {
	nameID := "amenity-name-" + id
	typeID := "amenity-type-" + id
	return Group{
		Div(Class("min-w-0"),
			FieldLabel(nameID, "Nombre"),
			Input(ID(nameID), Name("name"), Type("text"), Value(amenityType.Name), Required(), MaxLength("80"), Placeholder("Ej. Piscina"), If(errors.Has("name"), Attr("aria-invalid", "true")), Class("mt-1 w-full bg-white px-4 py-3 text-base text-ink")),
			FieldError(nameID, errors.ErrorStringFor("name")),
		),
		Div(Class("min-w-0"),
			FieldLabel(typeID, "Tipo de valor"),
			Select(ID(typeID), Name("value_type"), Required(), If(errors.Has("value_type"), Attr("aria-invalid", "true")), Class("mt-1 w-full bg-white px-4 py-3 text-base text-ink"),
				Option(Value(string(models.AmenityBoolean)), If(amenityType.ValueType == models.AmenityBoolean, Selected()), Text("Sí / no")),
				Option(Value(string(models.AmenityQuantity)), If(amenityType.ValueType == models.AmenityQuantity, Selected()), Text("Cantidad")),
			),
			FieldError(typeID, errors.ErrorStringFor("value_type")),
		),
	}
}

func amenityTypeRow(basePath string, amenityType models.AmenityType, state *AmenityManagerState) Node {
	id := strconv.FormatInt(amenityType.ID, 10)
	nameID, valueTypeID, formID := "amenity-name-"+id, "amenity-type-"+id, "amenity-edit-"+id
	formType := amenityType
	var errors validate.Errors
	if state != nil && state.Type.ID == amenityType.ID {
		formType, errors = state.Type, state.Errors
	}
	return Tr(Class("border-t border-line align-middle"),
		Td(Class("px-4 py-3"),
			Form(ID(formID), Method("post"), Action(basePath+"/"+id)),
			Label(For(nameID), Class("sr-only"), Text("Nombre de amenidad")),
			Input(ID(nameID), Attr("form", formID), Name("name"), Type("text"), Value(formType.Name), Required(), MaxLength("80"), If(errors.Has("name"), Attr("aria-invalid", "true")), Class("h-10 w-full min-w-48 bg-white px-3 text-sm text-ink")),
			FieldError(nameID, errors.ErrorStringFor("name")),
		),
		Td(Class("px-4 py-3"),
			Label(For(valueTypeID), Class("sr-only"), Text("Tipo de valor")),
			Select(ID(valueTypeID), Attr("form", formID), Name("value_type"), Required(), If(errors.Has("value_type"), Attr("aria-invalid", "true")), Class("h-10 w-full min-w-40 bg-white px-3 text-sm text-ink"),
				Option(Value(string(models.AmenityBoolean)), If(formType.ValueType == models.AmenityBoolean, Selected()), Text("Sí / no")),
				Option(Value(string(models.AmenityQuantity)), If(formType.ValueType == models.AmenityQuantity, Selected()), Text("Cantidad")),
			),
			FieldError(valueTypeID, errors.ErrorStringFor("value_type")),
		),
		Td(Class("px-4 py-3"),
			Div(Class("flex items-center justify-end gap-2"),
				Button(Type("submit"), Attr("form", formID), Class("inline-flex min-h-10 items-center rounded-lg border border-line px-3 text-sm font-bold text-ink hover:border-brand"), Text("Guardar")),
				Form(Method("post"), Action(basePath+"/"+id+"/delete"), Attr("data-confirmation-message", "¿Eliminar amenidad y sus valores de todas las cabañas?"),
					Button(Type("submit"), Class("inline-flex min-h-10 items-center rounded-lg px-3 text-sm font-bold text-brand-dark hover:bg-cream"), Text("Eliminar")),
				),
			),
		),
	)
}

func managerErrors(state *AmenityManagerState, id int64) validate.Errors {
	if state != nil && state.Type.ID == id {
		return state.Errors
	}
	return nil
}
