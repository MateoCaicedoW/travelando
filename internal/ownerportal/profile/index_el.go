package profile

import (
	"travelando/internal/system/components"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func IndexPageEl(owner models.Owner, state *ownerProfileState) Node {
	return Div(ID("owner-profile"), Class("mx-auto w-full max-w-3xl px-4 py-4 sm:px-6 sm:py-6"),
		components.PageHeading("Perfil del propietario", "Actualiza tus datos de contacto y acceso."),
		Form(Method("post"), Action("/owner/profile"), Class("space-y-4"),
			Div(Class("rounded-2xl border border-line bg-white p-4 sm:p-6"),
				H2(Class("text-xl font-extrabold text-ink"), Text("Datos personales")),
				P(Class("mt-1 text-sm text-muted"), Text("Estos datos se usan para identificarte y contactar a tus huéspedes.")),
				Div(Class("mt-4 grid gap-4 sm:grid-cols-2"),
					components.TextFieldWithError("owner-profile-name", "name", "Nombre", value("name", owner.Name, state), "Nombre completo", "text", true, fieldError("name", state)),
					components.TextFieldWithError("owner-profile-cedula", "cedula", "Cédula", value("cedula", owner.Cedula, state), "Número de cédula", "text", true, fieldError("cedula", state)),
					Div(Class("sm:col-span-2"), components.TextFieldWithError("owner-profile-whatsapp", "whatsapp", "WhatsApp", value("whatsapp", owner.WhatsApp, state), "Incluye el código de país", "tel", true, fieldError("whatsapp", state))),
				),
			),
			Div(Class("rounded-2xl border border-line bg-white p-4 sm:p-6"),
				H2(Class("text-xl font-extrabold text-ink"), Text("Cambiar PIN")),
				P(Class("mt-1 text-sm text-muted"), Text("Deja estos campos vacíos si no deseas cambiar tu PIN.")),
				Div(Class("mt-4 grid gap-4 sm:grid-cols-2"),
					PINField("owner-profile-current-pin", "current_pin", "PIN actual", "current-password", fieldError("current_pin", state)),
					PINField("owner-profile-new-pin", "new_pin", "Nuevo PIN", "new-password", fieldError("new_pin", state)),
					Div(Class("sm:col-span-2"), PINField("owner-profile-confirm-pin", "confirm_pin", "Confirmar PIN", "new-password", fieldError("confirm_pin", state))),
				),
			),
			components.FormActions(
				Button(Type("submit"), Class(components.FormPrimaryClass), Text("Guardar cambios")),
			),
		),
	)
}

func value(key, fallback string, state *ownerProfileState) string {
	if state != nil && state.Values != nil {
		return state.Values.Get(key)
	}
	return fallback
}

func fieldError(key string, state *ownerProfileState) string {
	if state == nil {
		return ""
	}
	return state.Errors.ErrorStringFor(key)
}
