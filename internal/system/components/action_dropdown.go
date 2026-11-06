package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ActionDropdown(id string, items ...Node) Node {
	rootID := "action-menu-" + id
	triggerID := rootID + "-trigger"
	menuID := rootID + "-menu"
	return Div(ID(rootID), Class("dropdown-menu relative inline-flex"),
		Button(Type("button"), ID(triggerID), Attr("aria-label", "Acciones"), Attr("aria-haspopup", "menu"), Attr("aria-controls", menuID), Attr("aria-expanded", "false"), Class("inline-flex size-9 items-center justify-center rounded-lg border border-line text-muted hover:border-brand hover:bg-cream hover:text-brand-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-dark"),
			lucide.Ellipsis(Class("size-5"), Attr("aria-hidden", "true")),
		),
		Div(ID(rootID+"-popover"), Attr("data-popover", ""), Attr("aria-hidden", "true"), Attr("data-align", "end"), Class("absolute right-0 top-full z-50 mt-1 w-48 rounded-lg border border-line bg-white p-1 shadow-lg"),
			Div(ID(menuID), Attr("role", "menu"), Attr("aria-labelledby", triggerID), Class("outline-hidden"), Group(items)),
		),
	)
}
