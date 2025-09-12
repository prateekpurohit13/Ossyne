package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username        string   `gorm:"unique;not null" json:"username"`
	Email           string   `gorm:"unique;not null" json:"email"`
	AvatarURL       string   `json:"avatar_url"`
	GithubID        *string  `gorm:"unique" json:"github_id"`
	ReputationScore int      `gorm:"default:0" json:"reputation_score"`
	Roles           JSONStringSlice `gorm:"type:json" json:"roles"`
	Projects        []Project `gorm:"foreignKey:OwnerID"`
}

type Project struct {
	gorm.Model
	OwnerID     uint   `gorm:"not null" json:"owner_id"`
	Title       string `gorm:"not null" json:"title"`
	ShortDesc   string `json:"short_desc"`
	RepoURL     string `gorm:"unique" json:"repo_url"`
	Tags        JSONStringSlice `gorm:"type:json" json:"tags"`
	Visibility  string `gorm:"type:enum('public', 'private');default:'public'" json:"visibility"`
	Tasks       []Task
}

type Task struct {
	gorm.Model
	ProjectID       uint    `gorm:"not null" json:"project_id"`
	Title           string  `gorm:"not null" json:"title"`
	Description     string  `json:"description"`
	DifficultyLevel string  `gorm:"type:enum('easy', 'medium', 'hard');default:'easy'" json:"difficulty_level"`
	EstimatedHours  int     `json:"estimated_hours"`
	Tags            JSONStringSlice `gorm:"type:json" json:"tags"`
	SkillsRequired  JSONStringSlice `gorm:"type:json" json:"skills_required"`
	BountyAmount    float64 `gorm:"type:decimal(10,2);default:0.00" json:"bounty_amount"`
	Status          string  `gorm:"type:enum('open', 'claimed', 'in_progress', 'submitted', 'completed', 'archived');default:'open'" json:"status"`
}

type Claim struct {
	gorm.Model
	TaskID    uint      `gorm:"not null;uniqueIndex:idx_task_user_claim" json:"task_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_task_user_claim" json:"user_id"`
	ClaimDate time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"claim_date"`
	Status    string    `gorm:"type:enum('pending', 'accepted', 'rejected', 'withdrawn');default:'pending';not null" json:"status"`
	MentorID  *uint     `json:"mentor_id,omitempty"`
	Notes     string    `gorm:"type:text" json:"notes"`

	Task   Task `gorm:"foreignKey:TaskID"`
	User   User `gorm:"foreignKey:UserID"`
	Mentor User `gorm:"foreignKey:MentorID"`
}

type Contribution struct {
	gorm.Model
	TaskID           uint       `gorm:"not null;uniqueIndex:idx_task_user_contrib" json:"task_id"`
	UserID           uint       `gorm:"not null;uniqueIndex:idx_task_user_contrib" json:"user_id"`
	PRURL            string     `gorm:"not null" json:"pr_url"`
	PRCommitHashes   JSONStringSlice `gorm:"type:json" json:"pr_commit_hashes"`
	SubmittedAt      time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"submitted_at"`
	VerificationStatus string   `gorm:"type:enum('unverified', 'auto_verified', 'manual_verified', 'rejected');default:'unverified';not null" json:"verification_status"`
	AcceptedAt       *time.Time `json:"accepted_at,omitempty"`
	PayoutAmount     float64    `gorm:"type:decimal(10,2);default:0.00" json:"payout_amount"`

	Task Task `gorm:"foreignKey:TaskID"`
	User User `gorm:"foreignKey:UserID"`
}