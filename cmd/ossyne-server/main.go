package main

import (
	"fmt"
	"net/http"
	"ossyne/internal/config"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"ossyne/internal/api"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type UserHandler struct {
	// will add a database dependency here later
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

	userHandler := &UserHandler{}
	projectHandler := &api.ProjectHandler{}
	taskHandler := &api.TaskHandler{}
	claimHandler := &api.ClaimHandler{}
	contributionHandler := &api.ContributionHandler{}

	e.POST("/users", userHandler.createUser)
	e.GET("/projects", projectHandler.ListProjects)
	e.POST("/projects", projectHandler.CreateProject)
	e.GET("/tasks", taskHandler.ListTasks)
	e.POST("/tasks", taskHandler.CreateTask)
	e.POST("/claims", claimHandler.CreateClaim)
	e.GET("/claims", claimHandler.ListClaims)
	e.POST("/contributions", contributionHandler.CreateContribution)
	e.GET("/contributions", contributionHandler.ListContributions)

	fmt.Printf("Starting server on port %s\n", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		e.Logger.Fatal(err)
	}
}