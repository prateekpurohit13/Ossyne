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

func (h *UserHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetMe(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found in context"})
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
	contributionService.PaymentService = paymentService
	authService := services.NewAuthService(cfg)

	userHandler := &api.UserHandler{}
	projectHandler := &api.ProjectHandler{}
	taskHandler := &api.TaskHandler{}
	claimHandler := &api.ClaimHandler{}
	contributionHandler := &api.ContributionHandler{Service: contributionService}
	mentorHandler := &api.MentorHandler{Service: contributionService}
	skillHandler := &api.SkillHandler{}
	userSkillHandler := &api.UserSkillHandler{}
	paymentHandler := api.NewPaymentHandler()

	e.GET("/auth/github", echo.WrapHandler(http.HandlerFunc(authService.HandleGitHubLogin)))
	e.GET("/auth/github/callback", echo.WrapHandler(http.HandlerFunc(authService.HandleGitHubCallback)))
	e.GET("/tasks", taskHandler.ListTasks)//keeping this public for browsing
	e.GET("/projects", projectHandler.ListProjects)
	//Authenticated Routes
	apiGroup := e.Group("/api")
	apiGroup.Use(api.AuthMiddleware)
	apiGroup.POST("/projects", projectHandler.CreateProject)
	apiGroup.POST("/tasks", taskHandler.CreateTask)
	apiGroup.POST("/claims", claimHandler.CreateClaim)
	apiGroup.POST("/contributions", contributionHandler.CreateContribution)
	apiGroup.PUT("/contributions/:id/accept", contributionHandler.AcceptContribution)
	apiGroup.PUT("/contributions/:id/reject", contributionHandler.RejectContribution)
	apiGroup.POST("/mentor/endorse", mentorHandler.EndorseUser)
	apiGroup.POST("/bounties/fund", paymentHandler.FundTaskBounty)
	apiGroup.PUT("/bounties/refund/:id", paymentHandler.RefundTaskBounty)
	apiGroup.GET("/users/me", userHandler.GetMe)
	apiGroup.GET("/users/me/payments", paymentHandler.GetMyPayments)
	apiGroup.GET("/users/:user_id/payments", paymentHandler.GetUserPayments)
	adminGroup := apiGroup.Group("/admin")
	adminGroup.POST("/skills", skillHandler.CreateSkill)
	adminGroup.POST("/users/skills", userSkillHandler.AddUserSkill)

	//Public routes
	e.GET("/users/:id", userHandler.GetUser)
	e.GET("/users/:id/projects", projectHandler.ListUserProjects)
	e.GET("/skills", skillHandler.ListSkills)
	e.GET("/users/:user_id/skills", userSkillHandler.ListUserSkills)
	e.GET("/claims", claimHandler.ListClaims)
	e.GET("/contributions", contributionHandler.ListContributions)

	//Development-only routes
	devGroup := e.Group("/dev")
	devGroup.POST("/users/create", userHandler.CreateUser)
	devGroup.POST("/claims", claimHandler.CreateClaimDev)
	devGroup.POST("/contributions", contributionHandler.CreateContributionDev)
	devGroup.GET("/users/:user_id/payments", paymentHandler.GetUserPayments)

	if err := e.Start(":" + cfg.ServerPort); err != nil {
		e.Logger.Fatal(err)
	}
}