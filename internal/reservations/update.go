package reservations

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	hxhttp "maragu.dev/gomponents-htmx/http"
	. "maragu.dev/gomponents/html"
)

func UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	update(w, r, 0)
}

func UpdateOwner(w http.ResponseWriter, r *http.Request, ownerID int64) {
	update(w, r, ownerID)
}

func update(w http.ResponseWriter, r *http.Request, ownerID int64) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := reservationID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Inquiries == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiry, err := loadReservation(set.Inquiries, ownerID, id)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	isOwner := ownerID > 0
	isPending := inquiry.Status == models.InquiryPending
	var update models.ReservationUpdate
	var pendingUpdate models.InquiryUpdate
	var fieldErrors validate.Errors
	if isPending {
		pendingUpdate, fieldErrors = pendingUpdateFromForm(r)
	} else {
		update, fieldErrors = updateFromForm(r)
	}
	state := &formState{Values: r.Form, Errors: fieldErrors}
	if fieldErrors.HasAny() {
		render(w, r, *inquiry, ownerID, state, http.StatusUnprocessableEntity)
		return
	}
	err = persistUpdate(set.Inquiries, isOwner, isPending, ownerID, id, update, pendingUpdate)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, services.ErrUnavailable) {
		state.Notice = "Las fechas se cruzan con otra reserva o bloqueo."
		render(w, r, *inquiry, ownerID, state, http.StatusConflict)
		return
	}
	if errors.Is(err, services.ErrStayTypeNotAccepted) || errors.Is(err, services.ErrInvalidInput) {
		state.Notice = "Revisa los datos ingresados."
		render(w, r, *inquiry, ownerID, state, http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		slog.Error("updating reservation", "inquiry_id", id, "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	message := updateSuccessMessage(isPending, isOwner)
	returnTo := ""
	if r.FormValue("return_to") == "reservations" {
		returnTo = reservationsReturnPath(isOwner)
	}
	if returnTo != "" {
		if err := helpers.SetFlash(r, helpers.FlashSuccess, message); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
		helpers.Redirect(w, r, returnTo)
		return
	}
	if hxhttp.IsRequest(r.Header) {
		response := Group{
			Div(ID("modal"), hx.SwapOOB("delete")),
			Div(ID("global-flash-container"), hx.SwapOOB("true"), components.FlashToast(message, helpers.FlashSuccess)),
		}
		if err := response.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	if isOwner {
		helpers.Redirect(w, r, "/owner/")
		return
	}
	helpers.Redirect(w, r, "/admin/")
}

func persistUpdate(svc services.InquiryService, isOwner, isPending bool, ownerID, id int64, update models.ReservationUpdate, pendingUpdate models.InquiryUpdate) error {
	if isPending {
		if isOwner {
			return svc.UpdatePendingInquiryForOwner(ownerID, id, pendingUpdate)
		}
		return svc.UpdatePendingInquiry(id, pendingUpdate)
	}
	if isOwner {
		return svc.UpdateReservationForOwner(ownerID, id, update)
	}
	return svc.UpdateReservation(id, update)
}

func updateSuccessMessage(isPending, isOwner bool) string {
	if isPending {
		return "Solicitud actualizada. El estimado fue recalculado."
	}
	if isOwner {
		return "Reserva actualizada. El cambio de comisión quedó registrado."
	}
	return "Reserva actualizada."
}

func reservationsReturnPath(isOwner bool) string {
	if isOwner {
		return "/owner/reservations"
	}
	return "/admin/reservations"
}
