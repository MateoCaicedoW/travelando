package cabins

import (
	"net/url"
	"strconv"

	lucide "github.com/eduardolat/gomponents-lucide"
	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func listEl(page services.CabinPageResult, search string) Node {
	return Div(Class("mx-auto w-full max-w-7xl px-5 py-8 sm:px-8 sm:py-12"),
		Div(Class("flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between"),
			components.PageHeading("Cabañas", "Administra información, tarifas, horarios y disponibilidad."),
			Button(Type("button"), Attr("hx-get", "/admin/cabins/new"), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("mb-7 inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-brand-dark px-4 py-2 font-extrabold text-white transition hover:brightness-90"), lucide.Plus(Class("size-4"), Attr("aria-hidden", "true")), Text("Agregar cabaña")),
		),
		Div(Class("mt-6 w-full max-w-md"),
			components.FieldLabel("q", "Buscar cabañas"),
			Input(ID("q"), Name("q"), Type("search"), Value(search), Attr("maxlength", "200"), Attr("hx-get", "/admin/cabins"), Attr("hx-trigger", "keyup changed delay:300ms"), Attr("hx-target", "#cabin-list-results"), Attr("hx-swap", "outerHTML"), Placeholder("Nombre, ubicación, contacto o WhatsApp"), Class("mt-1 h-10 w-full rounded-lg bg-white px-3 py-2 text-sm text-ink")),
		),
		cabinListResultsEl(page, search),
	)
}

func cabinListResultsEl(page services.CabinPageResult, search string) Node {
	return Div(ID("cabin-list-results"),
		P(Attr("aria-live", "polite"), Class("mt-5 text-sm text-muted"), Textf("%d cabañas encontradas", page.Total)),
		If(page.Total == 0, Div(Class("mt-3 rounded-2xl border border-dashed border-line bg-white px-5 py-12 text-center"),
			If(search == "", P(Class("text-lg font-extrabold text-ink"), Text("Empieza agregando una cabaña"))),
			If(search == "", P(Class("mt-1 text-sm text-muted"), Text("Agrega la cabaña con sus datos de contacto y tarifas."))),
			If(search != "", P(Class("text-lg font-extrabold text-ink"), Text("No encontramos cabañas"))),
			If(search != "", P(Class("mt-1 text-sm text-muted"), Text("Prueba con otro nombre, ubicación o contacto."))),
		)),
		Div(Class("mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3"), Map(page.Cabins, func(cabin models.Cabin) Node {
			status := "Publicada"
			ownerContact := cabin.OwnerName + " · WhatsApp " + cabin.OwnerWhatsApp
			if cabin.OwnerID == 0 {
				status, ownerContact = "Sin dueño", "Sin dueño asignado"
			} else if !cabin.OwnerActive {
				status = "Dueño inactivo"
			}
			if !cabin.Active {
				status = "Oculta"
			}
			editURL := "/admin/cabins/" + strconv.FormatInt(cabin.ID, 10) + "/edit"
			return components.CabinCard(cabin, components.CabinCardOptions{
				Badge:   Span(Class("shrink-0 rounded-full bg-cream px-2.5 py-1 text-xs font-bold text-brand-dark"), Text(status)),
				Details: P(Class("text-sm text-muted"), Text(ownerContact)),
				Price:   components.CabinPrices(cabin),
				Action:  A(Href(editURL), Attr("hx-get", editURL), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg border border-line px-4 py-2 text-sm font-bold text-ink hover:border-brand hover:text-brand-dark"), lucide.Pencil(Class("size-4"), Attr("aria-hidden", "true")), Text("Editar cabaña")),
			})
		})),
		If(page.TotalPages > 1, Nav(Attr("aria-label", "Paginación de cabañas"), Class("mt-8 grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2 border-t border-line pt-6 sm:gap-4"),
			If(page.Page > 1, A(Href(cabinListPageURL(search, page.Page-1)), Attr("hx-get", cabinListPageURL(search, page.Page-1)), Attr("hx-target", "#cabin-list-results"), Attr("hx-swap", "outerHTML"), Class("inline-flex min-h-11 w-full items-center justify-center rounded-lg border border-line bg-white px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark sm:w-fit sm:px-4"), Text("← Anterior"))),
			P(Class("col-start-2 justify-self-center rounded-full bg-cream px-3 py-2 text-center text-xs font-extrabold text-brand-dark"), Textf("Página %d de %d", page.Page, page.TotalPages)),
			If(page.Page < page.TotalPages, A(Href(cabinListPageURL(search, page.Page+1)), Attr("hx-get", cabinListPageURL(search, page.Page+1)), Attr("hx-target", "#cabin-list-results"), Attr("hx-swap", "outerHTML"), Class("col-start-3 inline-flex min-h-11 w-full items-center justify-center rounded-lg border border-line bg-white px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark sm:w-fit sm:justify-self-end sm:px-4"), Text("Siguiente →"))),
		)),
	)
}

func cabinListPageURL(search string, page int) string {
	query := url.Values{"page": {strconv.Itoa(page)}}
	if search != "" {
		query.Set("q", search)
	}
	return "/admin/cabins?" + query.Encode()
}
