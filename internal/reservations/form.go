package reservations

import (
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"time"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

type formState struct {
	Values url.Values
	Errors validate.Errors
	Notice string
}

func updateFromForm(r *http.Request) (models.ReservationUpdate, validate.Errors) {
	stayType := r.FormValue("stay_type")
	dateFrom := r.FormValue("date_from")
	fieldErrors := form.Validate(r, validate.Fields(
		validate.Field("guest_name", validate.Required("Ingresa el nombre del huésped."), validate.MaxLength(120, "Usa máximo 120 caracteres.")),
		validate.Field("guest_phone", validate.Required("Ingresa el teléfono."), validPhone),
		validate.Field("guest_email", validate.Required("Ingresa el correo."), validEmail),
		validate.Field("stay_type", validate.Required("Selecciona el tipo de reserva."), validate.WithinOptions([]string{string(models.StayDayUse), string(models.StayOvernight)}, "Selecciona un tipo válido.")),
		validate.Field("date_from", validate.Required("Selecciona la fecha de llegada."), validDate("Selecciona una fecha válida.")),
		validate.Field("date_to", validate.Required("Selecciona la fecha de salida."), validDate("Selecciona una fecha válida."), validDateRange(stayType, dateFrom)),
		validate.Field("guests", validate.Required("Indica la cantidad de huéspedes."), positiveInteger("Ingresa una cantidad válida.")),
		validate.Field("final_total", validate.Required("Ingresa el total acordado."), totalAmount),
	))
	update := models.ReservationUpdate{
		GuestName: r.FormValue("guest_name"), GuestPhone: r.FormValue("guest_phone"), GuestEmail: r.FormValue("guest_email"),
		StayType: models.StayType(stayType), DateFrom: dateFrom, DateTo: r.FormValue("date_to"), UseFinalTotal: r.FormValue("use_final_total") == "on",
	}
	update.Guests, _ = strconv.Atoi(r.FormValue("guests"))
	update.FinalTotal, _ = strconv.ParseInt(r.FormValue("final_total"), 10, 64)
	return update, fieldErrors
}

func pendingUpdateFromForm(r *http.Request) (models.InquiryUpdate, validate.Errors) {
	stayType := r.FormValue("stay_type")
	dateFrom := r.FormValue("date_from")
	fieldErrors := form.Validate(r, validate.Fields(
		validate.Field("guest_name", validate.Required("Ingresa el nombre del huésped."), validate.MaxLength(120, "Usa máximo 120 caracteres.")),
		validate.Field("guest_phone", validate.Required("Ingresa el teléfono."), validPhone),
		validate.Field("guest_email", validate.Required("Ingresa el correo."), validEmail),
		validate.Field("stay_type", validate.Required("Selecciona el tipo de reserva."), validate.WithinOptions([]string{string(models.StayDayUse), string(models.StayOvernight)}, "Selecciona un tipo válido.")),
		validate.Field("date_from", validate.Required("Selecciona la fecha de llegada."), validDate("Selecciona una fecha válida.")),
		validate.Field("date_to", validate.Required("Selecciona la fecha de salida."), validDate("Selecciona una fecha válida."), validDateRange(stayType, dateFrom)),
		validate.Field("guests", validate.Required("Indica la cantidad de huéspedes."), positiveInteger("Ingresa una cantidad válida.")),
	))
	update := models.InquiryUpdate{
		GuestName: r.FormValue("guest_name"), GuestPhone: r.FormValue("guest_phone"), GuestEmail: r.FormValue("guest_email"),
		StayType: models.StayType(stayType), DateFrom: dateFrom, DateTo: r.FormValue("date_to"),
	}
	update.Guests, _ = strconv.Atoi(r.FormValue("guests"))
	return update, fieldErrors
}

func totalAmount(values []string) error {
	if len(values) == 0 || values[0] == "" {
		return nil
	}
	value, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || value < 0 || value > services.MaxAmountCOP {
		return errors.New("Ingresa un total válido entre 0 y 1.000.000.000.000 COP.")
	}
	return nil
}

func validPhone(values []string) error {
	if len(values) == 0 || values[0] == "" {
		return nil
	}
	if len(helpers.DigitsOnly(values[0])) < 8 || len(helpers.DigitsOnly(values[0])) > 15 {
		return errors.New("Ingresa un teléfono válido de 8 a 15 dígitos.")
	}
	return nil
}

func validEmail(values []string) error {
	if len(values) == 0 || values[0] == "" {
		return nil
	}
	address, err := mail.ParseAddress(values[0])
	if err != nil || address.Address != values[0] || len(values[0]) > 254 {
		return errors.New("Ingresa un correo válido.")
	}
	return nil
}

func validDate(message string) validate.ValidatorFn {
	return func(values []string) error {
		if len(values) == 0 || values[0] == "" {
			return nil
		}
		if _, err := time.Parse(time.DateOnly, values[0]); err != nil {
			return errors.New(message)
		}
		return nil
	}
}

func validDateRange(stayType, dateFrom string) validate.ValidatorFn {
	return func(values []string) error {
		if len(values) == 0 || values[0] == "" || dateFrom == "" {
			return nil
		}
		from, errFrom := time.Parse(time.DateOnly, dateFrom)
		to, errTo := time.Parse(time.DateOnly, values[0])
		if errFrom != nil || errTo != nil {
			return nil
		}
		if stayType == string(models.StayDayUse) && !helpers.SameDate(from, to) {
			return errors.New("La pasadía debe iniciar y terminar el mismo día.")
		}
		if stayType == string(models.StayOvernight) && (!to.After(from) || to.Sub(from).Hours()/24 > 365) {
			return errors.New("La salida debe ser posterior a la llegada y máximo a 365 días.")
		}
		return nil
	}
}

func positiveInteger(message string) validate.ValidatorFn {
	return func(values []string) error {
		if len(values) == 0 || values[0] == "" {
			return nil
		}
		value, err := strconv.Atoi(values[0])
		if err != nil || value < 1 {
			return errors.New(message)
		}
		return nil
	}
}
