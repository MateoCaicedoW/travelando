package cabins

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Blocks == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	fieldErrors := blockFormErrors(r)
	if fieldErrors.HasAny() {
		renderBlockErrors(w, r, id, fieldErrors)
		return
	}
	block := models.CabinBlock{CabinID: id, DateFrom: r.FormValue("date_from"), DateTo: r.FormValue("date_to"), Reason: r.FormValue("reason")}
	if err := set.Blocks.AddBlock(&block); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			fieldErrors["date_to"] = append(fieldErrors["date_to"], errors.New("Revisa el rango de fechas."))
			renderBlockErrors(w, r, id, fieldErrors)
			return
		}
		message := "No se pudo bloquear ese rango."
		if errors.Is(err, services.ErrUnavailable) {
			message = "Ese rango tiene una reserva confirmada."
		} else {
			slog.Error("adding cabin block", "error", err)
		}
		if err := helpers.SetFlash(r, helpers.FlashError, message); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
		helpers.Redirect(w, r, "/admin/cabins/"+strconv.FormatInt(id, 10)+"/edit")
		return
	}
	if hxhttp.IsRequest(r.Header) {
		refreshCabinEdit(w, r, id)
		return
	}
	helpers.Redirect(w, r, "/admin/cabins/"+strconv.FormatInt(id, 10)+"/edit")
}

func renderBlockErrors(w http.ResponseWriter, r *http.Request, id int64, fieldErrors map[string][]error) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil || set.Blocks == nil || set.Inquiries == nil || set.Owners == nil || set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	data, err := loadCabinEditData(set.Cabins, set.Blocks, set.Inquiries, set.Owners, set.Amenities, id)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	state := &cabinFormState{BlockErrors: fieldErrors, BlockValues: r.Form}
	renderCabinEditContent(w, r, data, state, http.StatusUnprocessableEntity)
}
