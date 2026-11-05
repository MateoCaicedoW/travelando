package requests

import (
	"testing"

	"travelando/internal/system/models"
)

func TestFinalTotalMustIncludeDeposit(t *testing.T) {
	validate := validFinalTotal(string(models.InquiryReserved), 50000)
	if err := validate([]string{"49999"}); err == nil {
		t.Fatal("reserved total below deposit should be rejected")
	}
	if err := validate([]string{"50000"}); err != nil {
		t.Fatalf("total equal to deposit should be valid: %v", err)
	}
	if err := validFinalTotal(string(models.InquiryLost), 50000)(nil); err != nil {
		t.Fatalf("lost inquiry should not require a total: %v", err)
	}
}
