package ownerportal

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

type blockFormState struct {
	Errors validate.Errors
	Values url.Values
}

func AddBlock(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	state := &blockFormState{Errors: externalBlockFormErrors(r), Values: r.Form}
	if state.Errors.HasAny() {
		if err := renderOwnerCalendarWithBlockModal(w, r, state, http.StatusUnprocessableEntity); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	cabinID, _ := strconv.ParseInt(r.FormValue("cabin_id"), 10, 64)
	block := models.CabinBlock{CabinID: cabinID, DateFrom: r.FormValue("date_from"), DateTo: r.FormValue("date_to")}
	svc := services.ServiceSetFromCtx(r.Context()).Blocks
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := svc.AddExternalBlock(ownerIDFromSession(r), &block); err != nil {
		switch {
		case errors.Is(err, services.ErrUnavailable):
			if helpers.IsHTMX(r) {
				addBlockFormError(state, "date_to", "Ese rango ya tiene una reserva o bloqueo.")
				if err := renderOwnerCalendarWithBlockModal(w, r, state, http.StatusUnprocessableEntity); err != nil {
					server.Error(w, err, http.StatusInternalServerError)
				}
				return
			}
			if err := helpers.SetFlash(r, helpers.FlashError, "Ese rango ya tiene una reserva o bloqueo."); err != nil {
				http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
				return
			}
			helpers.Redirect(w, r, ownerMonthURLForDate(block.DateFrom))
		case errors.Is(err, services.ErrNotFound):
			http.NotFound(w, r)
		case errors.Is(err, services.ErrInvalidInput):
			addBlockFormError(state, "date_to", "Revisa el rango de fechas.")
			if err := renderOwnerCalendarWithBlockModal(w, r, state, http.StatusUnprocessableEntity); err != nil {
				server.Error(w, err, http.StatusInternalServerError)
			}
		default:
			slog.Error("adding owner external block", "error", err)
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if helpers.IsHTMX(r) {
		if err := renderOwnerCalendar(w, r, http.StatusOK, helpers.FlashSuccess, "Reserva externa agregada al calendario.", true); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Reserva externa agregada al calendario."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, ownerMonthURLForDate(block.DateFrom))
}

func DeleteBlock(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	blockID, err := strconv.ParseInt(r.PathValue("blockID"), 10, 64)
	if err != nil || blockID < 1 {
		http.NotFound(w, r)
		return
	}
	cabinID, err := strconv.ParseInt(r.FormValue("cabin_id"), 10, 64)
	if err != nil || cabinID < 1 {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Blocks
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := svc.DeleteExternalBlock(ownerIDFromSession(r), cabinID, blockID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("deleting owner external block", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if helpers.IsHTMX(r) {
		if err := renderOwnerCalendar(w, r, http.StatusOK, helpers.FlashSuccess, "Reserva externa quitada del calendario.", false); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Reserva externa quitada del calendario."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, ownerMonthURLForDate(r.FormValue("date_from")))
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
