package dashboard

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

type commissionFormState struct {
	Value string
	Error string
}

func UpdateCommission(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	svc := set.Commissions
	if svc == nil || set.Inquiries == nil || set.Deposits == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	fieldErrors := form.Validate(r, validate.Fields(
		validate.Field("commission", validate.Required("Ingresa el porcentaje de comisión."), commissionRateValidator),
	))
	if fieldErrors.HasAny() {
		state := &commissionFormState{Value: r.FormValue("commission"), Error: fieldErrors.ErrorStringFor("commission")}
		if hxhttp.IsRequest(r.Header) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			if err := commissionFormEl(state.Value, state.Error).Render(w); err != nil {
				server.Error(w, err, http.StatusInternalServerError)
			}
			return
		}
		data, err := loadDashboardData(set.Inquiries, set.Commissions, set.Deposits)
		if err != nil {
			server.Error(w, err, http.StatusInternalServerError)
			return
		}
		renderDashboard(w, r, data, state, nil, http.StatusUnprocessableEntity)
		return
	}
	rate, err := parseCommissionRate(r.FormValue("commission"))
	if err != nil {
		http.Error(w, "Porcentaje inválido", http.StatusBadRequest)
		return
	}
	if err := svc.UpdateCommissionRate(rate); err != nil {
		slog.Error("updating global commission rate", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if hxhttp.IsRequest(r.Header) {
		if err := commissionSuccessEl(rate).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Comisión global actualizada."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, "/admin/")
}

func parseCommissionRate(value string) (int, error) {
	rate, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 || rate > 100 {
		return 0, errors.New("commission must be between 0 and 100 percent")
	}
	return int(math.Round(rate * 100)), nil
}

func commissionRateValidator(values []string) error {
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return nil
	}
	if _, err := parseCommissionRate(values[0]); err != nil {
		return errors.New("Ingresa un porcentaje entre 0 y 100.")
	}
	return nil
}
