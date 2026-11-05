package services

import "testing"

func TestDepositServiceStoresGlobalCOPAmount(t *testing.T) {
	database, conn := cabinTestDatabase(t)
	defer database.Close()

	deposits := NewDepositService(conn)
	amount, err := deposits.DepositAmount()
	if err != nil || amount != 0 {
		t.Fatalf("initial deposit = %d, err %v; want zero", amount, err)
	}
	if err := deposits.UpdateDepositAmount(250000); err != nil {
		t.Fatal(err)
	}
	amount, err = deposits.DepositAmount()
	if err != nil || amount != 250000 {
		t.Fatalf("updated deposit = %d, err %v; want 250000", amount, err)
	}
	for _, invalid := range []int64{-1, MaxAmountCOP + 1} {
		if err := deposits.UpdateDepositAmount(invalid); err == nil {
			t.Errorf("deposit amount %d should be rejected", invalid)
		}
	}
}
