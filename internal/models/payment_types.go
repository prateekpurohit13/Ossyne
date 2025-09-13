package models

const (
	PaymentTypeBountyPayout  = "bounty_payout"
	PaymentTypeEscrowDeposit = "escrow_deposit"
	PaymentTypeAdminTransfer = "admin_transfer"
)

const (
	PaymentStatusPending  = "pending"
	PaymentStatusEscrowed = "escrowed"
	PaymentStatusReleased = "released"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)