package dashboard

import (
	"fmt"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func dashboardEl(inquiries []models.Inquiry, commissionBP int, commissionState *commissionFormState, cabinPage int) Node {
	commissionValue := helpers.FormatCommissionRate(commissionBP)
	commissionError := ""
	if commissionState != nil {
		commissionValue = commissionState.Value
		commissionError = commissionState.Error
	}
	pending, reserved := 0, 0
	commissionPaid, commissionDue := int64(0), int64(0)
	for _, inquiry := range inquiries {
		if inquiry.Status == models.InquiryPending {
			pending++
		}
		if inquiry.Status == models.InquiryReserved {
			reserved++
			if inquiry.CommissionPaidAt != "" {
				commissionPaid += inquiry.Commission
			} else {
				commissionDue += inquiry.Commission
			}
		}
	}
	return Div(ID("dashboard-content"), Class("mx-auto w-full max-w-7xl px-5 py-8 sm:px-8 sm:py-12"),
		components.PageHeading("Solicitudes y reservas", "Revisa cada consulta y el valor de comisión registrado."),
		Div(Class("mb-8 grid gap-3 sm:grid-cols-2 xl:grid-cols-4"),
			components.MetricCard("Solicitudes pendientes", strconv.Itoa(pending), "Por dar seguimiento", "text-brand-dark"),
			components.MetricCard("Reservas cerradas", strconv.Itoa(reserved), "Reportadas desde el enlace", "text-forest"),
			components.MetricCard("Comisión cobrada", "$"+helpers.FormatCOP(commissionPaid)+" COP", "Pagada por los dueños", "text-forest"),
			components.MetricCard("Comisión pendiente", "$"+helpers.FormatCOP(commissionDue)+" COP", "Por cobrar a los dueños", "text-brand-dark"),
		),
		Div(Class("overflow-hidden rounded-2xl border border-line bg-white"),
			Div(Class("flex flex-wrap items-end justify-between gap-3 border-b border-line p-5 sm:px-6"),
				Div(
					P(Class("text-xs font-extrabold uppercase tracking-[0.16em] text-brand-dark"), Text("Seguimiento")),
					H2(Class("mt-1 text-xl font-extrabold text-ink"), Text("Actividad reciente")),
				),
				P(Class("text-sm text-muted"), Textf("%d solicitudes", len(inquiries))),
			),
			If(len(inquiries) == 0, Div(Class("px-6 py-12 text-center"),
				P(Class("text-lg font-extrabold text-ink"), Text("Aún no hay solicitudes")),
				P(Class("mt-1 text-sm text-muted"), Text("Cuando un huésped envíe el formulario, aparecerá aquí.")),
			)),
			If(len(inquiries) > 0, Div(Class("overflow-x-auto"),
				Table(Class("w-full min-w-[760px] text-left text-sm"),
					THead(Class("bg-cream/70 text-xs uppercase tracking-wider text-muted"), Tr(
						Th(Class("px-5 py-3 font-bold"), Scope("col"), Text("Huésped")),
						Th(Class("px-5 py-3 font-bold"), Scope("col"), Text("Cabaña / contacto")),
						Th(Class("px-5 py-3 font-bold"), Scope("col"), Text("Fechas")),
						Th(Class("px-5 py-3 font-bold"), Scope("col"), Text("Estado")),
						Th(Class("px-5 py-3 text-right font-bold"), Scope("col"), Text("Comisión")),
					)),
					TBody(Class("divide-y divide-line"), Map(inquiries, inquiryRow)),
				),
			)),
		),
		Div(Class("mt-8 grid items-start gap-4 lg:grid-cols-2"),
			inventoryLoadingEl(cabinPage),
			Div(Class("rounded-2xl border border-line bg-white p-5 sm:p-6"),
				P(Class("text-xs font-extrabold uppercase tracking-[0.16em] text-brand-dark"), Text("Configuración")),
				H2(Class("mt-1 text-lg font-extrabold text-ink"), Text("Comisión global")),
				P(Class("mt-2 text-sm text-muted"), Text("Una misma tasa para todas las reservas nuevas.")),
				commissionFormEl(commissionValue, commissionError),
			),
		),
	)
}

func inventoryLoadingEl(page int) Node {
	url := "/admin/inventory"
	if page > 1 {
		url += fmt.Sprintf("?cabins_page=%d", page)
	}
	return Div(ID("inventory-section"), Attr("hx-get", url), Attr("hx-trigger", "load"), Attr("hx-swap", "outerHTML"), Attr("aria-busy", "true"), Class("rounded-2xl border border-line bg-white p-5 sm:p-6"),
		P(Class("text-xs font-extrabold uppercase tracking-[0.16em] text-brand-dark"), Text("Inventario")),
		H2(Class("mt-1 text-lg font-extrabold text-ink"), Text("Cabañas")),
		P(Attr("role", "status"), Class("mt-3 text-sm text-muted"), Text("Cargando inventario…")),
		Div(ID("inventory-skeleton"), Attr("aria-hidden", "true"), Class("mt-2 animate-pulse"), Map(make([]int, cabinsPerPage), func(int) Node {
			return Div(Class("flex items-center justify-between gap-3 border-b border-line py-3 last:border-0"),
				Div(Class("min-w-0 flex-1"),
					Div(Class("h-4 w-1/2 rounded bg-cream")),
					Div(Class("mt-2 h-3 w-1/3 rounded bg-cream")),
				),
				Div(Class("h-3 w-16 rounded bg-cream")),
			)
		})),
	)
}

func commissionFormEl(value, message string) Node {
	inputClass := "w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"
	if message != "" {
		inputClass += " border-red-300 focus:ring-red-200"
	}
	return Form(ID("commission-form"), Method("post"), Action("/admin/commission"), Attr("hx-post", "/admin/commission"), Attr("hx-target", "#commission-form"), Attr("hx-swap", "outerHTML"), Class("mt-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end"),
		Div(Class("min-w-0"),
			components.FieldLabel("commission", "Porcentaje (%)"),
			Input(ID("commission"), Name("commission"), Type("number"), Min("0"), Max("100"), Step("0.01"), Value(value), Required(), If(message != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "commission-error")}), Class("min-h-12 "+inputClass)),
			components.FieldError("commission", message),
		),
		Button(Type("submit"), Class("min-h-12 rounded-lg bg-brand-dark px-5 py-2 font-extrabold text-white transition hover:brightness-90"), Text("Guardar comisión")),
	)
}

