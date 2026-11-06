package ownerportal

import (
	"travelando/internal/system/components"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ownerProfileEl(owner models.Owner, state *ownerProfileState) Node {
	value := func(key, fallback string) string {
		if state != nil && state.Values != nil {
			return state.Values.Get(key)
		}
		return fallback
	}
	fieldError := func(key string) string {
		if state == nil {
			return ""
		}
		return state.Errors.ErrorStringFor(key)
	}
	return Div(ID("owner-profile"), Class("mx-auto w-full max-w-3xl px-4 py-4 sm:px-6 sm:py-6"),
		components.PageHeading("Perfil del propietario", "Actualiza tus datos de contacto y acceso."),
		Form(Method("post"), Action("/owner/profile"), Class("space-y-4"),
			Div(Class("rounded-2xl border border-line bg-white p-4 sm:p-6"),
				H2(Class("text-xl font-extrabold text-ink"), Text("Datos personales")),
				P(Class("mt-1 text-sm text-muted"), Text("Estos datos se usan para identificarte y contactar a tus huéspedes.")),
				Div(Class("mt-4 grid gap-4 sm:grid-cols-2"),
					components.TextFieldWithError("owner-profile-name", "name", "Nombre", value("name", owner.Name), "Nombre completo", "text", true, fieldError("name")),
					components.TextFieldWithError("owner-profile-cedula", "cedula", "Cédula", value("cedula", owner.Cedula), "Número de cédula", "text", true, fieldError("cedula")),
					Div(Class("sm:col-span-2"), components.TextFieldWithError("owner-profile-whatsapp", "whatsapp", "WhatsApp", value("whatsapp", owner.WhatsApp), "Incluye el código de país", "tel", true, fieldError("whatsapp"))),
				),
			),
			Div(Class("rounded-2xl border border-line bg-white p-4 sm:p-6"),
				H2(Class("text-xl font-extrabold text-ink"), Text("Cambiar PIN")),
				P(Class("mt-1 text-sm text-muted"), Text("Deja estos campos vacíos si no deseas cambiar tu PIN.")),
				Div(Class("mt-4 grid gap-4 sm:grid-cols-2"),
					ownerProfilePINField("owner-profile-current-pin", "current_pin", "PIN actual", "current-password", fieldError("current_pin")),
					ownerProfilePINField("owner-profile-new-pin", "new_pin", "Nuevo PIN", "new-password", fieldError("new_pin")),
					Div(Class("sm:col-span-2"), ownerProfilePINField("owner-profile-confirm-pin", "confirm_pin", "Confirmar PIN", "new-password", fieldError("confirm_pin"))),
				),
			),
			components.FormActions(
				Button(Type("submit"), Class(components.FormPrimaryClass), Text("Guardar cambios")),
			),
		),
	)
}

func ownerProfilePINField(id, name, label, autocomplete, message string) Node {
	return Div(Class("min-w-0"),
		components.FieldLabel(id, label),
		Input(ID(id), Name(name), Type("password"), InputMode("numeric"), Pattern("[0-9]{6}"), MinLength("6"), MaxLength("6"), AutoComplete(autocomplete), If(message != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", id+"-error")}), Class("w-full rounded-xl border border-line bg-white px-4 py-3 text-base text-ink focus:border-brand focus:ring-2 focus:ring-brand/20")),
		components.FieldError(id, message),
	)
}
