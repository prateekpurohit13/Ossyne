package api

import (
	"fmt"
	"net/http"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"ossyne/internal/services"
	"strconv"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type UserHandler struct{}

func (h *UserHandler) CreateUser(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	result := db.DB.Create(&user)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
	}

	return c.JSON(http.StatusCreated, user)
}
//Development-only handler for creating claims without authentication
func (h *ClaimHandler) CreateClaimDev(c echo.Context) error {
	var req struct {
		TaskID uint `json:"task_id"`
		UserID uint `json:"user_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.TaskID == 0 || req.UserID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID and User ID are required for dev mode"})
	}

	claim := &models.Claim{
		TaskID: req.TaskID,
		UserID: req.UserID,
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	var task models.Task
	if err := tx.First(&task, claim.TaskID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch task"})
	}

	if task.Status != "open" {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task is not open for claiming"})
	}

	if err := tx.Create(claim).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create claim"})
	}

	task.Status = "claimed"
	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update task status"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	return c.JSON(http.StatusCreated, claim)
}

type ProjectHandler struct{}

func (h *ProjectHandler) CreateProject(c echo.Context) error {
	project := new(models.Project)
	if err := c.Bind(project); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Get user from context using the correct context key
	user, ok := c.Request().Context().Value(userContextKey).(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}
	project.OwnerID = user.ID

	result := db.DB.Create(&project)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
	}

	return c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) ListProjects(c echo.Context) error {
	var projects []models.Project
	db.DB.Find(&projects)
	return c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) ListUserProjects(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var projects []models.Project
	if err := db.DB.Where("owner_id = ?", userID).Find(&projects).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to fetch projects: %v", err)})
	}

	return c.JSON(http.StatusOK, projects)
}

type TaskHandler struct{}

func (h *TaskHandler) CreateTask(c echo.Context) error {
	task := new(models.Task)
	if err := c.Bind(task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if task.ProjectID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Project ID is required"})
	}

	var project models.Project
	if err := db.DB.First(&project, task.ProjectID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Project with ID %d not found", task.ProjectID)})
	}
	result := db.DB.Create(&task)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
	}

	return c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) ListTasks(c echo.Context) error {
	var tasks []models.Task
	projectIDStr := c.QueryParam("project_id")
	status := c.QueryParam("status")
	query := db.DB.Model(&models.Task{})

	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid project_id"})
		}
		query = query.Where("project_id = ?", projectID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Find(&tasks)
	return c.JSON(http.StatusOK, tasks)
}

type ClaimHandler struct{}

func (h *ClaimHandler) CreateClaim(c echo.Context) error {
	var req struct {
		TaskID uint `json:"task_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	user, ok := c.Request().Context().Value(userContextKey).(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	claim := &models.Claim{
		TaskID: req.TaskID,
		UserID: user.ID,
	}

	if claim.TaskID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID is required"})
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	var task models.Task
	if err := tx.First(&task, claim.TaskID).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Task with ID %d not found", claim.TaskID)})
	}
	if task.Status != "open" {
		tx.Rollback()
		return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("Task '%s' is not open for claims (current status: %s)", task.Title, task.Status)})
	}

	var existingClaim models.Claim
	if err := tx.Where("task_id = ? AND user_id = ?", claim.TaskID, claim.UserID).First(&existingClaim).Error; err == nil {
		tx.Rollback()
		return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("User %s has already claimed task '%s'", user.Username, task.Title)})
	}

	claim.Status = "pending"
	if err := tx.Create(&claim).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to create claim: %v", err)})
	}

	if err := tx.Model(&task).Update("status", "claimed").Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to update task status: %v", err)})
	}

	tx.Commit()
	return c.JSON(http.StatusCreated, claim)
}

func (h *ClaimHandler) ListClaims(c echo.Context) error {
	var claims []models.Claim
	userIDStr := c.QueryParam("user_id")
	taskIDStr := c.QueryParam("task_id")
	status := c.QueryParam("status")
	query := db.DB.Model(&models.Claim{})

	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id"})
		}
		query = query.Where("user_id = ?", userID)
	}
	if taskIDStr != "" {
		taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid task_id"})
		}
		query = query.Where("task_id = ?", taskID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Preload("Task").Preload("User").Find(&claims)
	return c.JSON(http.StatusOK, claims)
}

