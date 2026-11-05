package requests

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

var (
	errInvalidRequestForm    = errors.New("invalid request form")
	errBaseURLConfiguration  = errors.New("invalid BASE_URL configuration")
	errInquiryServiceMissing = errors.New("inquiry service unavailable")
)

func createRequest(w http.ResponseWriter, r *http.Request) (string, validate.Errors, error) {
	if err := helpers.ParseForm(w, r); err != nil {
		return "", nil, errInvalidRequestForm
	}
	cabinID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || cabinID < 1 {
		return "", nil, services.ErrNotFound
	}
	fieldErrors := inquiryFormErrors(r)
	if fieldErrors.HasAny() {
		return "", fieldErrors, nil
	}
	baseURL, err := appBaseURL(r)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", errBaseURLConfiguration, err)
	}
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		return "", nil, errInquiryServiceMissing
	}
	guests, err := strconv.Atoi(r.FormValue("guests"))
	if err != nil {
		return "", fieldError("guests", "Ingresa una cantidad válida de huéspedes."), nil
	}
	inquiry := models.Inquiry{
		CabinID: cabinID, GuestName: r.FormValue("guest_name"), GuestPhone: r.FormValue("guest_phone"),
		GuestEmail: r.FormValue("guest_email"), StayType: models.StayType(r.FormValue("stay_type")),
		DateFrom: r.FormValue("date_from"), DateTo: r.FormValue("date_to"), Guests: guests,
	}
	token, err := svc.CreateInquiry(&inquiry)
	if err != nil {
		return "", nil, err
	}
	stay := "Pasadía"
	if inquiry.StayType == models.StayOvernight {
		stay = "Estadía"
	}
	// shortcut: report token is visible to guest in the outgoing WhatsApp text, scoped to one inquiry; add contact authentication if that access becomes unacceptable.
	reportURL := strings.TrimRight(baseURL, "/") + "/report/" + token
	message := fmt.Sprintf("Hola, quiero consultar por %s.\n\nNombre: %s\nTeléfono: %s\nPlan: %s\nFechas: %s\nHorario: %s\nHuéspedes: %d\nPrecio estimado: $%s COP\n\nReferencia para reportar el resultado: %s",
		inquiry.CabinName, inquiry.GuestName, inquiry.GuestPhone, stay, helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo), helpers.ScheduleLabel(inquiry.StayType, inquiry.StartTime, inquiry.EndTime), inquiry.Guests, helpers.FormatCOP(inquiry.Estimated), reportURL)
	return "https://wa.me/" + inquiry.WhatsApp + "?" + url.Values{"text": {message}}.Encode(), nil, nil
}

func inquiryFormErrors(r *http.Request) validate.Errors {
	return form.Validate(r, validate.Fields(
		validate.Field("stay_type", validate.Required("Selecciona el tipo de plan."), validate.WithinOptions([]string{string(models.StayDayUse), string(models.StayOvernight)}, "Selecciona pasadía o estadía.")),
		validate.Field("date_from", validate.Required("Selecciona la fecha de llegada o de uso."), validStartDate()),
		validate.Field("date_to", validEndDate(r.FormValue("stay_type"), r.FormValue("date_from"))),
		validate.Field("guests", validate.Required("Indica cuántos huéspedes asistirán."), guestCountRange),
		validate.Field("guest_name", validate.Required("Ingresa tu nombre."), validate.MaxLength(120, "Usa máximo 120 caracteres.")),
		validate.Field("guest_phone", validate.Required("Ingresa tu teléfono."), validGuestPhone),
		validate.Field("guest_email", validate.Required("Ingresa tu correo."), validate.MaxLength(254, "Usa máximo 254 caracteres."), validGuestEmail),
	))
}

func validStartDate() validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) {
			return nil
		}
		_, err := time.Parse(time.DateOnly, values[0])
		if err != nil || values[0] < time.Now().Format(time.DateOnly) {
			return errors.New("Selecciona una fecha válida desde hoy en adelante.")
		}
		return nil
	}
}

func validEndDate(stayType, start string) validate.ValidatorFn {
	return func(values []string) error {
		value := ""
		if len(values) > 0 {
			value = values[0]
		}
		if stayType == string(models.StayDayUse) {
			if strings.TrimSpace(value) == "" {
				return nil
			}
			if value != start {
				return errors.New("Para pasadía, la fecha final debe coincidir con la fecha de uso.")
			}
			return nil
		}
		if stayType != string(models.StayOvernight) {
			return nil
		}
		if strings.TrimSpace(value) == "" {
			return errors.New("Selecciona la fecha de salida.")
		}
		from, fromErr := time.Parse(time.DateOnly, start)
		to, toErr := time.Parse(time.DateOnly, value)
		if fromErr != nil || toErr != nil || !to.After(from) || to.Sub(from).Hours()/24 > 365 {
			return errors.New("La salida debe ser posterior a la llegada y estar dentro de un año.")
		}
		return nil
	}
}

func guestCountRange(values []string) error {
	if missing(values) {
		return nil
	}
	guests, err := strconv.Atoi(values[0])
	if err != nil || guests < 1 {
		return errors.New("Ingresa una cantidad válida de huéspedes.")
	}
	return nil
}

func validGuestPhone(values []string) error {
	if missing(values) {
		return nil
	}
	digits := helpers.DigitsOnly(values[0])
	if len(digits) < 8 || len(digits) > 15 {
		return errors.New("Ingresa un teléfono con entre 8 y 15 dígitos.")
	}
	return nil
}

func validGuestEmail(values []string) error {
	if missing(values) {
		return nil
	}
	value := strings.TrimSpace(values[0])
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return errors.New("Ingresa un correo válido.")
	}
	return nil
}

func fieldError(field, message string) validate.Errors {
	return validate.Errors{field: {errors.New(message)}}
}

func missing(values []string) bool {
	return len(values) == 0 || strings.TrimSpace(values[0]) == ""
}
