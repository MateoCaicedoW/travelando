package dashboard

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"go.leapkit.dev/core/server"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func MarkCommissionPaid(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Inquiries == nil || set.Commissions == nil || set.Deposits == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := set.Inquiries.MarkCommissionPaid(id); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			http.Error(w, "Esta reserva no tiene saldo pendiente de comisión.", http.StatusConflict)
			return
		}
		slog.Error("marking reservation commission as paid", "inquiry_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	data, err := loadDashboardData(set.Inquiries, set.Commissions, set.Deposits)
	if err != nil {
		slog.Error("reloading dashboard after commission payment", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if !helpers.IsHTMX(r) {
		if err := helpers.SetFlash(r, helpers.FlashSuccess, "Saldo de comisión marcado como pagado."); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
		helpers.Redirect(w, r, "/admin/")
		return
	}
	w.WriteHeader(http.StatusOK)
	response := Group{
		dashboardEl(data.Inquiries, data.CommissionBP, data.DepositCOP, nil, nil, requestedCabinPage(r)),
		Div(ID("global-flash-container"), Attr("hx-swap-oob", "true"), components.FlashToast("Saldo de comisión marcado como pagado.", helpers.FlashSuccess)),
	}
	if err := response.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
