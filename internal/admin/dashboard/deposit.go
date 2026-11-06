package dashboard

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	hxhttp "maragu.dev/gomponents-htmx/http"
	. "maragu.dev/gomponents/html"
)

type depositFormState struct {
	Value string
	Error string
}

func UpdateDeposit(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Deposits == nil || set.Inquiries == nil || set.Commissions == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	fieldErrors := form.Validate(r, validate.Fields(
		validate.Field("deposit", validate.Required("Ingresa el valor del depósito."), depositAmountValidator),
	))
	if fieldErrors.HasAny() {
		state := &depositFormState{Value: r.FormValue("deposit"), Error: fieldErrors.ErrorStringFor("deposit")}
		if hxhttp.IsRequest(r.Header) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			if err := depositFormEl(state.Value, state.Error).Render(w); err != nil {
				server.Error(w, err, http.StatusInternalServerError)
			}
			return
		}
		data, err := loadDashboardData(set.Inquiries, set.Commissions, set.Deposits)
		if err != nil {
			server.Error(w, err, http.StatusInternalServerError)
			return
		}
		renderDashboard(w, r, data, nil, state, http.StatusUnprocessableEntity)
		return
	}
	amount, err := parseDepositAmount(r.FormValue("deposit"))
	if err != nil {
		http.Error(w, "Monto inválido", http.StatusBadRequest)
		return
	}
	if err := set.Deposits.UpdateDepositAmount(amount); err != nil {
		slog.Error("updating global deposit amount", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if hxhttp.IsRequest(r.Header) {
		if err := depositSuccessEl(amount).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Depósito global actualizado."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, "/admin/")
}

func parseDepositAmount(value string) (int64, error) {
	amount, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || amount < 0 || amount > services.MaxAmountCOP {
		return 0, errors.New("deposit must be between 0 and the maximum COP amount")
	}
	return amount, nil
}

func depositAmountValidator(values []string) error {
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return nil
	}
	if _, err := parseDepositAmount(values[0]); err != nil {
		return errors.New("Ingresa un monto válido en COP.")
	}
	return nil
}

func depositFormEl(value, message string) Node {
	inputClass := "w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"
	if message != "" {
		inputClass += " border-red-300 focus:ring-red-200"
	}
	return Form(ID("deposit-form"), Method("post"), Action("/admin/deposit"), hx.Post("/admin/deposit"), hx.Target("#deposit-form"), hx.Swap("outerHTML"), Class("mt-4"),
		Div(Class("min-w-0"),
			components.FieldLabel("deposit", "Monto fijo (COP)"),
			Input(ID("deposit"), Name("deposit"), Type("number"), Min("0"), Max(strconv.FormatInt(services.MaxAmountCOP, 10)), Step("1"), Value(value), Required(), If(message != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "deposit-error")}), Class("min-h-12 "+inputClass)),
			components.FieldError("deposit", message),
		),
		components.FormActions(
			Button(Type("submit"), Class(components.FormPrimaryClass), Text("Guardar depósito")),
		),
	)
}

func depositSuccessEl(amount int64) Node {
	return Group{
		depositFormEl(strconv.FormatInt(amount, 10), ""),
		Div(ID("global-flash-container"), hx.SwapOOB("true"), components.FlashToast("Depósito global actualizado.", helpers.FlashSuccess)),
	}
}
