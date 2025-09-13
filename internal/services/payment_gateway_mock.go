package services

import (
	"time"

	"github.com/google/uuid"
)

//MockPaymentGateway simulates interactions with an external payment provider like Stripe or PayPal.
type MockPaymentGateway struct{}

func (m *MockPaymentGateway) Escrow(amount float64, currency, description string) (string, error) {
	time.Sleep(50 * time.Millisecond)
	escrowID := "esc_" + uuid.New().String()
	return escrowID, nil
}

func (m *MockPaymentGateway) ReleaseEscrow(escrowID string, amount float64, currency, recipientUserID string) (string, error) {
	time.Sleep(100 * time.Millisecond)
	transactionID := "txn_" + uuid.New().String()
	return transactionID, nil
}

func (m *MockPaymentGateway) ProcessWithdrawal(userID uint, amount float64, currency string) (string, error) {
	time.Sleep(150 * time.Millisecond)
	transactionID := "wdr_" + uuid.New().String()
	return transactionID, nil
}

func (m *MockPaymentGateway) RefundEscrow(escrowID string, amount float64, currency string) error {
	time.Sleep(50 * time.Millisecond)
	return nil
}