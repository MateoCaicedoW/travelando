package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SearchableOption struct {
	Value string
	Label string
}

func SearchableSelect(id, name, label, placeholder, value string, options []SearchableOption, required bool, errorMessage string) Node {
	inputClass := "min-h-11 w-full rounded-xl border border-line bg-white px-4 py-2.5 pr-10 text-base text-ink placeholder:text-muted focus:border-brand focus:ring-2 focus:ring-brand/20"
	if errorMessage != "" {
		inputClass += " border-red-300 focus:border-red-500 focus:ring-red-200"
	}
	return Div(Class("min-w-0"),
		FieldLabel(id, label),
		Div(ID(id+"-combobox"), Attr("data-controller", "searchable-select"), Attr("data-action", "change->searchable-select#clearValidity"), Class("combobox travelando-combobox relative w-full"), If(required, Attr("data-required", "true")),
			Input(ID(id), Type("text"), Role("combobox"), Attr("data-searchable-select-target", "input"), Attr("data-action", "input->searchable-select#clearValidity focus->searchable-select#selectValue"), Placeholder(placeholder), Attr("autocomplete", "off"), Attr("autocorrect", "off"), Attr("spellcheck", "false"), Attr("aria-autocomplete", "list"), Attr("aria-expanded", "false"), Attr("aria-controls", id+"-listbox"), If(required, Group{Required(), Attr("aria-required", "true")}), If(errorMessage != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", id+"-error")}), Class(inputClass)),
			lucide.ChevronDown(Class("pointer-events-none absolute right-3 top-1/2 size-4 -translate-y-1/2 text-muted"), Attr("aria-hidden", "true")),
			Div(Attr("data-popover", ""), Attr("aria-hidden", "true"), Class("absolute left-0 top-full z-[60] mt-1 w-full overflow-hidden rounded-xl border border-line bg-white shadow-lg transition duration-150"),
				Div(ID(id+"-listbox"), Role("listbox"), Attr("data-empty", ""), Class("max-h-60 overflow-y-auto p-1"),
					Map(options, func(option SearchableOption) Node {
						return Div(Role("option"), Attr("data-value", option.Value), If(option.Value == value, Attr("aria-selected", "true")), Class("min-h-10 cursor-pointer rounded-lg px-3 py-2 text-sm text-ink"), Text(option.Label))
					}),
				),
			),
			Input(Type("hidden"), Name(name), Value(value)),
		),
		FieldError(id, errorMessage),
	)
}
