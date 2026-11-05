package dashboard

import (
	"fmt"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

type dashboardData struct {
	Inquiries    []models.Inquiry
	CommissionBP int
	DepositCOP   int64
}

func loadDashboardData(inquirySvc services.InquiryService, commissionSvc services.CommissionService, depositSvc services.DepositService) (dashboardData, error) {
	var data dashboardData
	var err error
	if data.Inquiries, err = inquirySvc.Inquiries(); err != nil {
		return dashboardData{}, fmt.Errorf("loading admin inquiries: %w", err)
	}
	if data.CommissionBP, err = commissionSvc.CommissionRate(); err != nil {
		return dashboardData{}, fmt.Errorf("loading global commission rate: %w", err)
	}
	if data.DepositCOP, err = depositSvc.DepositAmount(); err != nil {
		return dashboardData{}, fmt.Errorf("loading global deposit amount: %w", err)
	}
	return data, nil
}
