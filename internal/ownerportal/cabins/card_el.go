package cabins

import (
	"travelando/internal/system/components"
	"travelando/internal/system/models"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

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
		Action: Button(Type("button"), hx.Get(EditCabinURL(cabin.ID, page)), hx.Target("body"), hx.Swap("beforeend"), Class("inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg border border-line px-3 py-2 text-sm font-bold text-brand-dark transition hover:border-brand hover:bg-cream focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
			lucide.Pencil(Class("size-4"), Attr("aria-hidden", "true")), Text("Editar cabaña")),
	})
}
