package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func FlashToast(message, level string) Node {
	if message == "" {
		return nil
	}
	className := "fixed right-4 top-4 z-50 flex w-[calc(100vw-2rem)] max-w-sm items-start justify-between gap-4 rounded-xl border px-4 py-3 text-sm font-semibold shadow-md transition duration-200 ease-out motion-reduce:transition-none sm:right-6 sm:top-6"
	role := "status"
	if level == "error" {
		className += " border-red-200 bg-red-50 text-red-800"
		role = "alert"
	} else {
		className += " border-forest/20 bg-forest-soft text-forest"
	}
	return Div(ID("global-flash"), Attr("data-controller", "flash"), Attr("data-global-flash", ""), Attr("data-flash-level", level), Class(className), Role(role),
		Span(Text(message)),
		Button(Type("button"), Attr("data-action", "click->flash#dismiss"), Attr("data-flash-dismiss", ""), Attr("aria-label", "Cerrar notificación"), Class("shrink-0 font-bold underline underline-offset-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2"), Text("Cerrar")),
	)
}