func inventoryEl(page services.CabinPageResult) Node {
	return Div(ID("inventory-section"), Class("rounded-2xl border border-line bg-white p-5 sm:p-6"),
		P(Class("text-xs font-extrabold uppercase tracking-[0.16em] text-brand-dark"), Text("Inventario")),
		H2(Class("mt-1 text-lg font-extrabold text-ink"), Text("Cabañas")),
		If(len(page.Cabins) == 0, P(Class("mt-3 text-sm text-muted"), Text("Todavía no hay cabañas."))),
		Map(page.Cabins, func(c models.Cabin) Node {
			status := "Publicada"
			ownerContact := c.OwnerName
			if c.OwnerID == 0 {
				status, ownerContact = "Sin dueño", "Sin dueño asignado"
			} else if !c.OwnerActive {
				status = "Dueño inactivo"
			}
			if !c.Active {
				status = "Oculta"
			}
			ownerPhone := ""
			if c.OwnerWhatsApp != "" {
				ownerPhone = "WhatsApp " + c.OwnerWhatsApp + " · "
			}
			editURL := "/admin/cabins/" + strconv.FormatInt(c.ID, 10) + "/edit"
			return A(Href(editURL), Attr("aria-label", "Editar "+c.Name), Class("group -mx-2 flex items-center justify-between gap-3 border-b border-line px-2 py-3 last:border-0 hover:bg-cream/50 focus-visible:rounded-lg"),
				Div(Class("min-w-0 flex-1"),
					Div(Class("flex flex-wrap items-center gap-x-2 gap-y-1"),
						P(Class("font-bold text-ink group-hover:text-brand-dark"), Text(c.Name)),
						Span(Class("rounded-full bg-cream px-2 py-0.5 text-xs font-bold text-brand-dark"), Text(status)),
					),
					P(Class("mt-1 text-xs text-muted"), Text(c.Location+" · "+ownerContact)),
					P(Class("mt-1 text-xs text-muted"), Text(ownerPhone+strconv.Itoa(c.IncludedGuests)+" incluidos · $"+helpers.FormatCOP(c.ExtraGuestPrice)+" por adicional/reserva")),
					P(Class("mt-1 text-xs font-semibold text-ink"), Text("Pasadía $"+helpers.FormatCOP(c.DayPrice)+" · Noche $"+helpers.FormatCOP(c.NightPrice))),
				),
				Span(Class("shrink-0 text-xs font-bold text-brand-dark group-hover:underline"), Text("Editar")),
			)
		}),
		If(page.TotalPages > 1, Nav(Attr("aria-label", "Paginación del inventario"), Class("mt-4 grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2 border-t border-line pt-4 sm:gap-3"),
			If(page.Page > 1, A(Href(fmt.Sprintf("/admin/?cabins_page=%d", page.Page-1)), Attr("hx-get", fmt.Sprintf("/admin/inventory?cabins_page=%d", page.Page-1)), Attr("hx-target", "#inventory-section"), Attr("hx-swap", "outerHTML"), Class("inline-flex min-h-10 items-center justify-center rounded-lg border border-line px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark"), Text("← Anterior"))),
			P(Class("justify-self-center rounded-full bg-cream px-3 py-2 text-center text-xs font-extrabold text-brand-dark"), Textf("Página %d de %d", page.Page, page.TotalPages)),
			If(page.Page < page.TotalPages, A(Href(fmt.Sprintf("/admin/?cabins_page=%d", page.Page+1)), Attr("hx-get", fmt.Sprintf("/admin/inventory?cabins_page=%d", page.Page+1)), Attr("hx-target", "#inventory-section"), Attr("hx-swap", "outerHTML"), Class("col-start-3 inline-flex min-h-10 items-center justify-center rounded-lg border border-line px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark"), Text("Siguiente →"))),
		)),
	)
}

