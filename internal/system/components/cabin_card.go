package components

import (
	"strconv"

	lucide "github.com/eduardolat/gomponents-lucide"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type CabinCardOptions struct {
	Badge   Node
	Details Node
	Price   Node
	Action  Node
	Compact bool
}

func CabinCard(cabin models.Cabin, options CabinCardOptions) Node {
	extraGuest := ""
	if cabin.ExtraGuestPrice > 0 {
		extraGuest = "+$" + helpers.FormatCOP(cabin.ExtraGuestPrice) + " por huésped adicional"
	}
	photoAspect := "aspect-[4/3]"
	if options.Compact {
		photoAspect = "aspect-[16/10]"
	}
	return Article(Class("group flex h-full flex-col overflow-hidden rounded-2xl border border-line bg-white transition hover:-translate-y-0.5 hover:shadow-md"),
		cabinPhotoCarousel(cabin.Name, cabin.Photos, photoAspect),
		Div(Class("flex flex-1 flex-col p-4 sm:p-5"),
			Div(Class("flex items-start justify-between gap-3"),
				Div(Class("min-w-0"),
					P(Class("truncate text-label font-extrabold uppercase tracking-wider text-brand-dark"), Text(cabin.Location)),
					H2(Class("mt-1 truncate text-xl font-extrabold text-ink"), Text(cabin.Name)),
				),
				If(options.Badge != nil, Div(Class("shrink-0"), options.Badge)),
			),
			If(!options.Compact && cabin.Description != "", P(Class("mt-3 line-clamp-2 min-h-11 text-sm leading-6 text-muted"), Text(cabin.Description))),
			If(!options.Compact && len(cabin.Amenities) > 0, Div(Class("mt-3"), cabinAmenitySummary(cabin.Amenities))),
			Div(Class("mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-semibold text-muted"),
				Span(Class("inline-flex items-center gap-1.5"), lucide.Users(Class("size-4"), Attr("aria-hidden", "true")), Textf("%d huéspedes incluidos", cabin.IncludedGuests)),
				If(extraGuest != "", Span(Text(extraGuest))),
			),
			If(options.Details != nil, Div(Class("mt-3"), options.Details)),
			Div(Class("mt-auto pt-4"),
				If(options.Price != nil, Div(Class("border-t border-line pt-4"), options.Price)),
				If(options.Action != nil, Div(Class("mt-3"), options.Action)),
			),
		),
	)
}

func CabinPrices(cabin models.Cabin) Node {
	prices := make([]Node, 0, 2)
	if cabin.AcceptsDayUse {
		prices = append(prices, Div(P(Class("text-xs font-semibold text-muted"), Text("Pasadía")), P(Class("mt-1 text-lg font-extrabold text-ink"), Text("$"+helpers.FormatCOP(cabin.DayPrice)))))
	}
	if cabin.AcceptsOvernight {
		prices = append(prices, Div(P(Class("text-xs font-semibold text-muted"), Text("Por noche")), P(Class("mt-1 text-lg font-extrabold text-ink"), Text("$"+helpers.FormatCOP(cabin.NightPrice)))))
	}
	if len(prices) == 0 {
		return nil
	}
	class := "grid grid-cols-1 gap-3"
	if len(prices) > 1 {
		class = "grid grid-cols-2 gap-3"
	}
	return Div(Class(class), Group(prices))
}

func CabinPhotoCarousel(name string, photos []models.CabinPhoto) Node {
	return cabinPhotoCarousel(name, photos, "aspect-[4/3]")
}

// CabinPhotoCarouselCover fills the parent column height on desktop (keeps aspect on mobile).
func CabinPhotoCarouselCover(name string, photos []models.CabinPhoto) Node {
	return cabinPhotoCarousel(name, photos, "aspect-[4/3] lg:aspect-auto lg:h-full")
}

func cabinPhotoCarousel(name string, photos []models.CabinPhoto, aspect string) Node {
	if len(photos) == 0 {
		return Div(Role("img"), Attr("aria-label", "Fotos de "+name), Class("grid "+aspect+" place-items-center bg-cream text-muted"),
			Div(Class("text-center"), lucide.Image(Class("mx-auto size-8"), Attr("aria-hidden", "true")), P(Class("mt-2 text-sm font-semibold"), Text("Fotos próximamente"))),
		)
	}
	slides := make([]Node, len(photos))
	for i, photo := range photos {
		slides[i] = Div(Role("group"), Attr("aria-label", "Foto "+strconv.Itoa(i+1)+" de "+strconv.Itoa(len(photos))), Class("h-full w-full shrink-0 snap-center"),
			Button(Type("button"), Attr("data-action", "click->photo-carousel#open"), Attr("data-photo-carousel-index", strconv.Itoa(i)), Attr("aria-label", "Abrir foto "+strconv.Itoa(i+1)+" de "+strconv.Itoa(len(photos))), Class("h-full w-full cursor-zoom-in p-0 focus-visible:outline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-brand-dark"),
				Img(Src(photo.URL), Alt("Vista de "+name), Attr("loading", "lazy"), Class("h-full w-full object-cover")),
			),
		)
	}
	return Div(Role("region"), Attr("aria-label", "Fotos de "+name), Attr("data-controller", "photo-carousel"), Attr("data-photo-carousel", ""), Class("relative overflow-hidden bg-cream "+aspect),
		Div(Role("group"), Attr("aria-label", "Galería de fotos"), Attr("data-photo-carousel-target", "track"), Attr("data-carousel-track", ""), TabIndex("0"), Class("flex h-full snap-x snap-mandatory overflow-x-auto overscroll-x-contain scroll-smooth motion-reduce:scroll-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"),
			Group(slides),
		),
		If(len(photos) > 1, Group{
			Button(Type("button"), Attr("data-action", "click->photo-carousel#scroll"), Attr("data-carousel-direction", "-1"), Attr("aria-label", "Foto anterior"), Class("absolute left-3 top-1/2 grid size-10 -translate-y-1/2 place-items-center rounded-full border border-line/80 bg-white/90 text-ink shadow-sm transition hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), lucide.ChevronLeft(Class("size-5"), Attr("aria-hidden", "true"))),
			Button(Type("button"), Attr("data-action", "click->photo-carousel#scroll"), Attr("data-carousel-direction", "1"), Attr("aria-label", "Foto siguiente"), Class("absolute right-3 top-1/2 grid size-10 -translate-y-1/2 place-items-center rounded-full border border-line/80 bg-white/90 text-ink shadow-sm transition hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), lucide.ChevronRight(Class("size-5"), Attr("aria-hidden", "true"))),
		}),
	)
}

func cabinAmenitySummary(amenities []models.CabinAmenity) Node {
	items := make([]Node, 0, len(amenities))
	for _, amenity := range amenities {
		label := amenity.Name
		switch amenity.ValueType {
		case models.AmenityBoolean:
			if !amenity.BooleanValue {
				continue
			}
		case models.AmenityQuantity:
			if amenity.QuantityValue <= 0 {
				continue
			}
			label += " · " + strconv.Itoa(amenity.QuantityValue)
		default:
			continue
		}
		items = append(items, Span(Class("rounded-full bg-cream px-2.5 py-1 text-xs font-semibold text-brand-dark"), Text(label)))
	}
	return If(len(items) > 0, Div(Attr("aria-label", "Amenidades"), Class("flex flex-wrap gap-2"), Group(items)))
}

func CabinSchedule(stayType models.StayType, cabin models.Cabin) string {
	startTime, endTime := cabin.DayStart, cabin.DayEnd
	if stayType == models.StayOvernight {
		startTime, endTime = cabin.CheckIn, cabin.CheckOut
	}
	return helpers.ScheduleLabel(stayType, startTime, endTime)
}
