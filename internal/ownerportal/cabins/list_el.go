package cabins

import (
	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ListContentEl(result services.CabinPageResult) Node {
	return Div(ID("owner-content"), Class("mx-auto w-full max-w-7xl px-4 py-4 sm:px-6 sm:py-6"),
		Div(Class("flex flex-col sm:flex-row sm:items-start sm:justify-between sm:gap-x-4"),
			components.PageHeading("Mis cabañas", "Administra los datos y publicación de tus cabañas."),
			Button(Type("button"), hx.Get(ownerNewCabinURL(result.Page)), hx.Target("body"), hx.Swap("beforeend"), hx.DisabledElt("this"), Class("inline-flex min-h-10 items-center rounded-lg bg-brand-dark px-4 py-2 text-sm font-extrabold text-white hover:brightness-90"), Text("Agregar cabaña")),
		),
		If(len(result.Cabins) == 0, Div(Class("mt-4 rounded-2xl border border-line bg-white px-5 py-8 text-center text-sm text-muted"), Text("Todavía no tienes cabañas. Agrega una para publicarla."))),
		If(len(result.Cabins) > 0, Div(ID("owner-cabin-cards"), Class("grid gap-5 sm:grid-cols-2 xl:grid-cols-3"),
			Map(result.Cabins, func(cabin models.Cabin) Node { return ownerCabinCardEl(cabin, result.Page) }),
		)),
		If(result.TotalPages > 1, Nav(Attr("aria-label", "Paginación de mis cabañas"), Class("grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2 border-t border-line pt-4 sm:gap-4"),
			If(result.Page > 1, A(Href(ownerCabinsURL(result.Page-1)), hx.Get(ownerCabinsURL(result.Page-1)), hx.Target("#owner-content"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("inline-flex min-h-10 items-center justify-center rounded-lg border border-line px-3 text-sm font-bold text-ink hover:border-brand"), Text("← Anterior"))),
			P(Class("justify-self-center rounded-full bg-cream px-3 py-2 text-center text-xs font-extrabold text-brand-dark"), Textf("Página %d de %d", result.Page, result.TotalPages)),
			If(result.Page < result.TotalPages, A(Href(ownerCabinsURL(result.Page+1)), hx.Get(ownerCabinsURL(result.Page+1)), hx.Target("#owner-content"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("col-start-3 inline-flex min-h-10 items-center justify-center rounded-lg border border-line px-3 text-sm font-bold text-ink hover:border-brand"), Text("Siguiente →"))),
		)),
	)
}
