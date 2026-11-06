package reservations

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func ShowConfirmAdmin(w http.ResponseWriter, r *http.Request) {
	showConfirm(w, r, 0)
}

func ShowConfirmOwner(w http.ResponseWriter, r *http.Request, ownerID int64) {
	showConfirm(w, r, ownerID)
}

func ConfirmAdmin(w http.ResponseWriter, r *http.Request) {
	confirm(w, r, 0)
}

func ConfirmOwner(w http.ResponseWriter, r *http.Request, ownerID int64) {
	confirm(w, r, ownerID)
}

func showConfirm(w http.ResponseWriter, r *http.Request, ownerID int64) {
	id, err := reservationID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiry, err := loadReservation(svc, ownerID, id)
	if errors.Is(err, services.ErrNotFound) || err == nil && inquiry.Status != models.InquiryPending {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading pending inquiry for confirmation", "inquiry_id", id, "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	owner := ownerID > 0
	if err := confirmReservationDialog(*inquiry, owner, strconv.FormatInt(inquiry.Estimated, 10), "").Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func confirm(w http.ResponseWriter, r *http.Request, ownerID int64) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := reservationID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiry, err := loadReservation(svc, ownerID, id)
	if errors.Is(err, services.ErrNotFound) || err == nil && inquiry.Status != models.InquiryPending {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	totalValue := strings.TrimSpace(r.FormValue("final_total"))
	finalTotal, parseErr := strconv.ParseInt(totalValue, 10, 64)
	if parseErr != nil || finalTotal < inquiry.DepositCOP || finalTotal > services.MaxAmountCOP {
		message := "Ingresa un total válido que incluya el depósito."
		if parseErr == nil && finalTotal < inquiry.DepositCOP {
			message = "El total acordado debe incluir el depósito de $" + helpers.FormatCOP(inquiry.DepositCOP) + " COP."
		}
		renderConfirmationError(w, r, *inquiry, ownerID, totalValue, message, http.StatusUnprocessableEntity)
		return
	}
	if ownerID > 0 {
		err = svc.ConfirmPendingInquiryForOwner(ownerID, id, finalTotal)
	} else {
		err = svc.ConfirmPendingInquiry(id, finalTotal)
	}
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, services.ErrUnavailable) {
		renderConfirmationError(w, r, *inquiry, ownerID, totalValue, "La cabaña ya tiene otra reserva confirmada en esas fechas.", http.StatusConflict)
		return
	}
	if errors.Is(err, services.ErrInvalidInput) {
		renderConfirmationError(w, r, *inquiry, ownerID, totalValue, "Esta solicitud ya fue resuelta o el total no es válido.", http.StatusConflict)
		return
	}
	if err != nil {
		slog.Error("confirming pending inquiry", "inquiry_id", id, "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	inquiry, err = svc.AdminReservation(id)
	if ownerID > 0 {
		inquiry, err = svc.InquiryForOwner(ownerID, id)
	}
	if err != nil {
		slog.Error("loading confirmed reservation", "inquiry_id", id, "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	message := "Reserva confirmada."
	level := helpers.FlashSuccess
	if err := services.SendReservationConfirmation(*inquiry); err != nil {
		slog.Error("sending guest reservation confirmation", "inquiry_id", id, "error", err)
		message = "Reserva confirmada, pero no se pudo enviar el correo. Contacta al huésped para compartir la confirmación."
		level = helpers.FlashError
	}
	if err := helpers.SetFlash(r, level, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	path := "/admin/reservations"
	if ownerID > 0 {
		path = "/owner/reservations"
	}
	helpers.Redirect(w, r, path)
}

func renderConfirmationError(w http.ResponseWriter, r *http.Request, inquiry models.Inquiry, ownerID int64, total, message string, status int) {
	if hxhttp.IsRequest(r.Header) {
		w.WriteHeader(status)
		if err := confirmReservationForm(inquiry, ownerID > 0, total, message).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashError, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	path := "/admin/reservations"
	if ownerID > 0 {
		path = "/owner/reservations"
	}
	helpers.Redirect(w, r, path)
}
