package services

type CommissionService interface {
	CommissionRate() (int, error)
	UpdateCommissionRate(int) error
}

func (s *commissionService) CommissionRate() (int, error) {
	db, err := s.conn()
	if err != nil {
		return 0, err
	}
	var rate int
	err = db.QueryRow(`SELECT commission_bps FROM commission_settings WHERE id = 1`).Scan(&rate)
	return rate, err
}

func (s *commissionService) UpdateCommissionRate(rate int) error {
	if rate < 0 || rate > 10000 {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	result, err := db.Exec(`UPDATE commission_settings SET commission_bps = ? WHERE id = 1`, rate)
	if err != nil {
		return err
	}
	return requireUpdated(result)
}
