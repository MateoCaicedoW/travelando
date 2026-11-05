package components

import (
	"bytes"
	"strings"
	"testing"
)

func TestFlashToastUsesStimulusController(t *testing.T) {
	var output bytes.Buffer
	if err := FlashToast("Guardado", "success").Render(&output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`data-controller="flash"`,
		`data-action="click-&gt;flash#dismiss"`,
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("flash toast missing %q: %s", want, output.String())
		}
	}
}
