package owners

import (
	"strconv"

	lucide "github.com/eduardolat/gomponents-lucide"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ownersEl(owners []models.Owner, editID int64, editState *ownerFormState) Node {
	return Div(ID("owners-content"), Class("mx-auto w-full max-w-7xl px-5 py-8 sm:px-8 sm:py-12"),
		Div(Class("flex flex-wrap items-end justify-between gap-4"),
			components.PageHeading("Dueños", "Administra sus datos, cabañas asignadas y comisiones."),
			Button(Type("button"), Attr("hx-get", "/admin/owners/new"), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("mb-7 inline-flex min-h-11 items-center justify-center rounded-lg bg-brand-dark px-4 py-2 font-extrabold text-white hover:brightness-90"), Text("Agregar dueño")),
		),
		Section(Class("mt-2 overflow-hidden rounded-2xl border border-line bg-white"),
			Div(Class("border-b border-line p-5 sm:px-6"), H2(Class("text-lg font-extrabold text-ink"), Text("Dueños registrados"))),
			Div(Class("overflow-x-auto"),
				Table(Class("w-full min-w-[900px] text-left text-sm"),
					THead(Class("bg-cream/70 text-xs uppercase tracking-wider text-muted"), Tr(
						Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Dueño")),
						Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Cédula")),
						Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Celular / WhatsApp")),
						Th(Scope("col"), Class("px-5 py-3 text-right font-bold"), Text("Cabañas")),
						Th(Scope("col"), Class("px-5 py-3 text-right font-bold"), Text("Comisión pagada")),
						Th(Scope("col"), Class("px-5 py-3 text-right font-bold"), Text("Pendiente")),
						Th(Scope("col"), Class("px-5 py-3 font-bold"), Text("Estado")),
						Th(Scope("col"), Class("px-5 py-3 text-right font-bold"), Text("Acciones")),
					)),
					TBody(ID("owner-table-body"), Class("divide-y divide-line"),
						If(len(owners) == 0, Tr(Td(Attr("colspan", "8"), Class("px-5 py-12 text-center text-sm text-muted"), Text("Todavía no hay dueños registrados.")))),
						Map(owners, func(owner models.Owner) Node {
							if owner.ID == editID {
								return ownerEditRowEl(owner, editState)
							}
							return ownerRowEl(owner)
						}),
					),
				),
			),
		),
	)
}

func ownerCreateFormEl(state *ownerFormState) Node {
	return Form(ID("owner-create-form"), Method("post"), Action("/admin/owners"), Attr("hx-post", "/admin/owners"), Attr("hx-target", "#owners-content"), Attr("hx-swap", "outerHTML"),
		Attr("autocomplete", "off"), Class("mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4"),
		ownerNameInput(models.Owner{}, state, true),
		ownerCedulaInput(models.Owner{}, state, true),
		ownerPhoneInput(models.Owner{}, state, true),
		ownerPINInput(models.Owner{}, state, true),
		Div(Class("flex items-end"), Button(Type("submit"), Class("inline-flex min-h-11 w-full items-center justify-center rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white transition hover:brightness-90"), Text("Crear dueño"))),
	)
}

func ownerCreateDialogEl(state *ownerFormState, oob bool) Node {
	oobMode := ""
	if oob {
		oobMode = "outerHTML"
	}
	return components.CustomDialog(components.DialogProps{Label: "Agregar dueño", SwapOOB: oobMode, Yield: Div(Class("mx-auto w-full max-w-5xl p-4 sm:p-7"),
		Div(Class("flex items-start justify-between gap-4"),
			components.PageHeading("Agregar dueño", "Registra sus datos para asignarle cabañas."),
			Button(Type("button"), Attr("data-action", "click->modal#close"), Attr("aria-label", "Cerrar"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"), lucide.X(Class("size-5"), Attr("aria-hidden", "true"))),
		),
		P(Class("mt-2 text-sm text-muted"), Text("El dueño entra con cédula y PIN. Comparte el PIN directamente; no se volverá a mostrar.")),
		ownerCreateFormEl(state),
	)})
}

func ownerRowEl(owner models.Owner) Node {
	return Tr(ID(ownerRowID(owner.ID)),
		Td(Class("px-5 py-4 align-top"),
			P(Class("font-bold text-ink"), Text(owner.Name)),
			If(owner.Cedula == "", P(Class("mt-1 text-xs font-bold text-brand-dark"), Text("Completar cédula"))),
			If(!owner.HasPIN, P(Class("mt-1 text-xs font-bold text-brand-dark"), Text("Completar PIN"))),
		),
		Td(Class("px-5 py-4 align-top text-muted"), Text(owner.Cedula)),
		Td(Class("px-5 py-4 align-top text-muted"), Text(owner.WhatsApp)),
		Td(Class("px-5 py-4 text-right align-top font-bold text-ink"), Text(strconv.Itoa(owner.CabinCount))),
		Td(Class("px-5 py-4 text-right align-top font-bold text-ink"), Text("$"+helpers.FormatCOP(owner.CommissionPaid))),
		Td(Class("px-5 py-4 text-right align-top font-bold text-ink"), Text("$"+helpers.FormatCOP(owner.CommissionDue))),
		Td(Class("px-5 py-4 align-top"), ownerStatusEl(owner.Active)),
		Td(Class("px-5 py-4 text-right align-top"),
			Div(Class("inline-flex items-center gap-2"),
				Button(Type("button"), Attr("hx-get", ownerEditURL(owner.ID)), Attr("hx-target", "#owners-content"), Attr("hx-swap", "outerHTML"), Class("min-h-10 rounded-lg border border-line px-3 py-2 text-sm font-bold text-brand-dark hover:border-brand"), Text("Editar")),
				ownerActiveFormEl(owner),
			),
		),
	)
}

func ownerEditRowEl(owner models.Owner, state *ownerFormState) Node {
	return Tr(ID(ownerRowID(owner.ID)), Td(Attr("colspan", "8"),
		Form(Method("post"), Action("/admin/owners/"+strconv.FormatInt(owner.ID, 10)), Attr("hx-post", "/admin/owners/"+strconv.FormatInt(owner.ID, 10)), Attr("hx-target", "#owners-content"), Attr("hx-swap", "outerHTML"), Attr("autocomplete", "off"), Class("grid gap-4 bg-cream/30 p-5 sm:grid-cols-2 lg:grid-cols-4"),
			ownerNameInput(owner, state, false),
			ownerCedulaInput(owner, state, false),
			ownerPhoneInput(owner, state, false),
			ownerPINInput(owner, state, !owner.HasPIN),
			Div(Class("flex items-end gap-2 sm:col-span-2 lg:col-span-4"),
				Button(Type("submit"), Class("inline-flex min-h-10 items-center rounded-lg bg-brand-dark px-4 py-2 text-sm font-extrabold text-white hover:brightness-90"), Text("Guardar")),
				Button(Type("button"), Attr("hx-get", "/admin/owners"), Attr("hx-target", "#owners-content"), Attr("hx-swap", "outerHTML"), Class("inline-flex min-h-10 items-center rounded-lg border border-line bg-white px-4 py-2 text-sm font-bold text-ink hover:border-brand"), Text("Cancelar")),
			),
		),
	))
}

func ownerStatusEl(active bool) Node {
	label, style := "Inactivo", "bg-cream text-muted"
	if active {
		label, style = "Activo", "bg-forest-soft text-forest"
	}
	return Span(Class("inline-flex rounded-full px-2.5 py-1 text-xs font-bold "+style), Text(label))
}

func ownerActiveFormEl(owner models.Owner) Node {
	label, activeValue := "Activar", "1"
	if owner.Active {
		label, activeValue = "Desactivar", "0"
	}
	return Form(Method("post"), Action("/admin/owners/"+strconv.FormatInt(owner.ID, 10)+"/active"), Attr("hx-post", "/admin/owners/"+strconv.FormatInt(owner.ID, 10)+"/active"), Attr("hx-target", "#owners-content"), Attr("hx-swap", "outerHTML"),
		If(owner.Active, Attr("hx-confirm", "¿Desactivar dueño? Sus cabañas dejarán de mostrarse y perderá acceso, pero sus reservas y comisiones se conservarán.")),
		Input(Type("hidden"), Name("active"), Value(activeValue)),
		Button(Type("submit"), Class("min-h-10 rounded-lg border border-line px-3 py-2 text-sm font-bold text-ink hover:border-brand"), Text(label)),
	)
}

func ownerNameInput(owner models.Owner, state *ownerFormState, required bool) Node {
	value := owner.Name
	if state != nil && state.Values != nil {
		value = state.Values.Get("name")
	}
	err := ownerFieldError(state, "name")
	return components.TextFieldWithError(ownerInputID(owner, "name"), "name", "Nombre", value, "Nombre completo", "text", required, err)
}

func ownerCedulaInput(owner models.Owner, state *ownerFormState, required bool) Node {
	value := owner.Cedula
	if state != nil && state.Values != nil {
		value = state.Values.Get("cedula")
	}
	return components.TextFieldWithError(ownerInputID(owner, "cedula"), "cedula", "Cédula", value, "Número de cédula", "text", required, ownerFieldError(state, "cedula"))
}

func ownerPhoneInput(owner models.Owner, state *ownerFormState, required bool) Node {
	value := owner.WhatsApp
	if state != nil && state.Values != nil {
		value = state.Values.Get("whatsapp")
	}
	return components.TextFieldWithError(ownerInputID(owner, "phone"), "whatsapp", "Celular / WhatsApp", value, "573001234567", "tel", required, ownerFieldError(state, "whatsapp"))
}

func ownerPINInput(owner models.Owner, state *ownerFormState, required bool) Node {
	pinID := ownerInputID(owner, "pin")
	err := ownerFieldError(state, "pin")
	describedBy := pinID + "-help"
	if err != "" {
		describedBy = pinID + "-error " + describedBy
	}
	label, help := "Cambiar PIN (opcional)", "Déjalo vacío para conservar el PIN actual."
	if required {
		label, help = "PIN de seis dígitos", "El dueño usará este PIN con su cédula."
	}
	return Div(Class("min-w-0"),
		components.FieldLabel(pinID, label),
		Input(ID(pinID), Name("pin"), Type("password"), InputMode("numeric"), Pattern("[0-9]{6}"), MinLength("6"), MaxLength("6"), If(required, Required()), AutoComplete("new-password"), Attr("aria-describedby", describedBy), If(err != "", Attr("aria-invalid", "true")), Class("mt-1 min-h-11 w-full rounded-xl border border-line bg-white px-4 py-2.5 text-base text-ink focus:border-brand focus:ring-2 focus:ring-brand/20")),
		P(ID(pinID+"-help"), Class("mt-1 text-xs text-muted"), Text(help)),
		components.FieldError(pinID, err),
	)
}

func ownerFieldError(state *ownerFormState, field string) string {
	if state == nil || state.Errors == nil {
		return ""
	}
	return state.Errors.ErrorStringFor(field)
}

func ownerInputID(owner models.Owner, field string) string {
	if owner.ID == 0 {
		return "new-owner-" + field
	}
	return "owner-" + strconv.FormatInt(owner.ID, 10) + "-" + field
}

func ownerRowID(id int64) string   { return "owner-row-" + strconv.FormatInt(id, 10) }
func ownerEditURL(id int64) string { return "/admin/owners/" + strconv.FormatInt(id, 10) + "/edit" }
