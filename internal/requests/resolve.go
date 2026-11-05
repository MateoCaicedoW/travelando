package requests

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

func ReportResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	resolveReport(w, r)
}

func resolveReport(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		renderError(w, r, http.StatusInternalServerError, "No se pudo guardar el resultado.")
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
	fieldErrors := resolutionFormErrors(r, inquiry.DepositCOP)
	if fieldErrors.HasAny() {
		renderReportPage(w, r, http.StatusUnprocessableEntity, "Reportar resultado | Travelando", *inquiry, token, r.Form, fieldErrors)
		return
	}
	status := models.InquiryStatus(r.FormValue("status"))
	finalTotal := int64(0)
	if status == models.InquiryReserved {
		finalTotal, err = strconv.ParseInt(strings.ReplaceAll(r.FormValue("final_total"), ".", ""), 10, 64)
		if err != nil {
			fieldErrors["final_total"] = append(fieldErrors["final_total"], errors.New("Ingresa un valor final válido en pesos."))
			renderReportPage(w, r, http.StatusUnprocessableEntity, "Reportar resultado | Travelando", *inquiry, token, r.Form, fieldErrors)
			return
		}
	}
	if err := svc.ResolveInquiry(token, status, finalTotal); err != nil {
		if errors.Is(err, services.ErrUnavailable) {
			if !setReportFlash(w, r, "La cabaña ya tiene otra reserva confirmada en esas fechas. Contacta a Travelando.") {
				return
			}
			renderReportPage(w, r, http.StatusConflict, "Reportar resultado | Travelando", *inquiry, token, r.Form, nil)
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			if !setReportFlash(w, r, "Esta solicitud ya fue resuelta.") {
				return
			}
			renderReportPage(w, r, http.StatusConflict, "Reportar resultado | Travelando", *inquiry, token, r.Form, nil)
			return
		}
		slog.Error("resolving rental inquiry", "error", err)
		renderError(w, r, http.StatusInternalServerError, "No se pudo guardar el resultado.")
		return
	}
	inquiry, err = svc.InquiryByToken(token)
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "El resultado se guardó, pero no se pudo cargar el comprobante.")
		return
	}
	if status == models.InquiryReserved {
		if err := services.SendReservationConfirmation(*inquiry); err != nil {
			slog.Error("sending guest reservation confirmation", "error", err)
			if !setReportFlash(w, r, "La reserva quedó registrada, pero no se pudo enviar el correo. Contacta al huésped para compartir la confirmación.") {
				return
			}
			if helpers.IsHTMX(r) {
				helpers.Redirect(w, r, "/report/"+token)
				return
			}
		}
	}
	renderReportPage(w, r, 0, "Resultado registrado | Travelando", *inquiry, token, nil, nil)
}

func resolutionFormErrors(r *http.Request, depositCOP int64) validate.Errors {
	status := r.FormValue("status")
	return form.Validate(r, validate.Fields(
		validate.Field("status", validate.Required("Selecciona el resultado."), validate.WithinOptions([]string{string(models.InquiryReserved), string(models.InquiryLost)}, "Selecciona un resultado válido.")),
		validate.Field("final_total", validFinalTotal(status, depositCOP)),
	))
}

func validFinalTotal(status string, depositCOP int64) validate.ValidatorFn {
	return func(values []string) error {
		if status != string(models.InquiryReserved) {
			return nil
		}
		if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
			return errors.New("Ingresa el valor total de la reserva.")
		}
		value, err := strconv.ParseInt(strings.ReplaceAll(values[0], ".", ""), 10, 64)
		if err != nil || value < depositCOP || value > services.MaxAmountCOP {
			if depositCOP > 0 && value < depositCOP {
				return errors.New("El total acordado debe incluir el depósito de $" + helpers.FormatCOP(depositCOP) + " COP.")
			}
			return errors.New("Ingresa un valor válido entre 0 y 1.000.000.000.000 COP.")
		}
		return nil
	}
}

func setReportFlash(w http.ResponseWriter, r *http.Request, message string) bool {
	if err := helpers.SetFlash(r, helpers.FlashError, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return false
	}
	return true
}
