package requests

import (
	"strings"
	"testing"

	"go.leapkit.dev/core/form/validate"
	"travelando/internal/system/models"
)

func TestReportShowsCancelledReservationStatus(t *testing.T) {
	var html strings.Builder
	inquiry := models.Inquiry{Status: models.InquiryCancelled, CancelledFromStatus: models.InquiryPending, StayType: models.StayDayUse, DateFrom: "2099-02-10", DateTo: "2099-02-11"}
	if err := reportPage(inquiry, "report-token", nil, validate.Errors{}).Render(&html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Solicitud cancelada") || strings.Contains(html.String(), "Pendiente de confirmar") {
		t.Fatalf("report status = %s, want cancelled", html.String())
	}
}
