package components

import (
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Brand() Node {
	return brandAt("/")
}

func brandIcon() Node {
	return Img(Src("/assets/brand-orange-cream.png"), Alt(""), Class("size-9 shrink-0 rounded-xl sm:size-10"))
}

func brandAt(href string) Node {
	return A(Href(href), Class("inline-flex items-center gap-2.5 font-extrabold tracking-tight text-ink"),
		brandIcon(),
		Span(Class("brand-wordmark text-xl text-brand"), Text("Travelando")),
	)
}

func adminBrand() Node {
	return portalBrand("/admin/")
}

func portalBrand(home string) Node {
	return A(Href(home), Attr("aria-label", "Travelando"), Class("inline-flex min-w-0 items-center gap-2.5 font-extrabold tracking-tight text-ink"),
		brandIcon(),
		Span(Attr("data-portal-navigation-target", "brandLabel"), Class("brand-wordmark text-xl text-brand"), Text("Travelando")),
	)
}

func SiteHeader(path string) Node {
	return Header(Class("border-b border-line bg-cream"),
		Div(Class("mx-auto flex max-w-7xl items-center justify-between gap-4 px-2 py-3 sm:px-8 sm:py-4"),
			Brand(),
			If(path != "/" && !strings.HasPrefix(path, "/report/"), Nav(Attr("aria-label", "Navegación principal"), Class("flex shrink-0 items-center gap-2 text-sm font-bold"),
				A(Href("/#buscar"), Class("inline-flex min-h-10 items-center gap-2 rounded-lg bg-brand-dark px-4 text-white transition hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
					lucide.Search(Class("size-4"), Attr("aria-hidden", "true")), Text("Explorar")),
			)),
		),
	)
}

type portalLink struct {
	label string
	href  string
	icon  Node
}

func AdminHeader(path string) Node {
	links := []portalLink{
		portalLink{"Inicio", "/admin/", lucide.ClipboardList(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Reservas", "/admin/reservations", lucide.BookOpenCheck(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Calendario", "/admin/calendar", lucide.CalendarDays(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Cabañas", "/admin/cabins", lucide.BedDouble(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Amenidades", "/admin/amenities", lucide.Sparkles(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Dueños", "/admin/owners", lucide.UsersRound(Class("size-4"), Attr("aria-hidden", "true"))},
	}
	return Aside(ID("portal-sidebar"), Attr("aria-label", "Navegación de administración"), Attr("data-portal-navigation-target", "sidebar"), Class("fixed inset-y-0 left-0 z-50 hidden flex h-screen w-64 shrink-0 flex-col border-r border-line bg-cream px-4 py-5 transition-[width] duration-200 lg:sticky lg:top-0 lg:flex"),
		Div(Attr("data-portal-navigation-target", "headerRow"), Class("flex items-center justify-between gap-2 border-b border-line pb-4"),
			adminBrand(),
			Div(Attr("data-portal-navigation-target", "collapseRow"), Class("hidden justify-end lg:flex"),
				Button(Type("button"), Attr("aria-label", "Contraer navegación"), Title("Contraer navegación"), Attr("data-portal-navigation-target", "collapseButton"), Attr("data-action", "click->portal-navigation#toggleCompact"), Class("inline-flex size-9 shrink-0 items-center justify-center rounded-lg text-muted hover:bg-white hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
					lucide.ChevronLeft(Class("size-4"), Attr("data-portal-navigation-target", "collapseIcon"), Attr("aria-hidden", "true")),
					lucide.ChevronRight(Class("hidden size-4"), Attr("data-portal-navigation-target", "expandIcon"), Attr("aria-hidden", "true")),
				),
			),
		),
		Nav(Attr("aria-label", "Secciones de administración"), Class("grid gap-1 pt-4"), Map(links, func(link portalLink) Node {
			active := path == link.href || link.href != "/admin/" && strings.HasPrefix(path, link.href+"/")
			class := "flex min-h-11 items-center justify-start gap-3 rounded-lg px-3 text-sm font-bold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"
			if active {
				class += " bg-white text-brand-dark"
			} else {
				class += " text-muted hover:bg-white hover:text-ink"
			}
			return A(Href(link.href), If(active, Attr("aria-current", "page")), Attr("aria-label", link.label), Title(link.label), Attr("data-portal-navigation-target", "navigationLink"), Attr("data-action", "click->portal-navigation#closeMobile"), Class(class), link.icon, Span(Attr("data-portal-navigation-target", "linkLabel"), Text(link.label)))
		})),
		Div(Class("mt-auto border-t border-line pt-4"),
			Form(Method("post"), Action("/admin/logout"),
				Button(Type("submit"), Attr("aria-label", "Cerrar sesión"), Attr("data-portal-navigation-target", "logoutButton"), Class("flex min-h-11 w-full items-center justify-start gap-3 rounded-lg px-3 text-sm font-bold text-muted transition hover:bg-white hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
					lucide.LogOut(Class("size-4 shrink-0"), Attr("aria-hidden", "true")), Span(Attr("data-portal-navigation-target", "logoutLabel"), Text("Salir")),
				),
			),
		),
	)
}

func AdminMobileBar() Node {
	return Header(Class("sticky top-0 z-30 flex min-h-16 items-center gap-3 border-b border-line bg-white px-2 lg:hidden"),
		Button(Type("button"), Attr("aria-label", "Abrir navegación"), Attr("aria-controls", "portal-sidebar"), Attr("aria-expanded", "false"), Attr("data-portal-navigation-target", "mobileButton"), Attr("data-action", "click->portal-navigation#toggleMobile"), Class("inline-flex size-10 shrink-0 items-center justify-center rounded-lg border border-line text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
			lucide.Menu(Class("size-5"), Attr("aria-hidden", "true")),
		),
		adminBrand(),
	)
}

func PortalNavigationBackdrop() Node {
	return Button(Type("button"), Attr("aria-label", "Cerrar navegación"), Attr("data-portal-navigation-target", "backdrop"), Attr("data-action", "click->portal-navigation#closeMobile"), Class("fixed inset-0 z-40 hidden bg-ink/40 lg:hidden"))
}

func OwnerMobileBar() Node {
	return Header(Class("sticky top-0 z-30 flex min-h-16 items-center gap-3 border-b border-line bg-white px-2 lg:hidden"),
		Button(Type("button"), Attr("aria-label", "Abrir navegación"), Attr("aria-controls", "portal-sidebar"), Attr("aria-expanded", "false"), Attr("data-portal-navigation-target", "mobileButton"), Attr("data-action", "click->portal-navigation#toggleMobile"), Class("inline-flex size-10 shrink-0 items-center justify-center rounded-lg border border-line text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
			lucide.Menu(Class("size-5"), Attr("aria-hidden", "true")),
		),
		portalBrand("/owner/"),
	)
}

func OwnerNavigation(path string) Node {
	links := []portalLink{
		portalLink{"Calendario", "/owner/", lucide.CalendarDays(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Reservas", "/owner/reservations", lucide.BookOpenCheck(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Mis cabañas", "/owner/cabins", lucide.BedDouble(Class("size-4"), Attr("aria-hidden", "true"))},
		portalLink{"Amenidades", "/owner/amenities", lucide.Sparkles(Class("size-4"), Attr("aria-hidden", "true"))},
	}
	profileClass := "inline-flex size-11 shrink-0 items-center justify-center rounded-full bg-forest text-white transition hover:bg-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"
	if path == "/owner/profile" {
		profileClass += " ring-2 ring-brand ring-offset-2 ring-offset-cream"
	}
	return Aside(ID("portal-sidebar"), Attr("aria-label", "Navegación de dueño"), Attr("data-portal-navigation-target", "sidebar"), Class("fixed inset-y-0 left-0 z-50 hidden flex h-screen w-64 shrink-0 flex-col border-r border-line bg-cream px-4 py-5 transition-[width] duration-200 lg:sticky lg:top-0 lg:flex"),
		Div(Attr("data-portal-navigation-target", "headerRow"), Class("flex items-center justify-between gap-2 border-b border-line pb-4"),
			portalBrand("/owner/"),
			Div(Attr("data-portal-navigation-target", "collapseRow"), Class("hidden justify-end lg:flex"),
				Button(Type("button"), Attr("aria-label", "Contraer navegación"), Title("Contraer navegación"), Attr("data-portal-navigation-target", "collapseButton"), Attr("data-action", "click->portal-navigation#toggleCompact"), Class("inline-flex size-9 shrink-0 items-center justify-center rounded-lg text-muted hover:bg-white hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
					lucide.ChevronLeft(Class("size-4"), Attr("data-portal-navigation-target", "collapseIcon"), Attr("aria-hidden", "true")),
					lucide.ChevronRight(Class("hidden size-4"), Attr("data-portal-navigation-target", "expandIcon"), Attr("aria-hidden", "true")),
				),
			),
		),
		Nav(Attr("aria-label", "Secciones del portal de dueños"), Class("grid gap-1 pt-4"), Map(links, func(link portalLink) Node {
			active := path == link.href || link.href != "/owner/" && strings.HasPrefix(path, link.href+"/")
			class := "flex min-h-11 items-center justify-start gap-3 rounded-lg px-3 text-sm font-bold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"
			if active {
				class += " bg-white text-brand-dark"
			} else {
				class += " text-muted hover:bg-white hover:text-ink"
			}
			return A(Href(link.href), If(active, Attr("aria-current", "page")), Attr("aria-label", link.label), Title(link.label), Attr("data-portal-navigation-target", "navigationLink"), Attr("data-action", "click->portal-navigation#closeMobile"), Class(class), link.icon, Span(Attr("data-portal-navigation-target", "linkLabel"), Text(link.label)))
		})),
		Div(Attr("data-portal-navigation-target", "accountActions"), Class("mt-auto flex items-center justify-between gap-2 border-t border-line pt-4"),
			A(Href("/owner/profile"), If(path == "/owner/profile", Attr("aria-current", "page")), Attr("aria-label", "Perfil"), Title("Perfil"), Attr("data-action", "click->portal-navigation#closeMobile"), Class(profileClass),
				lucide.UserRound(Class("size-5"), Attr("aria-hidden", "true")),
			),
			Form(Method("post"), Action("/owner/logout"), Class("shrink-0"),
				Button(Type("submit"), Attr("aria-label", "Cerrar sesión"), Title("Salir"), Attr("data-portal-navigation-target", "logoutButton"), Class("inline-flex size-11 shrink-0 items-center justify-center rounded-lg text-muted transition hover:bg-white hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
					lucide.LogOut(Class("size-5"), Attr("aria-hidden", "true")),
				),
			),
		),
	)
}

func SiteFooter() Node {
	return Footer(Class("border-t border-line bg-cream"),
		Div(Class("mx-auto grid max-w-7xl gap-x-8 gap-y-8 px-5 py-10 text-sm text-muted sm:grid-cols-2 sm:px-8 sm:py-12 lg:grid-cols-[minmax(0,1.3fr)_minmax(8rem,.6fr)_minmax(0,1fr)_minmax(0,.9fr)]"),
			Div(Class("sm:col-span-2 lg:col-span-1"),
				Brand(),
				P(Class("mt-3 max-w-sm leading-6"), Text("Encuentra cabañas para pasadía o estadía, revisa su disponibilidad en tus fechas y habla directamente por WhatsApp con el contacto de cada una.")),
			),
			Div(
				H2(Class("font-extrabold text-ink"), Text("Explora")),
				Nav(Attr("aria-label", "Explorar"), Class("mt-3 grid gap-2"),
					A(Href("/#buscar"), Class("w-fit font-bold text-brand-dark underline decoration-line underline-offset-4 transition hover:decoration-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Text("Buscar cabañas")),
				),
			),
			Div(
				H2(Class("font-extrabold text-ink"), Text("¿Tienes una cabaña?")),
				P(Class("mt-3 max-w-xs leading-6"), Text("Gestiona tus cabañas, disponibilidad y reservas desde tu portal.")),
				A(Href("/owner/login"), Class("mt-3 inline-flex w-fit items-center font-bold text-brand-dark underline decoration-line underline-offset-4 transition hover:decoration-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), Text("Portal de propietarios")),
			),
			Div(
				H2(Class("font-extrabold text-ink"), Text("Conecta con Travelando")),
				P(Class("mt-3 max-w-xs leading-6"), Text("Escríbenos por WhatsApp o encuéntranos en redes sociales.")),
				Nav(Attr("aria-label", "Contacto y redes sociales"), Class("mt-2 flex items-center gap-2"),
					socialLink("https://api.whatsapp.com/message/MQDALY2E3MHQK1?autoload=1&app_absent=0", "WhatsApp", whatsAppIcon()),
					socialLink("https://www.instagram.com/travelando.co/", "Instagram", instagramIcon()),
					socialLink("https://www.tiktok.com/@travelandoconvictor", "TikTok", tikTokIcon()),
				),
			),
		),
	)
}

func socialLink(href, label string, icon Node) Node {
	return A(Href(href), Attr("aria-label", label), Title(label), Attr("target", "_blank"), Attr("rel", "noopener noreferrer"), Class("inline-flex size-10 items-center justify-center rounded-full text-muted transition hover:bg-white hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"), icon)
}

func whatsAppIcon() Node {
	return El("svg", Attr("viewBox", "0 0 24 24"), Attr("fill", "currentColor"), Attr("aria-hidden", "true"), Class("size-5"),
		El("path", Attr("d", "M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.198-.347.223-.644.075-.297-.149-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.297-.347.446-.52.149-.174.198-.298.297-.497.099-.198.05-.372-.025-.52-.075-.149-.669-1.612-.916-2.207-.241-.579-.487-.5-.669-.51-.173-.009-.371-.01-.57-.01-.198 0-.52.074-.792.372s-1.04 1.016-1.04 2.479 1.064 2.875 1.213 3.074c.149.198 2.095 3.2 5.076 4.487.709.306 1.262.489 1.694.626.712.226 1.36.194 1.872.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347M12.05 22a9.87 9.87 0 0 1-5.031-1.378l-.361-.214-3.741.982.999-3.648-.235-.374a9.86 9.86 0 0 1-1.511-5.26c.001-5.45 4.437-9.884 9.888-9.884 2.64.001 5.122 1.03 6.99 2.898a9.825 9.825 0 0 1 2.893 6.994C21.938 17.566 17.503 22 12.05 22m8.413-18.297A11.815 11.815 0 0 0 12.05.105C5.495.105.16 5.44.157 11.997c-.001 2.096.547 4.142 1.588 5.946L.057 24l6.304-1.654a11.88 11.88 0 0 0 5.684 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 0 0-3.479-8.198")),
	)
}

func instagramIcon() Node {
	return El("svg", Attr("viewBox", "0 0 24 24"), Attr("fill", "none"), Attr("stroke", "currentColor"), Attr("stroke-width", "1.8"), Attr("aria-hidden", "true"), Class("size-5"),
		El("rect", Attr("x", "3"), Attr("y", "3"), Attr("width", "18"), Attr("height", "18"), Attr("rx", "5")),
		El("circle", Attr("cx", "12"), Attr("cy", "12"), Attr("r", "4")),
		El("circle", Attr("cx", "17.5"), Attr("cy", "6.5"), Attr("r", ".75"), Attr("fill", "currentColor"), Attr("stroke", "none")),
	)
}

func tikTokIcon() Node {
	return El("svg", Attr("viewBox", "0 0 24 24"), Attr("fill", "currentColor"), Attr("aria-hidden", "true"), Class("size-5"),
		El("path", Attr("d", "M19.59 6.69a4.83 4.83 0 0 1-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 0 1-2.89 2.89 2.89 2.89 0 0 1-2.89-2.89 2.89 2.89 0 0 1 2.89-2.89c.3 0 .6.05.89.14V9.4a6.4 6.4 0 0 0-.89-.06 6.34 6.34 0 0 0-6.34 6.33A6.34 6.34 0 0 0 9.48 22a6.34 6.34 0 0 0 6.34-6.33V9.05a8.18 8.18 0 0 0 4.77 1.53V7.13c-.34 0-.68-.16-1-.44Z")),
	)
}

func PageHeading(title, description string) Node {
	return Div(Class("mb-4 max-w-2xl"),
		H1(Class("text-3xl font-extrabold tracking-tight text-ink sm:text-4xl"), Text(title)),
		P(Class("mt-1.5 max-w-prose text-base leading-6 text-muted"), Text(description)),
	)
}

func Notice(message string, isError bool) Node {
	class := "rounded-xl border border-forest/20 bg-forest-soft px-4 py-3 text-sm text-forest"
	if isError {
		class = "rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
	}
	return Div(Class(class), Role("status"), Text(message))
}

func FieldLabel(forID, label string) Node {
	return Label(For(forID), Class("mb-1.5 block text-[0.8125rem] font-bold text-ink"), Text(label))
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
