package auth

import (
	"travelando/internal/system/components"

	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ownerLoginEl(state *loginState) Node {
	cedula, cedulaError, pinError := "", "", ""
	if state != nil {
		cedula = state.Values.Get("cedula")
		cedulaError = state.Errors.ErrorStringFor("cedula")
		pinError = state.Errors.ErrorStringFor("pin")
	}
	return Div(ID("owner-login-form"), Class("mx-auto w-full max-w-md px-5 py-8 sm:px-8"),
		Div(Class("rounded-2xl border border-line bg-white p-6 sm:p-8"),
			components.PageHeading("Acceso de dueños", "Ingresa con tu cédula y PIN para administrar tus cabañas."),
			Form(Method("post"), Action("/owner/login"), hx.Post("/owner/login"), hx.Target("#owner-login-form"), hx.Swap("outerHTML"), Class("mt-4 space-y-4"),
				components.TextFieldWithError("owner-cedula", "cedula", "Cédula", cedula, "Número de cédula", "text", true, cedulaError),
				Div(Class("min-w-0"),
					components.FieldLabel("owner-pin", "PIN de seis dígitos"),
					Input(ID("owner-pin"), Name("pin"), Type("password"), InputMode("numeric"), Pattern("[0-9]{6}"), MinLength("6"), MaxLength("6"), AutoComplete("current-password"), Required(), If(pinError != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "owner-pin-error")}), Class("w-full rounded-xl border border-line bg-white px-4 py-3 text-base text-ink focus:border-brand focus:ring-2 focus:ring-brand/20")),
					components.FieldError("owner-pin", pinError),
				),
				Button(Type("submit"), Class("inline-flex min-h-11 w-full items-center justify-center rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white hover:brightness-90"), Text("Entrar")),
			),
		),
	)
}
