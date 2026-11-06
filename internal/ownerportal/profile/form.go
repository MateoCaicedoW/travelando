package profile

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"travelando/internal/system/helpers"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

type ownerProfileState struct {
	Errors validate.Errors
	Values url.Values
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
