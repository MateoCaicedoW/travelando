package calendar

import (
	"strconv"
	"time"

	"travelando/internal/system/components"
	"travelando/internal/system/models"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func BlockFormEl(cabins []models.Cabin, state *blockFormState, month time.Month, year int) Node {
	if len(cabins) == 0 {
		return Div()
	}
	fromValue, toValue, cabinValue := "", "", ""
	fromError, toError, cabinError := "", "", ""
	if state != nil {
		fromValue, toValue, cabinValue = state.Values.Get("date_from"), state.Values.Get("date_to"), state.Values.Get("cabin_id")
		fromError, toError, cabinError = state.Errors.ErrorStringFor("date_from"), state.Errors.ErrorStringFor("date_to"), state.Errors.ErrorStringFor("cabin_id")
	}
	cabinOptions := make([]components.SearchableOption, len(cabins))
	for i, cabin := range cabins {
		cabinOptions[i] = components.SearchableOption{Value: strconv.FormatInt(cabin.ID, 10), Label: cabin.Name}
	}
	formAction := "/owner/blocks?month=" + strconv.Itoa(int(month)) + "&year=" + strconv.Itoa(year)
	return Div(Class("w-full p-4 sm:p-6"),
		Div(Class("flex items-start justify-between gap-4"),
			components.PageHeading("Registrar reserva externa", "Incluye reservas hechas por otros medios o para familiares. No se cobra comisión."),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		),
		Form(Method("post"), Action(formAction), hx.Post(formAction), hx.Target("#owner-content"), hx.Swap("outerHTML"), Class("mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3"),
			Div(Class("min-w-0"),
				components.SearchableSelect("cabin_id", "cabin_id", "Cabaña", "Busca una cabaña", cabinValue, cabinOptions, true, cabinError),
			),
			components.TextFieldWithError("external-from", "date_from", "Desde", fromValue, "", "date", true, fromError),
			components.TextFieldWithError("external-to", "date_to", "Hasta (sin incluir)", toValue, "", "date", true, toError),
			P(Class("text-xs text-muted sm:col-span-2 lg:col-span-3"), Text("La fecha final no se incluye, igual que el checkout.")),
			Div(Class("sm:col-span-2 lg:col-span-3"), components.FormActions(
				Button(Type("submit"), Class(components.FormPrimaryClass), Text("Bloquear fechas")),
				Button(Type("button"), Attr("data-action", "click->modal#close"), Class(components.FormSecondaryClass), Text("Cancelar")),
			)),
		),
	)
}

func BlockDialogEl(cabins []models.Cabin, state *blockFormState, month time.Month, year int, oob bool) Node {
	swapOOB := ""
	if oob {
		swapOOB = "outerHTML"
	}
	return components.CustomDialog(components.DialogProps{Label: "Registrar reserva externa", SwapOOB: swapOOB, Yield: BlockFormEl(cabins, state, month, year)})
}
