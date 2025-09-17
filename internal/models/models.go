package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username          string          `gorm:"unique;not null" json:"username"`
	Email             string          `gorm:"unique;not null" json:"email"`
	AvatarURL         string          `json:"avatar_url"`
	GithubID          *string         `gorm:"unique" json:"github_id"`
	GitHubAccessToken *string         `gorm:"column:github_access_token" json:"-"`
	Ratings           int             `gorm:"default:0" json:"ratings"`
	Roles             JSONStringSlice `gorm:"type:json" json:"roles"`
	Projects          []Project       `gorm:"foreignKey:OwnerID"`
	UserSkills        []UserSkill     `gorm:"foreignKey:UserID"`
	Payments          []Payment       `gorm:"foreignKey:UserID"`
}

type Project struct {
	gorm.Model
	OwnerID    uint            `gorm:"not null" json:"owner_id"`
	Title      string          `gorm:"not null" json:"title"`
	ShortDesc  string          `json:"short_desc"`
	RepoURL    string          `gorm:"unique" json:"repo_url"`
	Tags       JSONStringSlice `gorm:"type:json" json:"tags"`
	Visibility string          `gorm:"type:enum('public', 'private');default:'public'" json:"visibility"`
	Tasks      []Task
}

type Task struct {
	gorm.Model
	ProjectID       uint            `gorm:"not null" json:"project_id"`
	Title           string          `gorm:"not null" json:"title"`
	Description     string          `json:"description"`
	DifficultyLevel string          `gorm:"type:enum('easy', 'medium', 'hard');default:'easy'" json:"difficulty_level"`
	EstimatedHours  int             `json:"estimated_hours"`
	Tags            JSONStringSlice `gorm:"type:json" json:"tags"`
	SkillsRequired  JSONStringSlice `gorm:"type:json" json:"skills_required"`
	BountyAmount    float64         `gorm:"type:decimal(10,2);default:0.00" json:"bounty_amount"`
	BountyEscrowID  *string         `json:"bounty_escrow_id,omitempty"`
	Status          string          `gorm:"type:enum('open', 'claimed', 'in_progress', 'submitted', 'completed', 'archived');default:'open'" json:"status"`
}

type Claim struct {
	gorm.Model
	TaskID    uint      `gorm:"not null;uniqueIndex:idx_task_user_claim" json:"task_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_task_user_claim" json:"user_id"`
	ClaimDate time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"claim_date"`
	Status    string    `gorm:"type:enum('pending', 'accepted', 'rejected', 'withdrawn');default:'pending';not null" json:"status"`
	MentorID  *uint     `json:"mentor_id,omitempty"`
	Notes     string    `gorm:"type:text" json:"notes"`
	Task      Task      `gorm:"foreignKey:TaskID"`
	User      User      `gorm:"foreignKey:UserID"`
	Mentor    User      `gorm:"foreignKey:MentorID"`
}

type Contribution struct {
	gorm.Model
	TaskID             uint            `gorm:"not null;uniqueIndex:idx_task_user_contrib" json:"task_id"`
	UserID             uint            `gorm:"not null;uniqueIndex:idx_task_user_contrib" json:"user_id"`
	PRURL              string          `gorm:"not null" json:"pr_url"`
	PRCommitHashes     JSONStringSlice `gorm:"type:json" json:"pr_commit_hashes"`
	SubmittedAt        time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP" json:"submitted_at"`
	VerificationStatus string          `gorm:"type:enum('unverified', 'auto_verified', 'manual_verified', 'rejected');default:'unverified';not null" json:"verification_status"`
	AcceptedAt         *time.Time      `json:"accepted_at,omitempty"`
	PayoutAmount       float64         `gorm:"type:decimal(10,2);default:0.00" json:"payout_amount"`
	PaymentID          *uint           `json:"payment_id,omitempty"`
	Task               *Task           `gorm:"foreignKey:TaskID"`
	User               *User           `gorm:"foreignKey:UserID"`
	Payment            *Payment        `gorm:"foreignKey:PaymentID"`
}

type Skill struct {
	gorm.Model
	Name        string `gorm:"unique;not null" json:"name"`
	Description string `json:"description"`
}

type UserSkill struct {
	UserID  uint   `gorm:"primaryKey" json:"user_id"`
	SkillID uint   `gorm:"primaryKey" json:"skill_id"`
	Level   string `gorm:"type:enum('beginner', 'intermediate', 'expert');default:'beginner';not null" json:"level"`
	User    User   `gorm:"foreignKey:UserID"`
	Skill   Skill  `gorm:"foreignKey:SkillID"`
}

type ReputationEventLog struct {
	gorm.Model
	UserID      uint   `gorm:"not null" json:"user_id"`
	EventType   string `gorm:"type:enum('contribution_accepted', 'mentor_endorsement', 'bounty_earned', 'manual_adjustment');not null" json:"event_type"`
	ScoreChange int    `gorm:"not null" json:"score_change"`
	RelatedID   *uint  `json:"related_id,omitempty"`
	Notes       string `gorm:"type:text" json:"notes"`

	User User `gorm:"foreignKey:UserID"`
}

type Payment struct {
	gorm.Model
	ContributionID *uint     `json:"contribution_id,omitempty"`
	UserID         uint      `gorm:"not null" json:"user_id"`
	Amount         float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency       string    `gorm:"type:varchar(3);default:'USD';not null" json:"currency"`
	Status         string    `gorm:"not null" json:"status"`
	Type           string    `gorm:"not null" json:"type"`
	TransactionID  string    `gorm:"unique" json:"transaction_id"`
	PaymentGateway string    `json:"payment_gateway"`
	PaymentDate    time.Time `json:"payment_date"`
}