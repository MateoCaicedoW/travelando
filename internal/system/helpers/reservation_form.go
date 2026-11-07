package helpers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"travelando/internal/system/models"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

// ReservationFormState holds the form state for creating a reservation.
type ReservationFormState struct {
	Errors validate.Errors
	Values url.Values
	Cabins []models.Cabin
}

// ReservationFormErrors validates a reservation creation request.
// maxDepositCOP is injected because helpers cannot import services
// (services imports helpers, which would be an import cycle).
func ReservationFormErrors(r *http.Request, maxDepositCOP int64) validate.Errors {
	return form.Validate(r, validate.Fields(
		validate.Field("cabin_id", validate.Required("Selecciona una cabaña."), ValidCabinID),
		validate.Field("guest_name", validate.Required("Ingresa el nombre del huésped."), validate.MaxLength(120, "Usa máximo 120 caracteres.")),
		validate.Field("guest_cedula", validate.Required("Ingresa la cédula del huésped."), validate.MinLength(5, "La cédula es muy corta."), validate.MaxLength(15, "Usa máximo 15 caracteres.")),
		validate.Field("guest_phone", validate.Required("Ingresa el teléfono del huésped."), validate.MinLength(8, "El teléfono es muy corto."), validate.MaxLength(15, "Usa máximo 15 caracteres.")),
		validate.Field("guest_email", validate.Required("Ingresa el correo del huésped."), validate.MaxLength(254, "Usa máximo 254 caracteres.")),
		validate.Field("stay_type", validate.Required("Selecciona el tipo de estadía."), ValidStayType),
		validate.Field("date_from", validate.Required("Selecciona la fecha inicial."), ValidDate),
		validate.Field("date_to", validate.Required("Selecciona la fecha final."), ValidDate),
		validate.Field("guests", validate.Required("Indica la cantidad de huéspedes."), IntegerRange(1, 999, "Ingresa una cantidad válida de huéspedes.")),
		validate.Field("deposit_cop", validate.Required("Ingresa el depósito en COP."), Integer64Range(0, maxDepositCOP, "Ingresa un depósito válido en COP.")),
	))
}

// ReservationFromForm parses and validates a reservation creation request.
func ReservationFromForm(r *http.Request, maxDepositCOP int64) (models.ReservationCreate, validate.Errors) {
	fieldErrors := ReservationFormErrors(r, maxDepositCOP)
	data := models.ReservationCreate{
		CabinID:     ParseInt64(r.FormValue("cabin_id")),
		GuestName:   strings.TrimSpace(r.FormValue("guest_name")),
		GuestCedula: strings.TrimSpace(r.FormValue("guest_cedula")),
		GuestPhone:  strings.TrimSpace(r.FormValue("guest_phone")),
		GuestEmail:  strings.TrimSpace(r.FormValue("guest_email")),
		StayType:    models.StayType(strings.TrimSpace(r.FormValue("stay_type"))),
		DateFrom:    strings.TrimSpace(r.FormValue("date_from")),
		DateTo:      strings.TrimSpace(r.FormValue("date_to")),
		Guests:      ParseInt(r.FormValue("guests")),
		DepositCOP:  ParseInt64(r.FormValue("deposit_cop")),
	}
	return data, fieldErrors
}

func ParseInt64(value string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return n
}

func ParseInt(value string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(value))
	return n
}

func ValidCabinID(values []string) error {
	if Missing(values) {
		return nil
	}
	if ParseInt64(values[0]) < 1 {
		return errors.New("Selecciona una cabaña válida.")
	}
	return nil
}

func ValidStayType(values []string) error {
	if Missing(values) {
		return nil
	}
	switch models.StayType(strings.TrimSpace(values[0])) {
	case models.StayDayUse, models.StayOvernight:
		return nil
	}
	return errors.New("Tipo de estadía no válido.")
}

func ValidDate(values []string) error {
	if Missing(values) {
		return nil
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(values[0])); err != nil {
		return errors.New("Fecha inválida.")
	}
	return nil
}

func PositiveInt(values []string) error {
	if Missing(values) {
		return nil
	}
	if ParseInt(values[0]) < 1 {
		return errors.New("La cantidad debe ser al menos 1.")
	}
	return nil
}

func Integer64Range(min, max int64, message string) validate.ValidatorFn {
	return func(values []string) error {
		if Missing(values) {
			return nil
		}
		value, err := strconv.ParseInt(strings.TrimSpace(values[0]), 10, 64)
		if err != nil || value < min || value > max {
			return errors.New(message)
		}
		return nil
	}
}

func Missing(values []string) bool {
	return len(values) == 0 || strings.TrimSpace(values[0]) == ""
}

func IntegerRange(min, max int, message string) validate.ValidatorFn {
	return func(values []string) error {
		if Missing(values) {
			return nil
		}
		value, err := strconv.Atoi(strings.TrimSpace(values[0]))
		if err != nil || value < min || value > max {
			return errors.New(message)
		}
		return nil
	}
}
