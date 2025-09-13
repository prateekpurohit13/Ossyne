package services
import (
	"fmt"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"time"
)

type PaymentService struct {
	PaymentGateway *MockPaymentGateway
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		PaymentGateway: &MockPaymentGateway{},
	}
}

func (s *PaymentService) FundTaskBounty(taskID, funderUserID uint, amount float64, currency string) error {
	tx := db.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var task models.Task
	if err := tx.First(&task, taskID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task with ID %d not found: %w", taskID, err)
	}

	if amount <= 0 {
		tx.Rollback()
		return fmt.Errorf("bounty amount must be positive")
	}

	if task.Status != "open" && task.BountyAmount > 0 {
		tx.Rollback()
		return fmt.Errorf("task '%s' (status: %s) already has a bounty of %.2f %s, cannot re-fund", task.Title, task.Status, task.BountyAmount, currency)
	}
	escrowID, err := s.PaymentGateway.Escrow(amount, currency, fmt.Sprintf("Bounty for task: %s", task.Title))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to escrow funds with payment gateway: %w", err)
	}

	payment := models.Payment{
		UserID:         funderUserID,
		Amount:         amount,
		Currency:       currency,
		Status:         models.PaymentStatusEscrowed,
		Type:           models.PaymentTypeEscrowDeposit,
		TransactionID:  escrowID,
		PaymentGateway: "MockPG",
		PaymentDate:    time.Now(),
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record escrow deposit in DB: %w", err)
	}

	task.BountyAmount = amount
	task.BountyEscrowID = &escrowID
	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update task with bounty details: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PaymentService) ReleaseBountyToContributor(contributionID uint) error {
	tx := db.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var contribution models.Contribution
	if err := tx.Preload("Task").Preload("User").First(&contribution, contributionID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("contribution with ID %d not found: %w", contributionID, err)
	}

	if contribution.VerificationStatus != "auto_verified" && contribution.VerificationStatus != "manual_verified" {
		tx.Rollback()
		return fmt.Errorf("contribution %d is not yet verified to release bounty (status: %s)", contributionID, contribution.VerificationStatus)
	}

	if contribution.Task.BountyEscrowID == nil || *contribution.Task.BountyEscrowID == "" {
		tx.Rollback()
		return fmt.Errorf("task %d has no escrowed bounty to release", contribution.TaskID)
	}

	if contribution.PaymentID != nil && *contribution.PaymentID != 0 {
		tx.Rollback()
		return fmt.Errorf("bounty for contribution %d has already been released (payment ID: %d)", contributionID, *contribution.PaymentID)
	}

	transactionID, err := s.PaymentGateway.ReleaseEscrow(
		*contribution.Task.BountyEscrowID,
		contribution.Task.BountyAmount,
		"USD",
		fmt.Sprintf("%d", contribution.UserID),
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to release escrowed funds: %w", err)
	}

	payout := models.Payment{
		ContributionID: &contribution.ID,
		UserID:         contribution.UserID,
		Amount:         contribution.Task.BountyAmount,
		Currency:       "USD",
		Status:         models.PaymentStatusReleased,
		Type:           models.PaymentTypeBountyPayout,
		TransactionID:  transactionID,
		PaymentGateway: "MockPG",
		PaymentDate:    time.Now(),
	}

	if err := tx.Create(&payout).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record payout in DB: %w", err)
	}

	contribution.PayoutAmount = payout.Amount
	payoutID := payout.ID
	contribution.PaymentID = &payoutID
	if err := tx.Save(&contribution).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to link payout to contribution: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PaymentService) RefundTaskBounty(taskID uint, reason string) error {
	tx := db.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var task models.Task
	if err := tx.First(&task, taskID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task with ID %d not found: %w", taskID, err)
	}

	if task.BountyEscrowID == nil || *task.BountyEscrowID == "" {
		tx.Rollback()
		return fmt.Errorf("task %d has no escrowed bounty to refund", taskID)
	}

	if err := s.PaymentGateway.RefundEscrow(*task.BountyEscrowID, task.BountyAmount, "USD"); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to refund escrowed funds: %w", err)
	}

	var escrowPayment models.Payment
	if err := tx.Where("transaction_id = ? AND type = ?", *task.BountyEscrowID, models.PaymentTypeEscrowDeposit).First(&escrowPayment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("escrow payment record for task %d not found: %w", taskID, err)
	}
	escrowPayment.Status = models.PaymentStatusRefunded
	if err := tx.Save(&escrowPayment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update escrow payment status to refunded: %w", err)
	}

	task.BountyAmount = 0.00
	task.BountyEscrowID = nil
	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear task bounty details: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PaymentService) GetUserPayments(userID uint) ([]models.Payment, error) {
	var payments []models.Payment
	if err := db.DB.Where("user_id = ?", userID).Order("payment_date DESC").Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve payments for user %d: %w", userID, err)
	}

	return payments, nil
}