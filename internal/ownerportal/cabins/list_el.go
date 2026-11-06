package cabins

import (
	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ownerCabinsContentEl(result services.CabinPageResult) Node {
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

func ownerCabinCardEl(cabin models.Cabin, page int) Node {
	status, statusStyle := "Publicada", "bg-forest-soft text-forest"
	statusDot := "bg-forest"
	if !cabin.Active {
		status, statusStyle = "Inactiva", "bg-cream text-muted"
		statusDot = "bg-muted"
	}
	return components.CabinCard(cabin, components.CabinCardOptions{
		Badge: Span(Class("inline-flex items-center gap-2 rounded-full bg-white px-3 py-1.5 text-xs font-bold shadow-sm "+statusStyle),
			Span(Attr("aria-hidden", "true"), Class("size-2 rounded-full "+statusDot)), Text(status)),
		Price: components.CabinPrices(cabin),
		Action: Button(Type("button"), hx.Get(ownerEditCabinURL(cabin.ID, page)), hx.Target("body"), hx.Swap("beforeend"), Class("inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg border border-line px-3 py-2 text-sm font-bold text-brand-dark transition hover:border-brand hover:bg-cream focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
			lucide.Pencil(Class("size-4"), Attr("aria-hidden", "true")), Text("Editar cabaña")),
	})
}
