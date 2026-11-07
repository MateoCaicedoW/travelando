package reservations

import (
	"travelando/internal/system/components"
	"travelando/internal/system/models"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ListPageEl(inquiries []models.Inquiry) Node {
	return Div(Class("mx-auto w-full max-w-7xl px-4 py-4 sm:px-6 sm:py-6"),
		Div(Class("flex flex-col sm:flex-row sm:items-start sm:justify-between sm:gap-x-4"),
			components.PageHeading("Reservas de mis cabañas", "Consulta solicitudes pendientes, reservas confirmadas y cancelaciones."),
			Button(Type("button"), hx.Get("/owner/reservations/new"), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-brand-dark px-4 py-2 font-extrabold text-white transition hover:brightness-90"), lucide.Plus(Class("size-4"), Attr("aria-hidden", "true")), Text("Agregar reserva")),
		),
		ListContentEl(inquiries),
	)
}

func ListContentEl(inquiries []models.Inquiry) Node {
	return Div(ID("owner-reservations-content"), Class("reservation-list-content overflow-hidden rounded-2xl border border-line bg-white"),
		If(len(inquiries) == 0, Div(Attr("role", "status"), Class("px-5 py-8 text-center"),
			P(Class("text-lg font-extrabold text-ink"), Text("Aún no tienes reservas")),
			P(Class("mt-1 text-sm text-muted"), Text("Las solicitudes y reservas aparecerán aquí.")),
		)),
		If(len(inquiries) > 0, Div(Class("reservation-actions-scroll overflow-x-auto"),
			Table(Attr("role", "table"), Class("responsive-data-table w-full min-w-[900px] text-left text-sm"),
				THead(Class("bg-cream/70 text-label uppercase tracking-wider text-muted"), Tr(
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Huésped")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Cabaña")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Fechas y estadía")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Total / estimado")),
					Th(Scope("col"), Class("px-4 py-2.5 font-bold"), Text("Estado")),
					Th(Scope("col"), Class("px-4 py-2.5 text-right font-bold"), Text("Acciones")),
				)),
				TBody(Class("divide-y divide-line"), Map(inquiries, ReservationRowEl)),
			),
		)),
	)
}
