package models

const (
	TaskStatusOpen        = "open"
	TaskStatusClaimed     = "claimed"
	TaskStatusInProgress  = "in_progress"
	TaskStatusSubmitted   = "submitted"
	TaskStatusCompleted   = "completed"
	TaskStatusArchived    = "archived"
)

const (
	VerificationStatusUnverified   = "unverified"
	VerificationStatusAutoVerified = "auto_verified"
	VerificationStatusManualVerified = "manual_verified"
	VerificationStatusRejected   = "rejected"
)