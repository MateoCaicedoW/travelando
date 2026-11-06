package profile

import (
	"travelando/internal/system/components"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func PINField(id, name, label, autocomplete, message string) Node {
	return Div(Class("min-w-0"),
		components.FieldLabel(id, label),
		Input(ID(id), Name(name), Type("password"), InputMode("numeric"), Pattern("[0-9]{6}"), MinLength("6"), MaxLength("6"), AutoComplete(autocomplete), If(message != "", Group{Attr("aria-invalid", "true"), Attr("aria-describedby", id+"-error")}), Class("w-full rounded-xl border border-line bg-white px-4 py-3 text-base text-ink focus:border-brand focus:ring-2 focus:ring-brand/20")),
		components.FieldError(id, message),
	)
}
