package ownerportal

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	"go.leapkit.dev/core/server/session"
	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type loginState struct {
	Errors validate.Errors
	Values url.Values
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if ownerIDFromSession(r) > 0 {
		helpers.Redirect(w, r, "/owner/")
		return
	}
	renderLoginPage(w, r, nil, http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	state := &loginState{Errors: ownerLoginErrors(r), Values: r.Form}
	if state.Errors.HasAny() {
		renderLoginPage(w, r, state, http.StatusUnprocessableEntity)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owner, err := svc.AuthenticateOwner(r.FormValue("cedula"), r.FormValue("pin"))
	if errors.Is(err, services.ErrUnauthorized) {
		state.Errors["pin"] = append(state.Errors["pin"], errors.New("Cédula o PIN incorrecto. Si fallas cinco veces, espera 15 minutos."))
		renderLoginPage(w, r, state, http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("authenticating cabin owner", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	sess := session.FromCtx(r.Context())
	if sess == nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	sess.Values[ownerSessionKey] = ownerSessionValue(owner.ID, owner.SessionVersion)
	helpers.Redirect(w, r, "/owner/")
}

func ownerLoginErrors(r *http.Request) validate.Errors {
	return form.Validate(r, validate.Fields(
		validate.Field("cedula", validate.Required("Ingresa tu cédula."), loginCedula),
		validate.Field("pin", validate.Required("Ingresa tu PIN."), loginPIN),
	))
}

func loginCedula(values []string) error {
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return nil
	}
	cedula := helpers.DigitsOnly(values[0])
	if len(cedula) < 5 || len(cedula) > 15 {
		return errors.New("Ingresa una cédula válida.")
	}
	return nil
}

func loginPIN(values []string) error {
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return nil
	}
	if len(values[0]) != 6 || strings.Trim(values[0], "0123456789") != "" {
		return errors.New("El PIN debe tener seis dígitos.")
	}
	return nil
}

func renderLoginPage(w http.ResponseWriter, r *http.Request, state *loginState, status int) {
	if helpers.IsHTMX(r) {
		if status != http.StatusOK && status != http.StatusUnprocessableEntity {
			status = http.StatusUnprocessableEntity
		}
		w.WriteHeader(status)
		if err := ownerLoginEl(state).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(status)
	page := system.Layout{Title: "Acceso de dueños | Travelando", Portal: system.PortalBare, Yield: ownerLoginEl(state)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func ownerLoginEl(state *loginState) Node {
	cedula, cedulaError, pinError := "", "", ""
	if state != nil {
		cedula = state.Values.Get("cedula")
		cedulaError = state.Errors.ErrorStringFor("cedula")
		pinError = state.Errors.ErrorStringFor("pin")
	}
	return Div(ID("owner-login-form"), Class("mx-auto w-full max-w-md px-5 py-12 sm:px-8"),
		Div(Class("rounded-2xl border border-line bg-white p-6 sm:p-8"),
			components.PageHeading("Acceso de dueños", "Ingresa con tu cédula y PIN para administrar tus cabañas."),
			Form(Method("post"), Action("/owner/login"), Attr("hx-post", "/owner/login"), Attr("hx-target", "#owner-login-form"), Attr("hx-swap", "outerHTML"), Class("mt-6 space-y-4"),
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

func Logout(w http.ResponseWriter, r *http.Request) {
	if sess := session.FromCtx(r.Context()); sess != nil {
		delete(sess.Values, ownerSessionKey)
	}
	helpers.Redirect(w, r, "/owner/login")
}
