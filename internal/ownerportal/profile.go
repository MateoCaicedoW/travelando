package ownerportal

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	"go.leapkit.dev/core/server/session"
	"travelando/internal/system"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

type ownerProfileState struct {
	Errors validate.Errors
	Values url.Values
}

func Profile(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owner, err := svc.OwnerByID(ownerIDFromSession(r))
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading owner profile", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	renderOwnerProfile(w, r, *owner, nil, http.StatusOK)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	r.Form = r.PostForm
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	ownerID := ownerIDFromSession(r)
	current, err := svc.OwnerByID(ownerID)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading owner profile for update", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	state := &ownerProfileState{Errors: ownerProfileErrors(r), Values: r.PostForm}
	if state.Errors.HasAny() {
		renderOwnerProfile(w, r, *current, state, http.StatusUnprocessableEntity)
		return
	}

	newPIN := r.FormValue("new_pin")
	if newPIN != "" {
		authenticated, err := svc.AuthenticateOwner(current.Cedula, r.FormValue("current_pin"))
		if errors.Is(err, services.ErrUnauthorized) || err == nil && authenticated.ID != ownerID {
			addOwnerProfileError(state, "current_pin", "El PIN actual es incorrecto.")
			renderOwnerProfile(w, r, *current, state, http.StatusUnprocessableEntity)
			return
		}
		if err != nil {
			slog.Error("verifying owner PIN for profile update", "error", err)
			server.Error(w, err, http.StatusInternalServerError)
			return
		}
	}

	updated := models.Owner{ID: ownerID, Name: r.FormValue("name"), Cedula: r.FormValue("cedula"), WhatsApp: r.FormValue("whatsapp")}
	if err := svc.UpdateOwner(&updated, newPIN); err != nil {
		switch {
		case errors.Is(err, services.ErrConflict):
			addOwnerProfileError(state, "cedula", "La cédula o el WhatsApp ya están registrados.")
		case errors.Is(err, services.ErrInvalidInput):
			addOwnerProfileError(state, "new_pin", "Revisa los datos ingresados.")
		case errors.Is(err, services.ErrNotFound):
			http.NotFound(w, r)
			return
		default:
			slog.Error("updating owner profile", "error", err)
			server.Error(w, err, http.StatusInternalServerError)
			return
		}
		renderOwnerProfile(w, r, *current, state, http.StatusUnprocessableEntity)
		return
	}

	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Perfil actualizado."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	if updated.Cedula != current.Cedula || updated.WhatsApp != current.WhatsApp || newPIN != "" {
		if sess := session.FromCtx(r.Context()); sess != nil {
			delete(sess.Values, ownerSessionKey)
		}
		helpers.Redirect(w, r, "/owner/login")
		return
	}
	helpers.Redirect(w, r, "/owner/profile")
}

func ownerProfileErrors(r *http.Request) validate.Errors {
	fieldErrors := form.Validate(r, validate.Fields(
		validate.Field("name", validate.Required("Ingresa tu nombre."), validProfileName),
		validate.Field("cedula", validate.Required("Ingresa tu cédula."), validProfileCedula),
		validate.Field("whatsapp", validate.Required("Ingresa tu WhatsApp."), validProfileWhatsApp),
	))

	currentPIN, newPIN, confirmPIN := r.FormValue("current_pin"), r.FormValue("new_pin"), r.FormValue("confirm_pin")
	if currentPIN != "" || newPIN != "" || confirmPIN != "" {
		if currentPIN == "" {
			addOwnerProfileFormError(&fieldErrors, "current_pin", "Ingresa tu PIN actual para cambiarlo.")
		} else if !validSixDigitPIN(currentPIN) {
			addOwnerProfileFormError(&fieldErrors, "current_pin", "El PIN debe tener seis dígitos.")
		}
		if newPIN == "" {
			addOwnerProfileFormError(&fieldErrors, "new_pin", "Ingresa el nuevo PIN.")
		} else if !validSixDigitPIN(newPIN) {
			addOwnerProfileFormError(&fieldErrors, "new_pin", "El PIN debe tener seis dígitos.")
		}
		if confirmPIN == "" {
			addOwnerProfileFormError(&fieldErrors, "confirm_pin", "Confirma el nuevo PIN.")
		} else if newPIN != confirmPIN {
			addOwnerProfileFormError(&fieldErrors, "confirm_pin", "Los PIN no coinciden.")
		}
	}
	return fieldErrors
}

func validProfileName(values []string) error {
	if len(values) == 0 || values[0] == "" {
		return nil
	}
	name := strings.TrimSpace(values[0])
	if name == "" {
		return errors.New("Ingresa tu nombre.")
	}
	if utf8.RuneCountInString(name) > 120 {
		return errors.New("Usa máximo 120 caracteres.")
	}
	return nil
}

func validProfileCedula(values []string) error {
	if missingProfileValue(values) {
		return nil
	}
	if length := len(helpers.DigitsOnly(values[0])); length < 5 || length > 15 {
		return errors.New("Ingresa una cédula válida de 5 a 15 dígitos.")
	}
	return nil
}

func validProfileWhatsApp(values []string) error {
	if missingProfileValue(values) {
		return nil
	}
	if length := len(helpers.DigitsOnly(values[0])); length < 10 || length > 15 {
		return errors.New("Ingresa entre 10 y 15 dígitos, con código de país.")
	}
	return nil
}

func missingProfileValue(values []string) bool {
	return len(values) == 0 || strings.TrimSpace(values[0]) == ""
}

func validSixDigitPIN(pin string) bool {
	return len(pin) == 6 && strings.Trim(pin, "0123456789") == ""
}

func addOwnerProfileError(state *ownerProfileState, field, message string) {
	addOwnerProfileFormError(&state.Errors, field, message)
}

func addOwnerProfileFormError(fieldErrors *validate.Errors, field, message string) {
	if *fieldErrors == nil {
		*fieldErrors = validate.Errors{}
	}
	(*fieldErrors)[field] = append((*fieldErrors)[field], errors.New(message))
}

func renderOwnerProfile(w http.ResponseWriter, r *http.Request, owner models.Owner, state *ownerProfileState, status int) {
	w.WriteHeader(status)
	page := system.Layout{Title: "Perfil del propietario | Travelando", Portal: system.PortalOwner, Yield: ownerProfileEl(owner, state)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
