package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username        string   `gorm:"unique;not null" json:"username"`
	Email           string   `gorm:"unique;not null" json:"email"`
	AvatarURL       string   `json:"avatar_url"`
	GithubID        string   `gorm:"unique" json:"github_id"`
	ReputationScore int      `gorm:"default:0" json:"reputation_score"`
	Roles           []string `gorm:"type:json" json:"roles"`
	Projects        []Project `gorm:"foreignKey:OwnerID"`
}

type Project struct {
	gorm.Model
	OwnerID     uint   `gorm:"not null" json:"owner_id"`
	Title       string `gorm:"not null" json:"title"`
	ShortDesc   string `json:"short_desc"`
	RepoURL     string `gorm:"unique" json:"repo_url"`
	Tags        []string `gorm:"type:json" json:"tags"`
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
	Tags            []string `gorm:"type:json" json:"tags"`
	SkillsRequired  []string `gorm:"type:json" json:"skills_required"`
	BountyAmount    float64 `gorm:"type:decimal(10,2);default:0.00" json:"bounty_amount"`
	Status          string  `gorm:"type:enum('open', 'claimed', 'in_progress', 'submitted', 'completed', 'archived');default:'open'" json:"status"`
}