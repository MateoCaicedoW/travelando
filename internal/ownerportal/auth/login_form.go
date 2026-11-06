package auth

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"travelando/internal/system/helpers"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

type loginState struct {
	Errors validate.Errors
	Values url.Values
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
