package requests

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"travelando/internal/home"
	"travelando/internal/system"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
)

func Create(w http.ResponseWriter, r *http.Request) {
	target, fieldErrors, err := createRequest(w, r)
	if fieldErrors.HasAny() {
		renderInquiryForm(w, r, fieldErrors, http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		switch {
		case errors.Is(err, errInvalidRequestForm):
			http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		case errors.Is(err, errBaseURLConfiguration):
			renderError(w, r, http.StatusServiceUnavailable, "Travelando necesita configurar BASE_URL para generar enlaces de seguimiento.")
		case errors.Is(err, errInquiryServiceMissing):
			renderError(w, r, http.StatusInternalServerError, "No se pudo registrar la solicitud.")
		case errors.Is(err, services.ErrInvalidInput):
			renderInquiryForm(w, r, fieldError("guests", "No se pudo calcular un precio válido para esa cantidad de huéspedes."), http.StatusUnprocessableEntity)
		case errors.Is(err, services.ErrUnavailable):
			if err := helpers.SetFlash(r, helpers.FlashError, "La cabaña ya no está disponible para esas fechas. Busca otras opciones."); err != nil {
				http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
				return
			}
			if helpers.IsHTMX(r) {
				helpers.Redirect(w, r, "/cabins/"+r.PathValue("id"))
				return
			}
			renderInquiryForm(w, r, nil, http.StatusConflict)
		case errors.Is(err, services.ErrStayTypeNotAccepted):
			renderInquiryForm(w, r, fieldError("stay_type", "Esta cabaña no acepta esa modalidad."), http.StatusUnprocessableEntity)
		case errors.Is(err, services.ErrNotFound):
			http.NotFound(w, r)
		default:
			slog.Error("creating inquiry", "error", err)
			renderError(w, r, http.StatusInternalServerError, "No se pudo registrar la solicitud. Intenta de nuevo.")
		}
		return
	}
	helpers.Redirect(w, r, target)
}

func renderInquiryForm(w http.ResponseWriter, r *http.Request, fieldErrors validate.Errors, status int) {
	cabinID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	svc := set.Cabins
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	cabin, err := svc.Cabin(cabinID)
	if errors.Is(err, services.ErrNotFound) || err == nil && (cabin == nil || !cabin.Active) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if set.Deposits == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	deposit, err := set.Deposits.DepositAmount()
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	data := home.CabinQueryData(r.Form)
	data.DepositCOP = deposit
	data.RequestValues = r.Form
	data.RequestErrors = fieldErrors
	w.WriteHeader(status)
	if helpers.IsHTMX(r) {
		if err := home.CabinEl(*cabin, data).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: cabin.Name + " | Travelando", Description: cabin.Description, Yield: home.CabinEl(*cabin, data)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func parsePositiveID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid ID")
	}
	return id, nil
}
