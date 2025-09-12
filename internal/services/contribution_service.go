package services

import (
	"fmt"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"time"
)

const (
	ReputationEventContributionAccepted = "contribution_accepted"
	ReputationEventMentorEndorsement    = "mentor_endorsement"
	ReputationEventBountyEarned         = "bounty_earned"
	ReputationEventManualAdjustment     = "manual_adjustment"
)

type ContributionService struct{}

func (s *ContributionService) VerifyAndAcceptContribution(contributionID uint, prURL string) error {
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
	if err := tx.First(&contribution, contributionID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("contribution with ID %d not found: %w", contributionID, err)
	}

	if contribution.VerificationStatus != "unverified" {
		tx.Rollback()
		return fmt.Errorf("contribution %d is already %s", contributionID, contribution.VerificationStatus)
	}

	fmt.Printf("Simulating Git verification for PR: %s\n", prURL)
	now := time.Now()
	contribution.VerificationStatus = "auto_verified"
	contribution.AcceptedAt = &now
	if err := tx.Save(&contribution).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update contribution status: %w", err)
	}

	var task models.Task
	if err := tx.First(&task, contribution.TaskID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task with ID %d not found: %w", contribution.TaskID, err)
	}
	if err := tx.Model(&task).Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update task status to completed: %w", err)
	}

	reputationAward := 100
	if task.BountyAmount > 0 {
		reputationAward += int(task.BountyAmount / 10)
	}

	var contributor models.User
	if err := tx.First(&contributor, contribution.UserID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("contributor with ID %d not found: %w", contribution.UserID, err)
	}
	contributor.Ratings += reputationAward
	if err := tx.Save(&contributor).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update contributor reputation: %w", err)
	}

	repLog := models.ReputationEventLog{
		UserID:      contributor.ID,
		EventType:   ReputationEventContributionAccepted,
		ScoreChange: reputationAward,
		RelatedID:   &contribution.ID,
		Notes:       fmt.Sprintf("Accepted contribution for task '%s'", task.Title),
	}
	if err := tx.Create(&repLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to log reputation event: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ContributionService) RejectContribution(contributionID uint, reason string) error {
	result := db.DB.Model(&models.Contribution{}).
		Where("id = ?", contributionID).
		Update("verification_status", "rejected").
		Update("accepted_at", nil)

	if result.Error != nil {
		return fmt.Errorf("failed to reject contribution: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("contribution with ID %d not found", contributionID)
	}

	var contribution models.Contribution
	if err := db.DB.First(&contribution, contributionID).Error; err == nil {
		db.DB.Model(&models.Task{}).
			Where("id = ?", contribution.TaskID).
			Update("status", "claimed")
	}

	return nil
}

func (s *ContributionService) MentorEndorsements(mentorID, userID, relatedID uint, notes string) error {
	tx := db.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("user with ID %d not found: %w", userID, err)
	}

	endorsementScore := 20
	user.Ratings += endorsementScore
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user ratings for endorsement: %w", err)
	}

	repLog := models.ReputationEventLog{
		UserID:      user.ID,
		EventType:   ReputationEventMentorEndorsement,
		ScoreChange: endorsementScore,
		RelatedID:   &relatedID,
		Notes:       notes,
	}
	if err := tx.Create(&repLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to log mentor endorsement event: %w", err)
	}

	var mentor models.User
	if err := tx.First(&mentor, mentorID).Error; err == nil {
		mentor.Ratings += 5
		tx.Save(&mentor)
		mentorLog := models.ReputationEventLog{
			UserID:      mentor.ID,
			EventType:   ReputationEventManualAdjustment,
			ScoreChange: 5,
			Notes:       fmt.Sprintf("Mentored user %d for related ID %d", userID, relatedID),
		}
		tx.Create(&mentorLog)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction for endorsement: %w", err)
	}

	return nil
}