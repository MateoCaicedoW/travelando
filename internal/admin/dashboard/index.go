package dashboard

import (
	"log/slog"
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Inquiries == nil || set.Commissions == nil || set.Deposits == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	data, err := loadDashboardData(set.Inquiries, set.Commissions, set.Deposits)
	if err != nil {
		slog.Error("loading admin dashboard", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	renderDashboard(w, r, data, nil, nil, http.StatusOK)
}

func renderDashboard(w http.ResponseWriter, r *http.Request, data dashboardData, commissionState *commissionFormState, depositState *depositFormState, status int) {
	if hxhttp.IsRequest(r.Header) {
		w.WriteHeader(status)
		if err := dashboardEl(data.Inquiries, data.CommissionBP, data.DepositCOP, commissionState, depositState, requestedCabinPage(r)).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Panel | Travelando", Portal: system.PortalAdmin, Yield: dashboardEl(data.Inquiries, data.CommissionBP, data.DepositCOP, commissionState, depositState, requestedCabinPage(r))}.New(r)
	w.WriteHeader(status)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}

}
