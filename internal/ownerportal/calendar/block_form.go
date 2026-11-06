package calendar

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

type blockFormState struct {
	Errors validate.Errors
	Values url.Values
}

func externalBlockFormErrors(r *http.Request) validate.Errors {
	return form.Validate(r, validate.Fields(
		validate.Field("cabin_id", validate.Required("Elige una cabaña."), validCabinID),
		validate.Field("date_from", validate.Required("Selecciona la fecha inicial."), validDate),
		validate.Field("date_to", validate.Required("Selecciona la fecha final."), validDate, afterDate(r.FormValue("date_from"))),
	))
}

func addBlockFormError(state *blockFormState, field, message string) {
	if state.Errors == nil {
		state.Errors = validate.Errors{}
	}
	state.Errors[field] = append(state.Errors[field], errors.New(message))
}

func validCabinID(values []string) error {
	if missing(values) {
		return nil
	}
	id, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || id < 1 {
		return errors.New("Elige una cabaña válida.")
	}
	return nil
}

func validDate(values []string) error {
	if missing(values) {
		return nil
	}
	if _, err := time.Parse(time.DateOnly, values[0]); err != nil {
		return errors.New("Selecciona una fecha válida.")
	}
	return nil
}

func afterDate(from string) validate.ValidatorFn {
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

func ownerMonthURLForDate(value string) string {
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return "/owner/"
	}
	return ownerMonthURL(date)
}
