package cabins

import (
	"errors"
	"net/http"
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

type cabinFormState struct {
	Errors       validate.Errors
	Values       url.Values
	AmenityTypes []models.AmenityType
	BlockErrors  validate.Errors
	BlockValues  url.Values
}

func cabinFormErrors(r *http.Request) validate.Errors {
	return cabinFormErrorsFor(r, true)
}

func cabinFormErrorsFor(r *http.Request, requireOwner bool) validate.Errors {
	errors := form.Validate(r, validate.Fields(
		validate.Field("name", validate.Required("Ingresa el nombre de la cabaña."), validate.MaxLength(120, "Usa máximo 120 caracteres.")),
		validate.Field("description", validate.MaxLength(2000, "Usa máximo 2000 caracteres.")),
		validate.Field("location", validate.Required("Ingresa la ubicación."), validate.MaxLength(200, "Usa máximo 200 caracteres.")),
		validate.Field("location_url", optionalURL("Ingresa un enlace HTTP(S) válido.")),
		validate.Field("included_guests", validate.Required("Indica cuántas personas incluye el precio."), positiveInteger("Ingresa una cantidad válida de personas.")),
		validate.Field("extra_guest_price", validate.Required("Ingresa el cargo por persona adicional."), integer64Range(0, services.MaxAmountCOP, "Ingresa un cargo válido en COP.")),
		validate.Field("day_price", validate.Required("Ingresa el precio de pasadía."), integer64Range(0, services.MaxAmountCOP, "Ingresa un precio válido en COP.")),
		validate.Field("night_price", validate.Required("Ingresa el precio por noche."), integer64Range(0, services.MaxAmountCOP, "Ingresa un precio válido en COP.")),
		validate.Field("day_start", validate.Required("Ingresa la hora de inicio."), validCabinTime("Ingresa una hora válida.")),
		validate.Field("day_end", validate.Required("Ingresa la hora de fin."), validCabinTime("Ingresa una hora válida."), afterCabinTime(r.FormValue("day_start"))),
		validate.Field("check_in", validate.Required("Ingresa la hora de entrada."), validCabinTime("Ingresa una hora válida.")),
		validate.Field("check_out", validate.Required("Ingresa la hora de salida."), validCabinTime("Ingresa una hora válida.")),
		validate.Field("cleaning_hours", validate.Required("Indica las horas de aseo."), integerRange(0, 24, "Ingresa un tiempo de aseo entre 0 y 24 horas.")),
	))
	if requireOwner {
		ownerErrors := form.Validate(r, validate.Fields(validate.Field("owner_id", validate.Required("Asigna un dueño."), validOwnerID)))
		if errors == nil && ownerErrors.HasAny() {
			errors = validate.Errors{}
		}
		for field, messages := range ownerErrors {
			errors[field] = append(errors[field], messages...)
		}
	}
	return errors
}

func cabinFromForm(r *http.Request, id int64) (models.Cabin, validate.Errors) {
	return cabinFromRequest(r, id, true, 0)
}

func OwnerCabinFromForm(r *http.Request, id, ownerID int64) (models.Cabin, validate.Errors) {
	return cabinFromRequest(r, id, false, ownerID)
}

func cabinFromRequest(r *http.Request, id int64, requireOwner bool, ownerID int64) (models.Cabin, validate.Errors) {
	fieldErrors := cabinFormErrorsFor(r, requireOwner)
	cabin := models.Cabin{
		ID: id, Name: r.FormValue("name"), Description: r.FormValue("description"), Location: r.FormValue("location"),
		LocationURL:   r.FormValue("location_url"),
		AcceptsDayUse: r.FormValue("accepts_day_use") == "on", AcceptsOvernight: r.FormValue("accepts_overnight") == "on",
		DayStart: r.FormValue("day_start"), DayEnd: r.FormValue("day_end"), CheckIn: r.FormValue("check_in"), CheckOut: r.FormValue("check_out"),
	}
	if requireOwner {
		if rawOwnerID := strings.TrimSpace(r.FormValue("owner_id")); rawOwnerID != "" {
			cabin.OwnerID, _ = strconv.ParseInt(rawOwnerID, 10, 64)
		}
	} else {
		cabin.OwnerID = ownerID
	}
	if !fieldErrors.Has("included_guests") {
		value, err := strconv.Atoi(r.FormValue("included_guests"))
		if err != nil {
			addCabinFormError(fieldErrors, "included_guests", "Ingresa una cantidad válida de personas.")
		} else {
			cabin.IncludedGuests = value
		}
	}
	if !fieldErrors.Has("extra_guest_price") {
		value, err := strconv.ParseInt(r.FormValue("extra_guest_price"), 10, 64)
		if err != nil {
			addCabinFormError(fieldErrors, "extra_guest_price", "Ingresa un cargo válido en COP.")
		} else {
			cabin.ExtraGuestPrice = value
		}
	}
	if !fieldErrors.Has("day_price") {
		value, err := strconv.ParseInt(r.FormValue("day_price"), 10, 64)
		if err != nil {
			addCabinFormError(fieldErrors, "day_price", "Ingresa un precio válido en COP.")
		} else {
			cabin.DayPrice = value
		}
	}
	if !fieldErrors.Has("night_price") {
		value, err := strconv.ParseInt(r.FormValue("night_price"), 10, 64)
		if err != nil {
			addCabinFormError(fieldErrors, "night_price", "Ingresa un precio válido en COP.")
		} else {
			cabin.NightPrice = value
		}
	}
	if !fieldErrors.Has("cleaning_hours") {
		value, err := strconv.Atoi(r.FormValue("cleaning_hours"))
		if err != nil {
			addCabinFormError(fieldErrors, "cleaning_hours", "Ingresa un tiempo de aseo entre 0 y 24 horas.")
		} else {
			cabin.CleaningHours = value
		}
	}
	return cabin, fieldErrors
}

func CabinAmenitiesFromForm(r *http.Request, types []models.AmenityType) ([]models.CabinAmenity, validate.Errors) {
	amenities := make([]models.CabinAmenity, 0, len(types))
	fieldErrors := validate.Errors{}
	for _, amenityType := range types {
		key := "amenity_" + strconv.FormatInt(amenityType.ID, 10)
		amenity := models.CabinAmenity{TypeID: amenityType.ID, Name: amenityType.Name, ValueType: amenityType.ValueType}
		switch amenityType.ValueType {
		case models.AmenityBoolean:
			values, submitted := r.Form[key]
			if submitted && (len(values) != 1 || values[0] != "on") {
				addCabinFormError(fieldErrors, key, "Selecciona un valor válido.")
			} else {
				amenity.BooleanValue = submitted
			}
		case models.AmenityQuantity:
			values, submitted := r.Form[key]
			if !submitted {
				amenity.QuantityValue = 0
			} else if len(values) != 1 {
				addCabinFormError(fieldErrors, key, "Ingresa una cantidad válida.")
			} else if amount, err := strconv.Atoi(values[0]); err != nil || amount < 0 {
				addCabinFormError(fieldErrors, key, "Ingresa una cantidad válida mayor o igual a cero.")
			} else {
				amenity.QuantityValue = amount
			}
		default:
			addCabinFormError(fieldErrors, key, "El tipo de amenidad no es válido.")
		}
		amenities = append(amenities, amenity)
	}
	return amenities, fieldErrors
}

func MergeCabinFormErrors(target, source validate.Errors) validate.Errors {
	if len(source) == 0 {
		return target
	}
	if target == nil {
		target = validate.Errors{}
	}
	for field, messages := range source {
		target[field] = append(target[field], messages...)
	}
	return target
}

func validOwnerID(values []string) error {
	if missing(values) {
		return nil
	}
	id, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || id < 1 {
		return errors.New("Selecciona un dueño válido.")
	}
	return nil
}

func cabinIDFromPath(r *http.Request) (int64, error) {
	value := r.PathValue("id")
	if value == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid cabin ID")
	}
	return id, nil
}

