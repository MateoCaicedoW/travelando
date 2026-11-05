package home

import (
	"strconv"
	"time"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func IndexEl(data PageData) Node {
	stayTypeClass := "w-full bg-white px-4 py-3 text-base text-ink outline-none focus:ring-2 focus:ring-brand/20"
	if data.SearchErrors.Has("stay_type") {
		stayTypeClass += " border-red-300 focus:ring-red-200"
	}
	return Div(ID("search-page"),
		Section(Class("bg-cream"),
			Div(Class("mx-auto grid max-w-7xl items-center gap-10 px-5 py-12 sm:px-8 sm:py-16 lg:grid-cols-[1fr_0.85fr] lg:gap-16 lg:py-20"),
				Div(Class("max-w-2xl"),
					H1(Class("max-w-xl text-4xl font-black leading-[1.06] tracking-[-0.04em] text-ink sm:text-6xl"),
						Text("Tu próximo plan está "), Span(Class("text-brand-dark"), Text("más cerca.")),
					),
					P(Class("mt-5 max-w-lg text-lg leading-8 text-muted"), Text("Encuentra una cabaña que se sienta como una pausa. Compara opciones y habla directamente con su contacto.")),
					Div(Class("mt-7 flex flex-wrap gap-2 text-sm font-semibold text-forest"),
						Span(Class("rounded-full border border-line bg-white px-3 py-2"), Text("Contacto directo")),
						Span(Class("rounded-full border border-line bg-white px-3 py-2"), Text("Disponibilidad por fecha")),
					),
				),
				Div(ID("buscar"), Class("scroll-mt-6 rounded-2xl border border-line bg-white p-5 sm:p-7"),
					H2(Class("text-2xl font-extrabold text-ink"), Text("¿Cuándo quieres escapar?")),
					P(Class("mt-1 text-sm text-muted"), Text("Te mostramos solo las cabañas disponibles.")),
					Form(Method("get"), Action("/"), Attr("hx-get", "/"), Attr("hx-target", "#search-page"), Attr("hx-swap", "outerHTML"), Attr("hx-push-url", "true"), Class("mt-6 grid gap-4 sm:grid-cols-2"),
						Div(Class("sm:col-span-2"),
							components.FieldLabel("stay_type", "Tipo de plan"),
							Select(ID("stay_type"), Name("stay_type"), If(data.SearchErrors.Has("stay_type"), Group{Attr("aria-invalid", "true"), Attr("aria-describedby", "stay_type-error")}), Class(stayTypeClass),
								Option(Value(string(models.StayDayUse)), If(data.StayType == models.StayDayUse, Selected()), Text("Pasadía")),
								Option(Value(string(models.StayOvernight)), If(data.StayType == models.StayOvernight, Selected()), Text("Estadía")),
							),
							components.FieldError("stay_type", data.SearchErrors.ErrorStringFor("stay_type")),
						),
						components.TextFieldWithError("date_from", "date_from", "Llegada o fecha de uso", searchValue(data, "date_from", data.DateFrom), "", "date", true, data.SearchErrors.ErrorStringFor("date_from")),
						components.TextFieldWithError("date_to", "date_to", "Salida (solo estadía)", searchValue(data, "date_to", data.DateTo), "", "date", false, data.SearchErrors.ErrorStringFor("date_to")),
						Div(Class("sm:col-span-2"), components.NumberFieldWithError("guests", "guests", "Huéspedes", searchValue(data, "guests", strconv.Itoa(data.Guests)), "1", "", data.SearchErrors.ErrorStringFor("guests"))),
						Button(Type("submit"), Class("sm:col-span-2 inline-flex min-h-12 items-center justify-center gap-2 rounded-lg bg-brand-dark px-5 py-3 font-extrabold text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
							Text("Buscar cabañas"), Span(Attr("aria-hidden", "true"), Text("→")),
						),
					),
				),
			),
		),
		Section(Class("mx-auto w-full max-w-7xl px-5 py-12 sm:px-8 sm:py-16"),
			If(data.DidSearch && !data.SearchErrors.HasAny(), Div(Class("mb-7 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"),
				Div(
					P(Class("text-xs font-extrabold uppercase tracking-[0.18em] text-brand-dark"), Text("Tu próxima escapada")),
					H2(Class("mt-1 text-3xl font-extrabold tracking-tight text-ink sm:text-4xl"), Text("Cabañas disponibles")),
				),
				P(Class("inline-flex w-fit items-center gap-2 rounded-full border border-line bg-white px-4 py-2 text-sm font-semibold text-muted"),
					Span(Class("size-2 rounded-full bg-forest"), Attr("aria-hidden", "true")),
					Textf("%s · %s · %d huéspedes", stayLabel(data.StayType), searchDateLabel(data.DateFrom), data.Guests),
				),
			)),
			If(data.DidSearch && !data.SearchErrors.HasAny() && data.Error == "" && len(data.Cabins) > 0, Div(Class("grid gap-5 md:grid-cols-2 xl:grid-cols-3"),
				Map(data.Cabins, func(cabin models.Cabin) Node {
					stay := "Pasadía"
					if data.StayType == models.StayOvernight {
						stay = "Estadía"
					}
					return components.CabinCard(cabin, components.CabinCardOptions{
						Badge:   Span(Class("shrink-0 rounded-full bg-cream px-3 py-1 text-xs font-bold text-brand-dark"), Text(stay)),
						Price:   homeCabinCardPrice(cabin, data),
						Action:  A(Href(resultsLink(cabin.ID, data)), Target("_blank"), Rel("noopener noreferrer"), Class("inline-flex min-h-11 w-full items-center justify-center whitespace-nowrap rounded-lg bg-brand-dark px-4 py-2 text-sm font-bold text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Text("Ver cabaña")),
						Compact: true,
					})
				}),
			)),
			If(data.DidSearch && !data.SearchErrors.HasAny() && data.Error == "" && len(data.Cabins) == 0, Div(Class("rounded-2xl border border-dashed border-line bg-cream/60 px-6 py-12 text-center"),
				H3(Class("text-lg font-extrabold text-ink"), Text("No encontramos cabañas disponibles")),
				P(Class("mt-1 text-sm text-muted"), Text("Prueba otra fecha o ajusta la cantidad de huéspedes.")),
			)),
			If((data.HasPrevPage || data.HasNextPage) && data.Error == "" && !data.SearchErrors.HasAny(), Nav(Attr("aria-label", "Paginación de cabañas"), Class("mt-8 grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2 border-t border-line pt-6 sm:gap-4"),
				If(data.HasPrevPage, A(Href(resultsPageLink(data, data.Page-1)), Attr("hx-get", resultsPageLink(data, data.Page-1)), Attr("hx-target", "#search-page"), Attr("hx-swap", "outerHTML"), Attr("hx-push-url", "true"), Class("inline-flex min-h-11 w-full items-center justify-center rounded-lg border border-line bg-white px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark sm:w-fit sm:px-4"), Text("← Anterior"))),
				P(Class("col-start-2 justify-self-center rounded-full bg-cream px-3 py-2 text-center text-xs font-extrabold text-brand-dark"), Textf("Página %d", data.Page)),
				If(data.HasNextPage, A(Href(resultsPageLink(data, data.Page+1)), Attr("hx-get", resultsPageLink(data, data.Page+1)), Attr("hx-target", "#search-page"), Attr("hx-swap", "outerHTML"), Attr("hx-push-url", "true"), Class("col-start-3 inline-flex min-h-11 w-full items-center justify-center rounded-lg border border-line bg-white px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark sm:w-fit sm:justify-self-end sm:px-4"), Text("Siguiente →"))),
			)),
			If(!data.DidSearch, Div(Class("grid gap-6 rounded-2xl border border-line bg-white p-6 sm:grid-cols-3 sm:p-8"),
				HowItWorksStep("01", "Elige fechas", "Busca por pasadía o estadía."),
				HowItWorksStep("02", "Compara cabañas", "Revisa precio, horarios y ubicación."),
				HowItWorksStep("03", "Habla con el contacto", "Continúa por WhatsApp para cerrar el plan."),
			)),
		),
	)
}

func homeCabinCardPrice(cabin models.Cabin, data PageData) Node {
	prices := make([]Node, 0, 2)
	if cabin.AcceptsDayUse {
		price, valid := priceFor(cabin, models.StayDayUse, data.DateFrom, "", data.Guests, data.DepositCOP)
		unit := "Incluye depósito"
		if !valid {
			unit = "No disponible"
		}
		prices = append(prices, homeCabinPriceTile("Pasadía", unit, price, valid))
	}
	if cabin.AcceptsOvernight {
		dateTo := ""
		unit := "1 noche · incluye depósito"
		if data.StayType == models.StayOvernight {
			dateTo = data.DateTo
			if dateTo != "" {
				unit = "Estadía completa · incluye depósito"
			}
		}
		price, valid := priceFor(cabin, models.StayOvernight, data.DateFrom, dateTo, data.Guests, data.DepositCOP)
		if !valid {
			unit = "No disponible"
		}
		prices = append(prices, homeCabinPriceTile("Estadía", unit, price, valid))
	}
	class := "grid grid-cols-1 gap-2"
	if len(prices) == 2 {
		class = "grid grid-cols-2 gap-2 sm:gap-3"
	}
	return Div(Class(class), Group(prices))
}

func homeCabinPriceTile(label, unit string, price int64, valid bool) Node {
	priceLabel := "$" + helpers.FormatCOP(price)
	if !valid {
		priceLabel = "—"
	}
	return Div(Class("min-w-0 rounded-xl bg-cream/70 px-3 py-2.5"),
		P(Class("text-xs font-bold text-muted"), Text(label)),
		P(Class("mt-1 break-words text-base font-extrabold leading-tight text-ink sm:text-lg"), Text(priceLabel)),
		P(Class("mt-0.5 text-[11px] font-medium text-muted"), Text(unit)),
	)
}

func searchDateLabel(value string) string {
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return value
	}
	return date.Format("02/01/2006")
}

func searchValue(data PageData, key, fallback string) string {
	if data.SearchValues != nil {
		if _, exists := data.SearchValues[key]; exists {
			return data.SearchValues.Get(key)
		}
	}
	return fallback
}

func HowItWorksStep(number, title, description string) Node {
	return Div(Class("flex gap-3"),
		Span(Class("grid size-8 shrink-0 place-items-center rounded-full bg-cream text-xs font-black text-brand-dark"), Text(number)),
		Div(
			H3(Class("font-extrabold text-ink"), Text(title)),
			P(Class("mt-1 text-sm leading-6 text-muted"), Text(description)),
		),
	)
}

func stayLabel(stayType models.StayType) string {
	if stayType == models.StayOvernight {
		return "estadía"
	}
	return "pasadía"
}
