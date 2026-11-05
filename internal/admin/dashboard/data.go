package dashboard

import (
	"fmt"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

type dashboardData struct {
	Inquiries    []models.Inquiry
	CommissionBP int
}

func loadDashboardData(inquirySvc services.InquiryService, commissionSvc services.CommissionService) (dashboardData, error) {
	var data dashboardData
	var err error
	if data.Inquiries, err = inquirySvc.Inquiries(); err != nil {
		return dashboardData{}, fmt.Errorf("loading admin inquiries: %w", err)
	}
	if data.CommissionBP, err = commissionSvc.CommissionRate(); err != nil {
		return dashboardData{}, fmt.Errorf("loading global commission rate: %w", err)
	}
	return data, nil
}
