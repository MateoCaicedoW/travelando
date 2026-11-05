package components

import (
	"strings"
	"testing"

	. "maragu.dev/gomponents"
)

func TestCustomDialogUsesAvailableWidth(t *testing.T) {
	var html strings.Builder
	if err := CustomDialog(DialogProps{Yield: Text("Contenido")}).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "w-full max-w-7xl") || strings.Contains(html.String(), "max-w-5xl") {
		t.Fatalf("dialog should fill available width up to 7xl: %s", html.String())
	}
}
