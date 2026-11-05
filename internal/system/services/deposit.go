package services

type DepositService interface {
	DepositAmount() (int64, error)
	UpdateDepositAmount(int64) error
}

func (s *depositService) DepositAmount() (int64, error) {
	db, err := s.conn()
	if err != nil {
		return 0, err
	}
	var amount int64
	err = db.QueryRow(`SELECT deposit_cop FROM commission_settings WHERE id = 1`).Scan(&amount)
	return amount, err
}

func (s *depositService) UpdateDepositAmount(amount int64) error {
	if amount < 0 || amount > MaxAmountCOP {
		return ErrInvalidInput
	}
	db, err := s.conn()
	if err != nil {
		return err
	}
	result, err := db.Exec(`UPDATE commission_settings SET deposit_cop = ? WHERE id = 1`, amount)
	if err != nil {
		return err
	}
	return requireUpdated(result)
}
