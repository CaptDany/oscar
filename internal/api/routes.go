package api

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/api/handlers"
	"github.com/oscar/oscar/internal/api/middleware"
	"github.com/oscar/oscar/pkg/errs"
)

type Handlers struct {
	Auth         *handlers.AuthHandler
	OAuth        *handlers.OAuthHandler
	Person       *handlers.PersonHandler
	Company      *handlers.CompanyHandler
	Deal         *handlers.DealHandler
	Pipeline     *handlers.PipelineHandler
	Activity     *handlers.ActivityHandler
	User         *handlers.UserHandler
	Upload       *handlers.UploadHandler
	Notification *handlers.NotificationHandler
	Team         *handlers.TeamHandler
	Product      *handlers.ProductHandler
	Settings     *handlers.SettingsHandler
	Invitation   *handlers.InvitationHandler
	CustomField  *handlers.CustomFieldHandler
	LineItem     *handlers.LineItemHandler
	AuditLog     *handlers.AuditLogHandler
	APIKey       *handlers.APIKeyHandler
	Attachment   *handlers.AttachmentHandler
}

func (s *Server) SetupRoutes(h *Handlers, authMiddleware echo.MiddlewareFunc, authMiddlewareWithTenant echo.MiddlewareFunc, rateLimiter *middleware.InMemoryRateLimiter) {
	api := s.Group("/api/v1", middleware.RateLimitMiddleware(rateLimiter))

	api.POST("/auth/register", h.Auth.Register)
	api.POST("/auth/login", h.Auth.Login)
	api.POST("/auth/refresh", h.Auth.Refresh)
	api.GET("/auth/verify-email/:token", h.Auth.VerifyEmail)
	api.POST("/auth/resend-verification", h.Auth.ResendVerification)

	api.GET("/auth/oauth/google", h.OAuth.GoogleLogin)
	api.GET("/auth/oauth/google/callback", h.OAuth.GoogleCallback)
	api.GET("/auth/oauth/apple", h.OAuth.AppleLogin)
	api.GET("/auth/oauth/apple/callback", h.OAuth.AppleCallback)

	api.GET("/invitations/:token/validate", h.Invitation.Validate)

	api.GET("/avatar/:user_id", h.Upload.GetAvatarURL)

	auth := api.Group("", authMiddleware)
	auth.POST("/auth/logout", h.Auth.Logout)
	auth.GET("/auth/me", h.Auth.Me)
	auth.PATCH("/users/me", h.User.UpdateMe)
	auth.GET("/export", h.User.Export)
	auth.POST("/upload/avatar", h.Upload.GetAvatarPresignedURL)
	auth.POST("/upload/avatar/confirm", h.Upload.ConfirmAvatarUpload)

	tenantScoped := auth.Group("", authMiddlewareWithTenant)

	branding := tenantScoped.Group("/upload/branding")
	branding.POST("/presigned", h.Upload.GetBrandingAssetPresignedURL)
	branding.POST("/confirm", h.Upload.ConfirmBrandingAssetUpload)
	branding.POST("/delete", h.Upload.DeleteBrandingAsset)
	branding.GET("/asset", h.Upload.GetBrandingAsset)

	settings := tenantScoped.Group("/settings")
	settings.GET("", h.Settings.GetSettings)
	settings.PATCH("", h.Settings.UpdateSettings)

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
	deals.GET("/:id/line-items", h.LineItem.ListByDeal)
	deals.POST("/:id/line-items", h.LineItem.Create)
	deals.PATCH("/:id/line-items/:line_item_id", h.LineItem.Update)
	deals.DELETE("/:id/line-items/:line_item_id", h.LineItem.Delete)

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
	users.PATCH("/:id", h.User.Update)
	users.PUT("/:id/roles", h.User.UpdateRoles, RequirePermission("users", "edit"))

	notifications := tenantScoped.Group("/notifications")
	notifications.GET("", h.Notification.List)
	notifications.GET("/count", h.Notification.CountUnread)
	notifications.GET("/:id", h.Notification.Get)
	notifications.POST("/:id/read", h.Notification.MarkAsRead)
	notifications.POST("/read-all", h.Notification.MarkAllAsRead)
	notifications.DELETE("/:id", h.Notification.Delete)

	teams := tenantScoped.Group("/teams")
	teams.GET("", h.Team.List)
	teams.POST("", h.Team.Create)
	teams.GET("/:id", h.Team.Get)
	teams.PATCH("/:id", h.Team.Update)
	teams.DELETE("/:id", h.Team.Delete)
	teams.GET("/:id/members", h.Team.ListMembers)
	teams.POST("/:id/members", h.Team.AddMember)
	teams.DELETE("/:id/members/:user_id", h.Team.RemoveMember)
	teams.POST("/:id/lead/:user_id", h.Team.SetLead)

	products := tenantScoped.Group("/products")
	products.GET("", h.Product.List)
	products.POST("", h.Product.Create)
	products.GET("/:id", h.Product.Get)
	products.PATCH("/:id", h.Product.Update)
	products.DELETE("/:id", h.Product.Delete)

	invitations := tenantScoped.Group("/invitations")
	invitations.GET("", h.Invitation.List, RequirePermission("users", "edit"))
	invitations.POST("", h.Invitation.Create, RequirePermission("users", "edit"))
	invitations.DELETE("/:id", h.Invitation.Delete, RequirePermission("users", "edit"))

	customFields := tenantScoped.Group("/custom-fields")
	customFields.GET("", h.CustomField.List, RequirePermission("custom_fields", "view"))
	customFields.POST("", h.CustomField.Create, RequirePermission("custom_fields", "edit"))
	customFields.POST("/reorder", h.CustomField.Reorder, RequirePermission("custom_fields", "edit"))
	customFields.GET("/:id", h.CustomField.Get, RequirePermission("custom_fields", "view"))
	customFields.PATCH("/:id", h.CustomField.Update, RequirePermission("custom_fields", "edit"))
	customFields.DELETE("/:id", h.CustomField.Delete, RequirePermission("custom_fields", "edit"))

	apiKeys := tenantScoped.Group("/api-keys")
	apiKeys.GET("", h.APIKey.List, RequirePermission("api_keys", "view"))
	apiKeys.POST("", h.APIKey.Create, RequirePermission("api_keys", "edit"))
	apiKeys.DELETE("/:id", h.APIKey.Delete, RequirePermission("api_keys", "edit"))

	attachments := tenantScoped.Group("/attachments")
	attachments.POST("/presigned", h.Attachment.GetPresignedURL, RequirePermission("attachments", "edit"))
	attachments.GET("/:entity_type/:entity_id", h.Attachment.ListByEntity, RequirePermission("attachments", "view"))
	attachments.DELETE("/:id", h.Attachment.Delete, RequirePermission("attachments", "edit"))

	auditLogs := tenantScoped.Group("/audit-logs")
	auditLogs.GET("", h.AuditLog.List, RequirePermission("audit_logs", "view"))
	auditLogs.GET("/:id", h.AuditLog.Get, RequirePermission("audit_logs", "view"))
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
			"persons":       "all",
			"companies":     "all",
			"deals":         "all",
			"activities":    "all",
			"settings":      "all",
			"users":         "all",
			"custom_fields": "all",
			"audit_logs":    "all",
			"api_keys":      "all",
			"attachments":   "all",
		},
		"Admin": {
			"persons":       "all",
			"companies":     "all",
			"deals":         "all",
			"activities":    "all",
			"settings":      "all",
			"users":         "all",
			"custom_fields": "all",
			"audit_logs":    "all",
			"api_keys":      "all",
			"attachments":   "all",
		},
		"Manager": {
			"persons":       "team",
			"companies":     "team",
			"deals":         "team",
			"activities":    "team",
			"custom_fields": "all",
			"audit_logs":    "team",
			"api_keys":      "none",
			"attachments":   "all",
		},
		"Sales Rep": {
			"persons":       "own",
			"companies":     "own",
			"deals":         "own",
			"activities":    "own",
			"custom_fields": "own",
			"audit_logs":    "own",
			"api_keys":      "none",
			"attachments":   "own",
		},
		"Read Only": {
			"persons":       "team",
			"companies":     "team",
			"deals":         "team",
			"activities":    "team",
			"custom_fields": "team",
			"audit_logs":    "team",
			"api_keys":      "none",
			"attachments":   "team",
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
