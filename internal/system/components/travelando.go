package components

import (
	"fmt"
	"strconv"
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Brand() Node {
	return brandAt("/")
}

func brandAt(href string) Node {
	return A(Href(href), Class("inline-flex items-center gap-2.5 font-extrabold tracking-tight text-ink"),
		Span(Class("grid size-9 place-items-center rounded-xl bg-brand text-lg font-black text-ink sm:size-10"), Text("T")),
		Span(Class("text-xl"), Text("Travelando")),
	)
}

func SiteHeader() Node {
	return Header(Class("border-b border-line bg-cream"),
		Div(Class("mx-auto flex max-w-7xl items-center justify-between gap-4 px-5 py-3 sm:px-8 sm:py-4"),
			Brand(),
			Nav(Attr("aria-label", "Navegación principal"), Class("flex shrink-0 items-center gap-2 text-sm font-bold"),
				A(Href("/#buscar"), Class("inline-flex min-h-10 items-center gap-2 rounded-lg bg-brand px-4 text-ink transition hover:bg-brand-dark hover:text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
					lucide.Search(Class("size-4"), Attr("aria-hidden", "true")), Text("Explorar")),
			),
		),
	)
}

type portalLink struct {
	label string
	href  string
	icon  Node
}

func AdminHeader(path string) Node {
	return portalHeader("/admin/", "/admin/logout", path, "Navegación de administración",
		portalLink{"Solicitudes", "/admin/", lucide.ClipboardList(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Calendario", "/admin/calendar", lucide.CalendarDays(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Cabañas", "/admin/cabins", lucide.BedDouble(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Amenidades", "/admin/amenities", lucide.Sparkles(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Dueños", "/admin/owners", lucide.UsersRound(Class("size-4"), Attr("aria-hidden", "true"))},
	)
}

func OwnerHeader(path string) Node {
	return portalHeader("/owner/", "/owner/logout", path, "Navegación de dueño",
		portalLink{"Calendario", "/owner/", lucide.CalendarDays(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Mis cabañas", "/owner/cabins", lucide.BedDouble(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Amenidades", "/owner/amenities", lucide.Sparkles(Class("size-4"), Attr("aria-hidden", "true"))},
	)
}

func portalHeader(home, logout, path, navLabel string, links ...portalLink) Node {
	portalTitle := "Portal de dueños"
	if home == "/admin/" {
		portalTitle = "Administración"
	}
	return Header(Class("border-b border-line bg-white"),
		Div(Class("mx-auto flex w-full max-w-7xl flex-wrap items-center gap-x-6 gap-y-3 px-5 py-3 sm:px-8 lg:flex-nowrap lg:gap-x-5"),
			Div(Class("flex min-w-0 items-center gap-3"),
				brandAt(home),
				Span(Attr("aria-hidden", "true"), Class("hidden h-8 w-px bg-line sm:block")),
				Span(Class("hidden text-xs font-extrabold uppercase leading-4 tracking-[0.14em] text-muted sm:block"), Text(portalTitle)),
			),
			Nav(Attr("aria-label", navLabel), Class("order-3 grid w-full grid-cols-2 gap-1.5 border-t border-line pt-2.5 lg:order-none lg:flex lg:w-auto lg:flex-1 lg:justify-center lg:gap-1 lg:border-0 lg:pt-0"), Map(links, func(link portalLink) Node {
				active := path == link.href || link.href != home && strings.HasPrefix(path, link.href+"/")
				class := "inline-flex min-h-10 items-center justify-center gap-2 rounded-lg px-3 text-sm font-bold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark lg:justify-start"
				if active {
					class += " bg-cream text-brand-dark"
				} else {
					class += " text-muted hover:bg-cream hover:text-ink"
				}
				return A(Href(link.href), If(active, Attr("aria-current", "page")), Class(class), link.icon, Span(Text(link.label)))
			})),
			Form(Method("post"), Action(logout), Class("ml-auto"),
				Button(Type("submit"), Attr("aria-label", "Cerrar sesión"), Class("inline-flex min-h-10 shrink-0 items-center gap-2 rounded-lg border border-line bg-white px-3 text-sm font-bold text-muted transition hover:border-brand hover:bg-cream hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
					lucide.LogOut(Class("size-4"), Attr("aria-hidden", "true")), Span(Text("Salir")),
				),
			),
		),
	)
}

func SiteFooter() Node {
	return Footer(Class("border-t border-line bg-cream"),
		Div(Class("mx-auto flex max-w-7xl flex-col gap-2 px-5 py-7 text-sm text-muted sm:flex-row sm:items-center sm:justify-between sm:px-8"),
			Brand(),
			P(Text("Encuentra tu próxima escapada.")),
		),
	)
}

func PageHeading(title, description string) Node {
	return Div(Class("mb-7 max-w-2xl"),
		H1(Class("text-3xl font-extrabold tracking-tight text-ink sm:text-4xl"), Text(title)),
		P(Class("mt-3 max-w-prose text-base leading-7 text-muted"), Text(description)),
	)
}

func Notice(message string, isError bool) Node {
	class := "rounded-xl border border-forest/20 bg-forest-soft px-4 py-3 text-sm text-forest"
	if isError {
		class = "rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
	}
	return Div(Class(class), Role("status"), Text(message))
}

func CabinCard(cabin models.Cabin, price int64, stayType models.StayType, href string) Node {
	image := Div(Class("grid aspect-[4/3] place-items-center overflow-hidden bg-cream"),
		Div(Class("text-center"),
			P(Class("text-sm font-bold text-muted"), Text("Refugio natural")),
		),
	)
	if cabin.CoverImage != "" {
		image = Div(Class("aspect-[4/3] overflow-hidden bg-cream"),
			Img(Src(cabin.CoverImage), Alt("Vista de "+cabin.Name), Class("h-full w-full object-cover transition duration-500 group-hover:scale-[1.03]")),
		)
	}
	stay := "Pasadía"
	if stayType == models.StayOvernight {
		stay = "Estadía"
	}
	startTime, endTime := cabin.DayStart, cabin.DayEnd
	if stayType == models.StayOvernight {
		startTime, endTime = cabin.CheckIn, cabin.CheckOut
	}
	return Article(Class("group overflow-hidden rounded-2xl border border-line bg-white transition hover:-translate-y-0.5 hover:shadow-md"),
		image,
		Div(Class("p-5 sm:p-6"),
			Div(Class("flex items-start justify-between gap-4"),
				Div(Class("min-w-0"),
					P(Class("text-xs font-bold uppercase tracking-wider text-brand-dark"), Text(cabin.Location)),
					H2(Class("mt-1 truncate text-xl font-extrabold text-ink"), Text(cabin.Name)),
				),
				Span(Class("shrink-0 rounded-full bg-cream px-3 py-1 text-xs font-bold text-brand-dark"), Text(stay)),
			),
			P(Class("mt-3 line-clamp-2 min-h-11 text-sm leading-6 text-muted"), Text(cabin.Description)),
			cabinAmenitySummary(cabin.Amenities),
			P(Class("mt-3 text-sm font-semibold text-muted"), Text(helpers.ScheduleLabel(stayType, startTime, endTime))),
			Div(Class("mt-5 flex items-end justify-between gap-3 border-t border-line pt-4"),
				Div(
					P(Class("text-xs font-semibold text-muted"), Text(fmt.Sprint(cabin.IncludedGuests)+" incluidas · $"+helpers.FormatCOP(cabin.ExtraGuestPrice)+" extra por reserva")),
					P(Class("mt-1 text-xl font-extrabold text-ink"), Text("$"+helpers.FormatCOP(price)),
						Span(Class("ml-1 text-xs font-medium text-muted"), Text("COP / "+stayLabel(stayType))),
					),
				),
				A(Href(href), Class("inline-flex min-h-11 shrink-0 items-center justify-center whitespace-nowrap rounded-lg bg-brand-dark px-3 py-2 text-sm font-bold text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark sm:px-4"), Text("Ver cabaña")),
			),
		),
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
	return If(len(items) > 0, Div(Attr("aria-label", "Amenidades"), Class("mt-3 flex flex-wrap gap-2"), Group(items)))
}

func FieldLabel(forID, label string) Node {
	return Label(For(forID), Class("mb-1.5 block text-sm font-bold text-ink"), Text(label))
}

func TextField(id, name, label, value, placeholder, inputType string, required bool) Node {
	return TextFieldWithError(id, name, label, value, placeholder, inputType, required, "")
}

func TextFieldWithError(id, name, label, value, placeholder, inputType string, required bool, errorMessage string) Node {
	inputClass := "w-full bg-white px-4 py-3 text-base text-ink outline-none placeholder:text-muted focus:ring-2 focus:ring-brand/20"
	if errorMessage != "" {
		inputClass += " border-red-300 focus:ring-red-200"
	}
	return Div(Class("min-w-0"),
		FieldLabel(id, label),
		Input(ID(id), Name(name), Type(inputType), Value(value), Placeholder(placeholder), If(required, Required()), If(errorMessage != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", id+"-error")}), Class(inputClass)),
		FieldError(id, errorMessage),
	)
}

func FieldError(id, message string) Node {
	return If(message != "", P(ID(id+"-error"), Class("mt-1.5 rounded-md bg-red-50 px-2 py-1 text-sm font-medium text-red-800"), Attr("aria-live", "polite"), Text(message)))
}

func stayLabel(stayType models.StayType) string {
	if stayType == models.StayOvernight {
		return "noche"
	}
	return "día"
}
