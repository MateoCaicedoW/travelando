package owners

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

type ownerFormState struct {
	Errors validate.Errors
	Values url.Values
}

func ownerFormErrors(r *http.Request, requirePIN bool) validate.Errors {
	pinValidators := []validate.ValidatorFn{validPIN}
	if requirePIN {
		pinValidators = append([]validate.ValidatorFn{validate.Required("Ingresa un PIN de seis dígitos.")}, pinValidators...)
	}
	return form.Validate(r, validate.Fields(
		validate.Field("name", validate.Required("Ingresa el nombre del dueño."), validate.MaxLength(120, "Usa máximo 120 caracteres.")),
		validate.Field("cedula", validate.Required("Ingresa la cédula."), validCedula),
		validate.Field("whatsapp", validate.Required("Ingresa el celular del dueño."), validOwnerPhone),
		validate.Field("pin", pinValidators...),
	))
}

func validCedula(values []string) error {
	if missingOwnerField(values) {
		return nil
	}
	cedula := helpers.DigitsOnly(values[0])
	if len(cedula) < 5 || len(cedula) > 15 {
		return errors.New("Ingresa una cédula válida de 5 a 15 dígitos.")
	}
	return nil
}

func validOwnerPhone(values []string) error {
	if missingOwnerField(values) {
		return nil
	}
	phone := helpers.DigitsOnly(values[0])
	if len(phone) < 10 || len(phone) > 15 {
		return errors.New("Ingresa entre 10 y 15 dígitos, con código de país.")
	}
	return nil
}

func validPIN(values []string) error {
	if missingOwnerField(values) {
		return nil
	}
	pin := values[0]
	if len(pin) != 6 || strings.Trim(pin, "0123456789") != "" {
		return errors.New("El PIN debe tener exactamente seis dígitos.")
	}
	return nil
}

func ownerFromForm(r *http.Request, id int64) (models.Owner, string) {
	return models.Owner{ID: id, Name: r.FormValue("name"), Cedula: r.FormValue("cedula"), WhatsApp: r.FormValue("whatsapp")}, r.FormValue("pin")
}

func addOwnerFormError(state *ownerFormState, field, message string) {
	if state.Errors == nil {
		state.Errors = validate.Errors{}
	}
	state.Errors[field] = append(state.Errors[field], errors.New(message))
}

func ownerIDFromPath(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid owner ID")
	}
	return id, nil
}

func missingOwnerField(values []string) bool {
	return len(values) == 0 || strings.TrimSpace(values[0]) == ""
}