func blockFormErrors(r *http.Request) validate.Errors {
	return form.Validate(r, validate.Fields(
		validate.Field("date_from", validate.Required("Selecciona la fecha inicial."), validBlockDate("Selecciona una fecha válida.")),
		validate.Field("date_to", validate.Required("Selecciona la fecha final."), validBlockDate("Selecciona una fecha válida."), afterBlockDate(r.FormValue("date_from"))),
		validate.Field("reason", validate.MaxLength(200, "Usa máximo 200 caracteres.")),
	))
}

func positiveInteger(message string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) {
			return nil
		}
		value, err := strconv.Atoi(values[0])
		if err != nil || value < 1 {
			return errors.New(message)
		}
		return nil
	}
}

func integerRange(min, max int, message string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) {
			return nil
		}
		value, err := strconv.Atoi(values[0])
		if err != nil || value < min || value > max {
			return errors.New(message)
		}
		return nil
	}
}

func integer64Range(min, max int64, message string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) {
			return nil
		}
		value, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || value < min || value > max {
			return errors.New(message)
		}
		return nil
	}
}

func optionalURL(message string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) || helpers.ValidExternalURL(values[0]) {
			return nil
		}
		return errors.New(message)
	}
}

func validCabinTime(message string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) || helpers.ValidTime(values[0]) {
			return nil
		}
		return errors.New(message)
	}
}

func afterCabinTime(start string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) || !helpers.ValidTime(start) || !helpers.ValidTime(values[0]) {
			return nil
		}
		if !helpers.ValidTimeRange(start, values[0]) {
			return errors.New("La hora de fin debe ser posterior a la hora de inicio.")
		}
		return nil
	}
}

func validBlockDate(message string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) {
			return nil
		}
		if _, err := time.Parse(time.DateOnly, values[0]); err != nil {
			return errors.New(message)
		}
		return nil
	}
}

func afterBlockDate(from string) validate.ValidatorFn {
	return func(values []string) error {
		if missing(values) || strings.TrimSpace(from) == "" {
			return nil
		}
		start, startErr := time.Parse(time.DateOnly, from)
		end, endErr := time.Parse(time.DateOnly, values[0])
		if startErr != nil || endErr != nil || !end.After(start) {
			return errors.New("La fecha final debe ser posterior a la inicial.")
		}
		return nil
	}
}

func missing(values []string) bool {
	return len(values) == 0 || strings.TrimSpace(values[0]) == ""
}

func addCabinFormError(fieldErrors validate.Errors, field, message string) {
	fieldErrors[field] = append(fieldErrors[field], errors.New(message))
}
