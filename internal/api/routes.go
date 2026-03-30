package api

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/api/handlers"
	"github.com/oscar/oscar/internal/api/middleware"
	"github.com/oscar/oscar/pkg/errs"
)

type Handlers struct {
	Auth     *handlers.AuthHandler
	Person   *handlers.PersonHandler
	Company  *handlers.CompanyHandler
	Deal     *handlers.DealHandler
	Pipeline *handlers.PipelineHandler
	Activity *handlers.ActivityHandler
	User     *handlers.UserHandler
	Upload   *handlers.UploadHandler
}

func (s *Server) SetupRoutes(h *Handlers, authMiddleware echo.MiddlewareFunc, authMiddlewareWithTenant echo.MiddlewareFunc, rateLimiter *middleware.InMemoryRateLimiter) {
	api := s.Group("/api/v1", middleware.RateLimitMiddleware(rateLimiter))

	api.POST("/auth/register", h.Auth.Register)
	api.POST("/auth/login", h.Auth.Login)
	api.POST("/auth/refresh", h.Auth.Refresh)

	auth := api.Group("", authMiddleware)
	auth.POST("/auth/logout", h.Auth.Logout)
	auth.GET("/auth/me", h.Auth.Me)
	auth.POST("/upload/avatar", h.Upload.GetAvatarPresignedURL)
	auth.POST("/upload/avatar/confirm", h.Upload.ConfirmAvatarUpload)

	tenantScoped := auth.Group("", authMiddlewareWithTenant)
	tenantScoped.GET("/avatar/:user_id", h.Upload.GetAvatarURL)

	persons := tenantScoped.Group("/persons")
	persons.GET("", h.Person.List)
	persons.POST("", h.Person.Create)
	persons.GET("/:id", h.Person.Get)
	persons.PATCH("/:id", h.Person.Update)
	persons.DELETE("/:id", h.Person.Delete)
	persons.POST("/:id/convert", h.Person.Convert)
	persons.POST("/:id/tags", h.Person.AddTag)
	persons.DELETE("/:id/tags", h.Person.RemoveTag)
	persons.GET("/search", h.Person.Search)

	companies := tenantScoped.Group("/companies")
	companies.GET("", h.Company.List)
	companies.POST("", h.Company.Create)
	companies.GET("/:id", h.Company.Get)
	companies.PATCH("/:id", h.Company.Update)
	companies.DELETE("/:id", h.Company.Delete)

	pipelines := tenantScoped.Group("/pipelines")
	pipelines.GET("", h.Pipeline.List)
	pipelines.POST("", h.Pipeline.Create)
	pipelines.GET("/:id", h.Pipeline.Get)
	pipelines.PATCH("/:id", h.Pipeline.Update)
	pipelines.DELETE("/:id", h.Pipeline.Delete)
	pipelines.GET("/:id/stages", h.Pipeline.ListStages)
	pipelines.POST("/:id/stages", h.Pipeline.CreateStage)
	pipelines.PATCH("/:id/stages/reorder", h.Pipeline.ReorderStages)
	pipelines.PATCH("/:id/stages/:stage_id", h.Pipeline.UpdateStage)
	pipelines.DELETE("/:id/stages/:stage_id", h.Pipeline.DeleteStage)

	deals := tenantScoped.Group("/deals")
	deals.GET("", h.Deal.List)
	deals.GET("/kanban", h.Deal.Kanban)
	deals.POST("", h.Deal.Create)
	deals.GET("/:id", h.Deal.Get)
	deals.PATCH("/:id", h.Deal.Update)
	deals.DELETE("/:id", h.Deal.Delete)
	deals.PATCH("/:id/stage", h.Deal.MoveStage)
	deals.POST("/:id/win", h.Deal.Win)
	deals.POST("/:id/lose", h.Deal.Lose)

	activities := tenantScoped.Group("/activities")
	activities.GET("", h.Activity.List)
	activities.POST("", h.Activity.Create)
	activities.GET("/:id", h.Activity.Get)
	activities.PATCH("/:id", h.Activity.Update)
	activities.POST("/:id/complete", h.Activity.Complete)
	activities.POST("/:id/uncomplete", h.Activity.Uncomplete)
	activities.DELETE("/:id", h.Activity.Delete)

	tenantScoped.GET("/timeline", h.Activity.Timeline)

	users := tenantScoped.Group("/users")
	users.GET("", h.User.List, RequirePermission("users", "view"))
	users.GET("/:id", h.User.Get, RequirePermission("users", "view"))
	users.PUT("/:id/roles", h.User.UpdateRoles, RequirePermission("users", "edit"))
}

func GetTenantID(c echo.Context) uuid.UUID {
	if id, ok := c.Get("tenant_id").(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetUserID(c echo.Context) uuid.UUID {
	if id, ok := c.Get("user_id").(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetRoles(c echo.Context) []string {
	if roles, ok := c.Get("roles").([]string); ok {
		return roles
	}
	return nil
}

func RequirePermission(resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles := GetRoles(c)
			if roles == nil {
				return errs.PermissionDenied(resource, action)
			}

			canAccess := false
			for _, role := range roles {
				if hasPermission(role, resource, action) {
					canAccess = true
					break
				}
			}

			if !canAccess {
				return errs.PermissionDenied(resource, action)
			}

			return next(c)
		}
	}
}

func hasPermission(role, resource, action string) bool {
	permissions := map[string]map[string]string{
		"Owner": {
			"persons":    "all",
			"companies":  "all",
			"deals":      "all",
			"activities": "all",
			"settings":   "all",
			"users":      "all",
		},
		"Admin": {
			"persons":    "all",
			"companies":  "all",
			"deals":      "all",
			"activities": "all",
			"settings":   "all",
			"users":      "all",
		},
		"Manager": {
			"persons":    "team",
			"companies":  "team",
			"deals":      "team",
			"activities": "team",
		},
		"Sales Rep": {
			"persons":    "own",
			"companies":  "own",
			"deals":      "own",
			"activities": "own",
		},
		"Read Only": {
			"persons":    "team",
			"companies":  "team",
			"deals":      "team",
			"activities": "team",
		},
	}

	rolePerms, ok := permissions[role]
	if !ok {
		return false
	}

	scope, ok := rolePerms[resource]
	if !ok {
		return false
	}

	switch action {
	case "view":
		return scope == "all" || scope == "team" || scope == "own"
	case "create", "edit", "delete", "export":
		return scope == "all" || scope == "team" || scope == "own"
	default:
		return scope == "all"
	}
}
