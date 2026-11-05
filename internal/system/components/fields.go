package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func NumberField(id, name, label, value, min, max string) Node {
	return NumberFieldWithError(id, name, label, value, min, max, "")
}

func NumberFieldWithError(id, name, label, value, min, max, errorMessage string) Node {
	inputClass := "w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"
	if errorMessage != "" {
		inputClass += " border-red-300 focus:ring-red-200"
	}
	return Div(Class("min-w-0"),
		FieldLabel(id, label),
		Input(ID(id), Name(name), Type("number"), If(min != "", Min(min)), If(max != "", Max(max)), Step("1"), Value(value), Required(), If(errorMessage != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", id+"-error")}), Class(inputClass)),
		FieldError(id, errorMessage),
	)
}

func TimeField(id, name, label, value string) Node {
	return TimeFieldWithError(id, name, label, value, "")
}

func TimeFieldWithError(id, name, label, value, errorMessage string) Node {
	inputClass := "w-full bg-white px-4 py-3 text-base text-ink focus:ring-2 focus:ring-brand/20"
	if errorMessage != "" {
		inputClass += " border-red-300 focus:ring-red-200"
	}
	return Div(Class("min-w-0"),
		FieldLabel(id, label),
		Input(ID(id), Name(name), Type("time"), Value(value), Required(), If(errorMessage != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", id+"-error")}), Class(inputClass)),
		FieldError(id, errorMessage),
	)
}
