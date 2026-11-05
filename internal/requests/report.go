package requests

import (
	"errors"
	"log/slog"
	"net/http"

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
	renderReportPage(w, r, 0, "Reportar resultado | Travelando", *inquiry, token, nil, nil)
}
