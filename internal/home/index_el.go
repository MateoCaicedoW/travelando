package home

import (
	"strconv"
	"time"

	lucide "github.com/eduardolat/gomponents-lucide"
	hx "maragu.dev/gomponents-htmx"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func IndexEl(data PageData) Node {
	stayTypeError := data.SearchErrors.ErrorStringFor("stay_type")
	stayTypeClass := "flex min-h-12 cursor-pointer items-center gap-3 rounded-xl border border-line bg-white px-4 py-3 text-sm font-bold text-ink transition has-[:checked]:border-brand-dark has-[:checked]:bg-cream has-[:checked]:text-brand-dark focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-brand-dark"
	if stayTypeError != "" {
		stayTypeClass += " border-red-300"
	}
	return Div(ID("search-page"),
		Section(Class("bg-cream"),
			Div(Class("mx-auto grid max-w-7xl items-center gap-5 px-2 py-2 sm:px-8 sm:py-10 lg:grid-cols-[1.05fr_0.95fr] lg:gap-8 lg:py-12"),
				homeHero(data.FeaturedCabin),
				Div(ID("buscar"), Class("scroll-mt-6 rounded-2xl border border-line bg-white p-4 sm:p-6"),
					H2(Class("text-2xl font-extrabold text-ink sm:text-3xl"), Text("Encuentra cabañas para tu plan")),
					P(Class("mt-1 text-sm text-muted"), Text("Elige cómo viajas y cuándo quieres ir.")),
					Form(Method("get"), Action("/#search-results"), hx.Get("/"), hx.Target("#search-page"), hx.Swap("outerHTML show:#search-results:top"), hx.PushURL("true"), Class("mt-6 grid gap-4 sm:grid-cols-2"),
						FieldSet(ID("stay_type"), If(stayTypeError != "", Attr("aria-describedby", "stay_type-error")), If(stayTypeError != "", Attr("aria-invalid", "true")), Class("sm:col-span-2"),
							Legend(Class("mb-2 block text-sm font-bold text-ink"), Text("Tipo de plan")),
							Div(Class("grid grid-cols-2 gap-3"),
								Label(Class(stayTypeClass),
									Input(Type("radio"), Name("stay_type"), Value(string(models.StayDayUse)), Required(), If(data.StayType == models.StayDayUse, Checked()), Class("size-4 accent-brand-dark")),
									Text("Pasadía"),
								),
								Label(Class(stayTypeClass),
									Input(Type("radio"), Name("stay_type"), Value(string(models.StayOvernight)), Required(), If(data.StayType == models.StayOvernight, Checked()), Class("size-4 accent-brand-dark")),
									Text("Estadía"),
								),
							),
							components.FieldError("stay_type", stayTypeError),
						),
						components.TextFieldWithError("date_from", "date_from", "Llegada o fecha de uso", searchValue(data, "date_from", data.DateFrom), "", "date", true, data.SearchErrors.ErrorStringFor("date_from")),
						components.TextFieldWithError("date_to", "date_to", "Salida (solo estadía)", searchValue(data, "date_to", data.DateTo), "", "date", false, data.SearchErrors.ErrorStringFor("date_to")),
						Div(Class("sm:col-span-2"), components.NumberFieldWithError("guests", "guests", "Huéspedes", searchValue(data, "guests", strconv.Itoa(data.Guests)), "1", "", data.SearchErrors.ErrorStringFor("guests"))),
						Button(Type("submit"), Class("sm:col-span-2 inline-flex min-h-14 items-center justify-center gap-2 rounded-lg bg-brand-dark px-4 py-2.5 font-extrabold text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
							Text("Buscar cabañas disponibles"), lucide.Search(Class("size-5"), Attr("aria-hidden", "true")),
						),
					),
				),
			),
		),
		Section(ID("search-results"), Class("mx-auto w-full max-w-7xl px-2 py-2 sm:px-8 sm:py-16"),
			If(data.DidSearch && !data.SearchErrors.HasAny(), Div(Class("mb-7 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"),
				Div(
					P(Class("text-label font-extrabold uppercase tracking-[0.08em] text-brand-dark"), Text("Tu próxima escapada")),
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
			If(data.DidSearch && !data.SearchErrors.HasAny() && data.Error == "" && len(data.Cabins) == 0, Div(Class("rounded-2xl border border-dashed border-line bg-cream/60 px-5 py-8 text-center"),
				H3(Class("text-lg font-extrabold text-ink"), Text("No encontramos cabañas disponibles")),
				P(Class("mt-1 text-sm text-muted"), Text("Prueba otra fecha o ajusta la cantidad de huéspedes.")),
			)),
			If(!data.DidSearch && len(data.SampleCabins) > 0, Div(Class("mt-12"),
				Div(Class("mb-6 max-w-2xl"),
					H2(Class("text-3xl font-extrabold tracking-tight text-ink sm:text-4xl"), Text("Cabañas para empezar a explorar")),
					P(Class("mt-2 text-base leading-7 text-muted"), Text("Mira algunas opciones del catálogo. La disponibilidad se confirma al buscar tus fechas.")),
				),
				Div(Class("grid gap-5 md:grid-cols-2 xl:grid-cols-3"),
					Map(data.SampleCabins, func(cabin models.Cabin) Node {
						return components.CabinCard(cabin, components.CabinCardOptions{
							Badge:   homeCabinStayBadge(cabin),
							Action:  A(Href("/cabins/"+strconv.FormatInt(cabin.ID, 10)), Class("inline-flex min-h-11 w-full items-center justify-center whitespace-nowrap rounded-lg bg-brand-dark px-4 py-2 text-sm font-bold text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Text("Ver cabaña")),
							Compact: true,
						})
					}),
				),
			)),
			If((data.HasPrevPage || data.HasNextPage) && data.Error == "" && !data.SearchErrors.HasAny(), Nav(Attr("aria-label", "Paginación de cabañas"), Class("mt-8 grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2 border-t border-line pt-6 sm:gap-4"),
				If(data.HasPrevPage, A(Href(resultsPageLink(data, data.Page-1)), hx.Get(resultsPageLink(data, data.Page-1)), hx.Target("#search-page"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("inline-flex min-h-11 w-full items-center justify-center rounded-lg border border-line bg-white px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark sm:w-fit sm:px-4"), Text("← Anterior"))),
				P(Class("col-start-2 justify-self-center rounded-full bg-cream px-3 py-2 text-center text-xs font-extrabold text-brand-dark"), Textf("Página %d", data.Page)),
				If(data.HasNextPage, A(Href(resultsPageLink(data, data.Page+1)), hx.Get(resultsPageLink(data, data.Page+1)), hx.Target("#search-page"), hx.Swap("outerHTML"), hx.PushURL("true"), Class("col-start-3 inline-flex min-h-11 w-full items-center justify-center rounded-lg border border-line bg-white px-3 text-sm font-bold text-ink transition hover:border-brand hover:bg-cream hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark sm:w-fit sm:justify-self-end sm:px-4"), Text("Siguiente →"))),
			)),
			If(!data.DidSearch, Div(Class("mt-12 grid gap-6 border-t border-line py-8 sm:mt-16 sm:grid-cols-3 sm:py-10"),
				HowItWorksStep("01", "Elige fechas", "Busca por pasadía o estadía."),
				HowItWorksStep("02", "Compara cabañas", "Revisa precio, horarios y ubicación."),
				HowItWorksStep("03", "Habla con el contacto", "Continúa por WhatsApp para cerrar el plan."),
			)),
		),
	)
}

func homeCabinStayBadge(cabin models.Cabin) Node {
	label := "Estadía"
	if cabin.AcceptsDayUse {
		label = "Pasadía"
		if cabin.AcceptsOvernight {
			label = "Pasadía y estadía"
		}
	}
	return Span(Class("shrink-0 rounded-full bg-cream px-3 py-1 text-xs font-bold text-brand-dark"), Text(label))
}

func homeHero(cabin *models.Cabin) Node {
	if cabin == nil || len(cabin.Photos) == 0 {
		return Div(Class("flex min-h-64 items-end rounded-2xl bg-forest-soft p-6 sm:min-h-80 sm:p-8 lg:min-h-[520px]"),
			H1(Class("max-w-[12ch] text-4xl font-black leading-[1.06] tracking-[-0.04em] text-ink sm:text-5xl"),
				Text("Tu próximo plan está "), Span(Class("text-brand-dark"), Text("más cerca.")),
			),
		)
	}
	photo := cabin.Photos[0]
	caption := cabin.Name
	if cabin.Location != "" {
		caption += " · " + cabin.Location
	}
	return Figure(Class("relative isolate aspect-[3/2] overflow-hidden rounded-2xl bg-forest lg:aspect-auto lg:min-h-[520px]"),
		Img(Src(photo.URL), Alt("Fotografía de una cabaña"), Attr("fetchpriority", "high"), Class("absolute inset-0 size-full object-cover")),
		Div(Attr("aria-hidden", "true"), Class("absolute inset-0 bg-gradient-to-t from-ink/90 via-ink/25 to-transparent")),
		FigCaption(Class("absolute inset-x-0 bottom-0 p-6 text-white sm:p-8"),
			H1(Class("max-w-[12ch] text-4xl font-black leading-[1.06] tracking-[-0.04em] sm:text-5xl"),
				Text("Tu próximo plan está "), Span(Class("text-brand"), Text("más cerca.")),
			),
			P(Class("mt-4 text-sm font-semibold text-white"), Text(caption)),
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
		P(Class("text-label font-bold text-muted"), Text(label)),
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
