package reservations

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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

func EditAdmin(w http.ResponseWriter, r *http.Request) {
	edit(w, r, 0)
}

func UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	update(w, r, 0)
}

func EditOwner(w http.ResponseWriter, r *http.Request, ownerID int64) {
	edit(w, r, ownerID)
}

func UpdateOwner(w http.ResponseWriter, r *http.Request, ownerID int64) {
	update(w, r, ownerID)
}

func edit(w http.ResponseWriter, r *http.Request, ownerID int64) {
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
		slog.Error("loading reservation for editing", "inquiry_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if inquiry.Status == models.InquiryCancelled {
		http.NotFound(w, r)
		return
	}
	render(w, r, *inquiry, ownerID, nil, http.StatusOK)
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
	var update models.ReservationUpdate
	var pendingUpdate models.InquiryUpdate
	var fieldErrors validate.Errors
	if inquiry.Status == models.InquiryPending {
		pendingUpdate, fieldErrors = pendingUpdateFromForm(r)
	} else {
		update, fieldErrors = updateFromForm(r)
	}
	state := &formState{Values: r.Form, Errors: fieldErrors}
	if fieldErrors.HasAny() {
		render(w, r, *inquiry, ownerID, state, http.StatusUnprocessableEntity)
		return
	}
	if inquiry.Status == models.InquiryPending {
		if ownerID > 0 {
			err = set.Inquiries.UpdatePendingInquiryForOwner(ownerID, id, pendingUpdate)
		} else {
			err = set.Inquiries.UpdatePendingInquiry(id, pendingUpdate)
		}
	} else if ownerID > 0 {
		err = set.Inquiries.UpdateReservationForOwner(ownerID, id, update)
	} else {
		err = set.Inquiries.UpdateReservation(id, update)
	}
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
	message := "Reserva actualizada."
	if inquiry.Status == models.InquiryPending {
		message = "Solicitud actualizada. El estimado fue recalculado."
	} else if ownerID > 0 {
		message = "Reserva actualizada. El cambio de comisión quedó registrado."
	}
	returnTo := ""
	if r.FormValue("return_to") == "reservations" {
		returnTo = "/admin/reservations"
		if ownerID > 0 {
			returnTo = "/owner/reservations"
		}
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
	if ownerID > 0 {
		helpers.Redirect(w, r, "/owner/")
		return
	}
	helpers.Redirect(w, r, "/admin/")
}

func loadReservation(svc services.InquiryService, ownerID, id int64) (*models.Inquiry, error) {
	if ownerID > 0 {
		return svc.InquiryForOwner(ownerID, id)
	}
	return svc.AdminReservation(id)
}

func render(w http.ResponseWriter, r *http.Request, inquiry models.Inquiry, ownerID int64, state *formState, status int) {
	owner := ownerID > 0
	returnTo := r.URL.Query().Get("return_to")
	if state != nil && state.Values.Get("return_to") != "" {
		returnTo = state.Values.Get("return_to")
	}
	if returnTo != "reservations" {
		returnTo = ""
	}
	if hxhttp.IsRequest(r.Header) && state != nil {
		if status != http.StatusOK {
			w.WriteHeader(status)
		}
		if err := reservationForm(inquiry, state, owner, returnTo).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	label := "Editar reserva"
	if inquiry.Status == models.InquiryPending {
		label = "Editar solicitud"
	}
	dialog := components.CustomDialog(components.DialogProps{Label: label, Yield: reservationForm(inquiry, state, owner, returnTo)})
	if err := dialog.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func reservationID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid reservation ID")
	}
	return id, nil
}
