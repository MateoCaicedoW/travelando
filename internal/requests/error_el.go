package requests

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func errorPage() Node {
	return Div(Class("mx-auto w-full max-w-xl px-5 py-16 text-center sm:px-8"),
		P(Class("text-5xl text-brand"), Text("!")),
		H1(Class("mt-4 text-2xl font-extrabold text-ink"), Text("No se pudo continuar")),
		P(Class("mt-2 leading-7 text-muted"), Text("Intenta de nuevo o vuelve al inicio.")),
		A(Href("/"), Class("mt-6 inline-flex min-h-11 items-center rounded-lg bg-brand-dark px-5 font-bold text-white transition hover:brightness-90"), Text("Volver a Travelando")),
	)
}