func commissionSuccessEl(rate int) Node {
	return Group{
		commissionFormEl(helpers.FormatCommissionRate(rate), ""),
		Div(ID("global-flash-container"), Attr("hx-swap-oob", "true"), components.FlashToast("Comisión global actualizada.", helpers.FlashSuccess)),
	}
}

func inquiryRow(inquiry models.Inquiry) Node {
	status := "Pendiente"
	statusClass := "bg-cream text-brand-dark"
	commission := "—"
	if inquiry.Status == models.InquiryReserved {
		status = "Reservada"
		statusClass = "bg-forest text-white"
		commission = "$" + helpers.FormatCOP(inquiry.Commission)
	} else if inquiry.Status == models.InquiryLost {
		status = "No concretada"
		statusClass = "bg-gray-100 text-muted"
	}
	dateText := helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo)
	return Tr(ID("inquiry-row-"+strconv.FormatInt(inquiry.ID, 10)),
		Td(Class("px-5 py-4 align-top"), P(Class("font-bold text-ink"), Text(inquiry.GuestName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.GuestPhone))),
		Td(Class("px-5 py-4 align-top"), P(Class("font-bold text-ink"), Text(inquiry.CabinName)), P(Class("mt-1 text-xs text-muted"), Text(inquiry.OwnerName))),
		Td(Class("px-5 py-4 align-top text-muted"), P(Text(dateText)), P(Class("mt-1 text-xs"), Text(helpers.ScheduleLabel(inquiry.StayType, inquiry.StartTime, inquiry.EndTime))), P(Class("mt-1 text-xs"), Text(helpers.StayTypeLabel(inquiry.StayType)+" · "+strconv.Itoa(inquiry.Guests)+" huéspedes"))),
		Td(Class("px-5 py-4 align-top"), Span(Class("inline-flex rounded-full px-2.5 py-1 text-xs font-bold "+statusClass), Text(status))),
		Td(Class("px-5 py-4 text-right align-top font-extrabold text-ink"),
			P(Text(commission)),
			If(inquiry.Status == models.InquiryReserved, P(Class("mt-1 text-xs font-normal text-muted"), Text("sobre $"+helpers.FormatCOP(inquiry.FinalTotal)))),
			If(inquiry.Status == models.InquiryReserved && inquiry.CommissionPaidAt != "", P(Class("mt-1 text-xs font-bold text-forest"), Text("Pagada"))),
			If(inquiry.Status == models.InquiryReserved && inquiry.CommissionPaidAt == "" && inquiry.Commission > 0,
				Form(Method("post"), Action("/admin/inquiries/"+strconv.FormatInt(inquiry.ID, 10)+"/commission-paid"), Attr("hx-post", "/admin/inquiries/"+strconv.FormatInt(inquiry.ID, 10)+"/commission-paid"), Attr("hx-target", "#dashboard-content"), Attr("hx-swap", "outerHTML"), Attr("hx-confirm", "¿Marcar comisión completa como pagada?"), Class("mt-2"),
					Button(Type("submit"), Class("text-xs font-bold text-brand-dark hover:underline"), Text("Marcar pagada")),
				),
			),
		),
	)
}
