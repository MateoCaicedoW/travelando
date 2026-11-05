package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DialogProps struct {
	Yield   Node
	Label   string
	SwapOOB string
}

func CustomDialog(p DialogProps) Node {
	label := p.Label
	if label == "" {
		label = "Crear cabaña"
	}
	return Div(
		ID("modal"),
		If(p.SwapOOB != "", Attr("hx-swap-oob", p.SwapOOB)),
		Class("fixed inset-0 z-50 flex items-start justify-center overflow-y-auto overscroll-contain p-2 sm:p-3 md:p-5"),
		Data("modal-layer", "true"),
		Data("controller", "modal"),
		Data("modal-target", "modal"),
		Div(
			Class("fixed inset-0 bg-ink/60 modal-backdrop"),
			Data("action", "click->modal#close"),
			Data("modal-closer", "true"),
		),
		Div(
			Class("relative z-10 max-h-[calc(100dvh-1rem)] w-full max-w-5xl overflow-y-auto rounded-2xl border border-line bg-white shadow-xl modal-content sm:max-h-[calc(100dvh-1.5rem)] md:max-h-[calc(100dvh-2.5rem)]"),
			Role("dialog"),
			Attr("aria-modal", "true"),
			Attr("aria-label", label),
			Attr("tabindex", "-1"),
			p.Yield,
		),
	)
}
