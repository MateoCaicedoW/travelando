package system

import (
	"net/http"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
	"travelando/internal/system/assets"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
)

// Layout defines the shared document around a page's content.
type Layout struct {
	Title       string
	Description string
	Portal      string
	Yield       Node
}

const (
	PortalAdmin = "admin"
	PortalOwner = "owner"
	PortalBare  = "bare"
)

// New builds the shared HTML document and resolves development or bundled assets.
func (props Layout) New(request ...*http.Request) Node {
	head := []Node{
		Meta(Charset("UTF-8")),
		Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
		TitleEl(Text(props.Title)),
		If(props.Description != "", Meta(Name("description"), Content(props.Description))),
		Link(Rel("stylesheet"), Href(assets.Path("/css/tailwind.css"))),
		Script(Type("module"), Src(assets.Path("/js/app.js"))),
	}
	if assets.IsDev() {
		head = append([]Node{Script(Type("module"), Src("http://localhost:5173/@vite/client"))}, head...)
	}

	var toast Node
	var path string
	if len(request) > 0 {
		flash := helpers.ConsumeFlash(request[0])
		toast = components.FlashToast(flash.Message, flash.Level)
		path = request[0].URL.Path
	}
	header, footer := components.SiteHeader(), components.SiteFooter()
	switch props.Portal {
	case PortalAdmin:
		header, footer = components.AdminHeader(path), nil
	case PortalOwner:
		header, footer = components.OwnerHeader(path), nil
	case PortalBare:
		header, footer = nil, nil
	}
	var page Node
	if props.Portal == PortalAdmin {
		page = Div(Attr("data-controller", "admin-navigation"), Class("flex min-h-screen w-full"),
			components.AdminNavigationBackdrop(),
			components.AdminHeader(path),
			Div(Attr("data-admin-navigation-target", "content"), Class("flex min-w-0 flex-1 flex-col"),
				components.AdminMobileBar(),
				Div(ID("global-flash-container"), toast),
				Main(Class("w-full flex-1"), props.Yield),
			),
		)
	} else {
		page = Div(Class("flex min-h-screen flex-col"),
			If(header != nil, header),
			Div(ID("global-flash-container"), toast),
			Main(Class("w-full flex-1"), props.Yield),
			If(footer != nil, footer),
		)
	}
	return Doctype(HTML(Lang("es"), Class("h-full"),
		Head(head...),
		Body(Class("min-h-full bg-white font-sans text-ink antialiased"),
			page,
		),
	))
}
