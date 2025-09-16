package main

import (
	"fmt"
	"net/http"
	"ossyne/internal/api"
	"ossyne/internal/config"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"ossyne/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type UserHandler struct {
	// will add a database dependency here later
}

type ContributionHandler struct {
	Service *services.ContributionService
}

func (h *UserHandler) createUser(c echo.Context) error {
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

func (h *UserHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic(fmt.Sprintf("cannot load config: %v", err))
	}
	if err := db.Init(cfg); err != nil {
		panic(fmt.Sprintf("cannot connect to db: %v", err))
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	paymentService := services.NewPaymentService()
	contributionService := services.NewContributionService(paymentService)

	userHandler := &UserHandler{}
	projectHandler := &api.ProjectHandler{}
	taskHandler := &api.TaskHandler{}
	claimHandler := &api.ClaimHandler{}
	contributionHandler := &api.ContributionHandler{Service: contributionService}
	mentorHandler := &api.MentorHandler{Service: contributionService}
	skillHandler := &api.SkillHandler{}
	userSkillHandler := &api.UserSkillHandler{}
	paymentHandler := api.NewPaymentHandler()

	e.POST("/users", userHandler.createUser)
	e.GET("/users/:id", userHandler.GetUser)
	e.GET("/projects", projectHandler.ListProjects)
	e.GET("/users/:id/projects", projectHandler.ListUserProjects)
	e.POST("/projects", projectHandler.CreateProject)
	e.GET("/tasks", taskHandler.ListTasks)
	e.POST("/tasks", taskHandler.CreateTask)
	e.POST("/claims", claimHandler.CreateClaim)
	e.GET("/claims", claimHandler.ListClaims)
	e.POST("/contributions", contributionHandler.CreateContribution)
	e.GET("/contributions", contributionHandler.ListContributions)
	e.PUT("/contributions/:id/accept", contributionHandler.AcceptContribution)
	e.PUT("/contributions/:id/reject", contributionHandler.RejectContribution)
	e.POST("/mentor/endorse", mentorHandler.EndorseUser)
	e.POST("/skills", skillHandler.CreateSkill)
	e.GET("/skills", skillHandler.ListSkills)
	e.POST("/users/skills", userSkillHandler.AddUserSkill)
	e.GET("/users/:user_id/skills", userSkillHandler.ListUserSkills)
	e.POST("/bounties/fund", paymentHandler.FundTaskBounty)
	e.PUT("/bounties/refund/:id", paymentHandler.RefundTaskBounty)
	e.GET("/users/:user_id/payments", paymentHandler.GetUserPayments)

	if err := e.Start(":" + cfg.ServerPort); err != nil {
		e.Logger.Fatal(err)
	}
}