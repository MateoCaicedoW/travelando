package auth

import (
	"travelando/internal/system/components"

	"go.leapkit.dev/core/form/validate"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func loginPageEl(fieldErrors validate.Errors) Node {
	passwordClass := "w-full bg-white px-4 py-3 text-base text-ink outline-none focus:ring-2 focus:ring-brand/20"
	if fieldErrors.Has("password") {
		passwordClass += " border-red-300 focus:ring-red-200"
	}
	return Div(ID("login-panel"), Class("mx-auto w-full max-w-md px-5 py-8 sm:py-16"),
		Div(Class("rounded-2xl border border-line bg-white p-6 sm:p-9"),
			H1(Class("text-3xl font-extrabold tracking-tight text-ink"), Text("Tu espacio de gestión")),
			P(Class("mt-3 text-sm leading-6 text-muted"), Text("Ingresa para administrar cabañas, solicitudes y comisiones.")),
			Form(Class("mt-7 space-y-4"), Method("post"), Action("/admin/login"), hx.Post("/admin/login"), hx.Target("#login-panel"), hx.Swap("outerHTML"),
				Div(Class("space-y-2"),
					Label(For("password"), Class("text-[0.8125rem] font-semibold text-ink"), Text("Contraseña")),
					Input(ID("password"), Name("password"), Type("password"), AutoComplete("current-password"), Required(), AutoFocus(), If(fieldErrors.Has("password"), Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "password-error")}), Class(passwordClass)),
					components.FieldError("password", fieldErrors.ErrorStringFor("password")),
				),
				Button(Type("submit"), Class("inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-brand-dark px-4 py-3 font-extrabold text-white transition hover:brightness-90"), Text("Ingresar al panel")),
			),
		),
	)
}
