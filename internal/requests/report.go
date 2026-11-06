package requests

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func Report(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		renderError(w, r, http.StatusInternalServerError, "No se pudo cargar la solicitud.")
		return
	}
	token := r.PathValue("token")
	inquiry, err := svc.InquiryByToken(token)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading rental report", "error", err)
		renderError(w, r, http.StatusInternalServerError, "No se pudo cargar la solicitud.")
		return
	}
	cabin, err := reportCabin(services.ServiceSetFromCtx(r.Context()).Cabins, inquiry.CabinID)
	if err != nil {
		slog.Error("loading cabin for rental report", "error", err)
		renderError(w, r, http.StatusInternalServerError, "No se pudo cargar la cabaña.")
		return
	}
	renderReportPage(w, r, 0, "Reportar resultado | Travelando", *inquiry, cabin, token, nil, nil)
}

func reportCabin(svc services.CabinService, cabinID int64) (models.Cabin, error) {
	if svc == nil {
		return models.Cabin{}, errors.New("cabin service unavailable")
	}
	cabin, err := svc.Cabin(cabinID)
	if err != nil {
		return models.Cabin{}, err
	}
	if cabin == nil {
		return models.Cabin{}, errors.New("cabin not found")
	}
	return *cabin, nil
}
