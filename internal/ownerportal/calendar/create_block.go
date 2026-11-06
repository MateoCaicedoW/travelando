package calendar

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

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
	if err := svc.AddExternalBlock(auth.OwnerIDFromSession(r), &block); err != nil {
		switch {
		case errors.Is(err, services.ErrUnavailable):
			if hxhttp.IsRequest(r.Header) {
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
	if hxhttp.IsRequest(r.Header) {
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
