package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const (
	FormActionsClass   = "mt-4 flex flex-col gap-3 border-t border-line pt-4 sm:flex-row sm:flex-wrap sm:items-center sm:justify-end"
	FormPrimaryClass   = "inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg bg-brand-dark px-5 py-2.5 font-extrabold text-white hover:brightness-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark sm:w-auto"
	FormSecondaryClass = "inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-lg border border-line px-4 py-2 font-bold text-muted hover:text-ink sm:w-auto"
)

func FormActions(children ...Node) Node {
	return Div(Class(FormActionsClass), Group(children))
}
