package api

import (
	"fmt"
	"net/http"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"strconv"

	"github.com/labstack/echo/v4"
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

type ProjectHandler struct{}

func (h *ProjectHandler) CreateProject(c echo.Context) error {
	project := new(models.Project)
	if err := c.Bind(project); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if project.OwnerID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Owner ID is required"})
	}

	var owner models.User
	if err := db.DB.First(&owner, project.OwnerID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Owner with ID %d not found", project.OwnerID)})
	}

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
	claim := new(models.Claim)
	if err := c.Bind(claim); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if claim.TaskID == 0 || claim.UserID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID and User ID are required"})
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

	var user models.User
	if err := tx.First(&user, claim.UserID).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("User with ID %d not found", claim.UserID)})
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

type ContributionHandler struct{}

func (h *ContributionHandler) CreateContribution(c echo.Context) error {
	contribution := new(models.Contribution)
	if err := c.Bind(contribution); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if contribution.TaskID == 0 || contribution.UserID == 0 || contribution.PRURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID, User ID, and PR URL are required"})
	}

	var user models.User
	if err := db.DB.First(&user, contribution.UserID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("User with ID %d not found", contribution.UserID)})
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