type ContributionHandler struct {
	Service *services.ContributionService
}

func (h *ContributionHandler) CreateContribution(c echo.Context) error {
	var req struct {
		TaskID uint   `json:"task_id"`
		PRURL  string `json:"pr_url"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	user, ok := c.Request().Context().Value(userContextKey).(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	contribution := &models.Contribution{
		TaskID: req.TaskID,
		UserID: user.ID,
		PRURL:  req.PRURL,
	}

	if contribution.TaskID == 0 || contribution.PRURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID and PR URL are required"})
	}

	var task models.Task
	if err := db.DB.First(&task, contribution.TaskID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Task with ID %d not found", contribution.TaskID)})
	}
	if task.Status != "claimed" && task.Status != "in_progress" {
		return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("Task '%s' is not in a state to accept contributions (current status: %s)", task.Title, task.Status)})
	}

	var existingContribution models.Contribution
	if err := db.DB.Where("task_id = ? AND user_id = ?", contribution.TaskID, contribution.UserID).First(&existingContribution).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("User %s has already submitted a contribution for task '%s'", user.Username, task.Title)})
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	contribution.VerificationStatus = "unverified"
	if err := tx.Create(&contribution).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to create contribution: %v", err)})
	}
	if err := tx.Model(&task).Update("status", "submitted").Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to update task status: %v", err)})
	}
	tx.Commit()
	return c.JSON(http.StatusCreated, contribution)
}

// Development-only handler for creating contributions without authentication
func (h *ContributionHandler) CreateContributionDev(c echo.Context) error {
	var req struct {
		TaskID uint   `json:"task_id"`
		UserID uint   `json:"user_id"`
		PRURL  string `json:"pr_url"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.TaskID == 0 || req.UserID == 0 || req.PRURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID, User ID, and PR URL are required for dev mode"})
	}

	contribution := &models.Contribution{
		TaskID: req.TaskID,
		UserID: req.UserID,
		PRURL:  req.PRURL,
	}

	var task models.Task
	if err := db.DB.First(&task, contribution.TaskID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Task with ID %d not found", contribution.TaskID)})
	}
	if task.Status != "claimed" && task.Status != "in_progress" {
		return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("Task '%s' is not in a state to accept contributions (current status: %s)", task.Title, task.Status)})
	}

	var existingContribution models.Contribution
	if err := db.DB.Where("task_id = ? AND user_id = ?", contribution.TaskID, contribution.UserID).First(&existingContribution).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("User %d has already submitted a contribution for task '%s'", req.UserID, task.Title)})
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	contribution.VerificationStatus = "unverified"
	if err := tx.Create(&contribution).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to create contribution: %v", err)})
	}
	if err := tx.Model(&task).Update("status", "submitted").Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to update task status: %v", err)})
	}
	tx.Commit()
	return c.JSON(http.StatusCreated, contribution)
}

func (h *ContributionHandler) ListContributions(c echo.Context) error {
	var contributions []models.Contribution
	userIDStr := c.QueryParam("user_id")
	taskIDStr := c.QueryParam("task_id")
	status := c.QueryParam("verification_status")

	query := db.DB.Model(&models.Contribution{})

	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id"})
		}
		query = query.Where("user_id = ?", userID)
	}
	if taskIDStr != "" {
		taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid task_id"})
		}
		query = query.Where("task_id = ?", taskID)
	}
	if status != "" {
		query = query.Where("verification_status = ?", status)
	}

	query.Preload("Task").Preload("User").Find(&contributions)
	return c.JSON(http.StatusOK, contributions)
}

func (h *ContributionHandler) AcceptContribution(c echo.Context) error {
	contributionIDStr := c.Param("id")
	contributionID, err := strconv.ParseUint(contributionIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contribution ID"})
	}
	var contribution models.Contribution
	if err := db.DB.First(&contribution, contributionID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Contribution not found"})
	}

	if h.Service == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Contribution service not initialized"})
	}
	if err := h.Service.VerifyAndAcceptContribution(uint(contributionID), contribution.PRURL); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to accept contribution: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Contribution accepted and contributor credited"})
}

