package components

import (
	"strings"
	"testing"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func TestFormActionsPlaceActionsRightAtBottom(t *testing.T) {
	var html strings.Builder
	if err := FormActions(
		Button(Type("submit"), Class(FormPrimaryClass), Text("Guardar")),
		Button(Type("button"), Class(FormSecondaryClass), Text("Cancelar")),
	).Render(&html); err != nil {
		t.Fatal(err)
	}

	out := html.String()
	for _, want := range []string{FormActionsClass, FormPrimaryClass, FormSecondaryClass} {
		if !strings.Contains(out, want) {
			t.Errorf("form actions missing %q: %s", want, out)
		}
	}
	if strings.Index(out, "Guardar") > strings.Index(out, "Cancelar") {
		t.Errorf("primary action should render before cancel: %s", out)
	}
	if !strings.Contains(out, "sm:justify-end") {
		t.Errorf("form actions should align right on desktop: %s", out)
	}
}
