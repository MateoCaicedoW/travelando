package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func MetricCard(label, value, detail, valueClass string) Node {
	return Div(Class("rounded-xl border border-line bg-white p-4 sm:p-5"),
		P(Class("text-label font-extrabold uppercase tracking-wider text-muted"), Text(label)),
		P(Class("mt-2 text-2xl font-black tracking-tight "+valueClass), Text(value)),
		P(Class("mt-1 text-xs text-muted"), Text(detail)),
	)
}

func QuickList[T any](title string, values []T, render func(T) Node) Node {
	return Div(Class("rounded-2xl border border-line bg-white p-4 sm:p-5"),
		P(Class("text-label font-extrabold uppercase tracking-[0.08em] text-brand-dark"), Text("Inventario")),
		H2(Class("mt-1 text-lg font-extrabold text-ink"), Text(title)),
		If(len(values) == 0, P(Class("mt-3 text-sm text-muted"), Text("Todavía no hay registros."))),
		Map(values, render),
	)
}