func (h *ContributionHandler) RejectContribution(c echo.Context) error {
	contributionIDStr := c.Param("id")
	contributionID, err := strconv.ParseUint(contributionIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid contribution ID"})
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	if h.Service == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Contribution service not initialized"})
	}
	if err := h.Service.RejectContribution(uint(contributionID), req.Reason); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to reject contribution: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Contribution rejected"})
}

type MentorHandler struct {
	Service *services.ContributionService
}

func (h *MentorHandler) EndorseUser(c echo.Context) error {
	var req struct {
		MentorID  uint   `json:"mentor_id"`
		UserID    uint   `json:"user_id"`
		RelatedID uint   `json:"related_id"`
		Notes     string `json:"notes"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	if req.MentorID == 0 || req.UserID == 0 || req.RelatedID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Mentor ID, User ID, and Related ID are required"})
	}
	if h.Service == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Contribution service not initialized"})
	}
	if err := h.Service.MentorEndorsements(req.MentorID, req.UserID, req.RelatedID, req.Notes); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to endorse user: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User endorsed successfully!"})
}

type SkillHandler struct{}

func (h *SkillHandler) CreateSkill(c echo.Context) error {
	skill := new(models.Skill)
	if err := c.Bind(skill); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if skill.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Skill name is required"})
	}

	result := db.DB.Create(&skill)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
	}
	return c.JSON(http.StatusCreated, skill)
}

func (h *SkillHandler) ListSkills(c echo.Context) error {
	var skills []models.Skill
	db.DB.Find(&skills)
	return c.JSON(http.StatusOK, skills)
}

type UserSkillHandler struct{}

func (h *UserSkillHandler) AddUserSkill(c echo.Context) error {
	userSkill := new(models.UserSkill)
	if err := c.Bind(userSkill); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	if userSkill.UserID == 0 || userSkill.SkillID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID and Skill ID are required"})
	}

	result := db.DB.Create(&userSkill)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
	}
	return c.JSON(http.StatusCreated, userSkill)
}

func (h *UserSkillHandler) ListUserSkills(c echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var userSkills []models.UserSkill
	db.DB.Preload("Skill").Where("user_id = ?", userID).Find(&userSkills)
	return c.JSON(http.StatusOK, userSkills)
}

func (h *UserHandler) GetUser(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var user models.User
	if err := db.DB.Preload("UserSkills.Skill").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("User with ID %d not found", userID)})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to fetch user: %v", err)})
	}
	return c.JSON(http.StatusOK, user)
}

type PaymentHandler struct {
	Service *services.PaymentService
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		Service: services.NewPaymentService(),
	}
}

func (h *PaymentHandler) FundTaskBounty(c echo.Context) error {
	var req struct {
		TaskID   uint    `json:"task_id"`
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	user, ok := c.Request().Context().Value(userContextKey).(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	if req.TaskID == 0 || req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID and positive Amount are required"})
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if err := h.Service.FundTaskBounty(req.TaskID, user.ID, req.Amount, req.Currency); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to fund task bounty: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Task bounty funded and escrowed successfully!"})
}

func (h *PaymentHandler) RefundTaskBounty(c echo.Context) error {
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid task ID"})
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	if err := h.Service.RefundTaskBounty(uint(taskID), req.Reason); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to refund task bounty: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Task bounty refunded successfully!"})
}

func (h *PaymentHandler) GetUserPayments(c echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	payments, err := h.Service.GetUserPayments(uint(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to retrieve user payments: %v", err)})
	}
	return c.JSON(http.StatusOK, payments)
}

func (h *PaymentHandler) GetMyPayments(c echo.Context) error {
	user, ok := c.Request().Context().Value(userContextKey).(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	payments, err := h.Service.GetUserPayments(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to retrieve user payments: %v", err)})
	}
	return c.JSON(http.StatusOK, payments)
}

func (h *UserHandler) GetMe(c echo.Context) error {
	user, ok := c.Request().Context().Value(userContextKey).(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}
	return c.JSON(http.StatusOK, user)
}