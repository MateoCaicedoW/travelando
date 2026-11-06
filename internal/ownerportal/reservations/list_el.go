package reservations

import (
	"travelando/internal/system/components"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ListPageEl(inquiries []models.Inquiry) Node {
	return Div(Class("mx-auto w-full max-w-7xl px-4 py-4 sm:px-6 sm:py-6"),
		components.PageHeading("Reservas de mis cabañas", "Consulta solicitudes pendientes, reservas confirmadas y cancelaciones."),
		ListContentEl(inquiries),
	)
}

func ListContentEl(inquiries []models.Inquiry) Node {
	return Div(ID("owner-reservations-content"), Class("reservation-list-content overflow-hidden rounded-2xl border border-line bg-white"),
		Div(Class("flex flex-wrap items-center justify-between gap-3 border-b border-line p-4 sm:px-5"),
			H2(Class("text-xl font-extrabold text-ink"), Text("Todas las reservas")),
			P(Class("text-sm text-muted"), Textf("%d registros", len(inquiries))),
		),
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
