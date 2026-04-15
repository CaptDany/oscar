package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/db/repositories"
	"github.com/oscar/oscar/internal/domain/tenant"
	"github.com/oscar/oscar/pkg/errs"
)

type SettingsHandler struct {
	tenantRepo   *repositories.TenantRepository
	brandingRepo *repositories.BrandingRepository
}

func NewSettingsHandler(tenantRepo *repositories.TenantRepository, brandingRepo *repositories.BrandingRepository) *SettingsHandler {
	return &SettingsHandler{
		tenantRepo:   tenantRepo,
		brandingRepo: brandingRepo,
	}
}

func (h *SettingsHandler) GetSettings(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	tenantData, err := h.tenantRepo.GetByID(c.Request().Context(), tenantID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	var settings map[string]interface{}
	if tenantData.Settings != nil {
		json.Unmarshal(tenantData.Settings, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}

	branding, _ := h.brandingRepo.Get(c.Request().Context(), tenantID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"name":            tenantData.Name,
			"currency":        settings["currency"],
			"timezone":        settings["timezone"],
			"primary_color":   branding.PrimaryColor,
			"secondary_color": branding.SecondaryColor,
			"accent_color":    branding.AccentColor,
			"font_family":     branding.FontFamily,
			"logo_light_url":  branding.LogoLightURL,
			"logo_dark_url":   branding.LogoDarkURL,
			"favicon_url":     branding.FaviconURL,
		},
	})
}

func (h *SettingsHandler) UpdateSettings(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	roles := c.Get("roles").([]string)

	isOwnerOrAdmin := false
	for _, role := range roles {
		if role == "Owner" || role == "Admin" {
			isOwnerOrAdmin = true
			break
		}
	}
	if !isOwnerOrAdmin {
		return errs.Forbidden("Only Owner or Admin can update settings").HTTPError(c)
	}

	var req struct {
		Name           *string `json:"name"`
		Currency       *string `json:"currency"`
		Timezone       *string `json:"timezone"`
		PrimaryColor   *string `json:"primary_color"`
		SecondaryColor *string `json:"secondary_color"`
		AccentColor    *string `json:"accent_color"`
		FontFamily     *string `json:"font_family"`
		LogoLightURL   *string `json:"logo_light_url"`
		LogoDarkURL    *string `json:"logo_dark_url"`
		FaviconURL     *string `json:"favicon_url"`
	}

	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	tenantData, err := h.tenantRepo.GetByID(c.Request().Context(), tenantID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	settings := make(map[string]interface{})
	if tenantData.Settings != nil {
		json.Unmarshal(tenantData.Settings, &settings)
	}

	if req.Currency != nil {
		settings["currency"] = *req.Currency
	}
	if req.Timezone != nil {
		settings["timezone"] = *req.Timezone
	}

	settingsJSON, _ := json.Marshal(settings)

	_, err = h.tenantRepo.Update(c.Request().Context(), tenantID, &tenant.UpdateTenantRequest{
		Name:     req.Name,
		Settings: settingsJSON,
	})
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if req.PrimaryColor != nil || req.SecondaryColor != nil || req.AccentColor != nil ||
		req.FontFamily != nil || req.LogoLightURL != nil || req.LogoDarkURL != nil || req.FaviconURL != nil {
		_, err = h.brandingRepo.Update(c.Request().Context(), tenantID, &tenant.UpdateBrandingRequest{
			PrimaryColor:   req.PrimaryColor,
			SecondaryColor: req.SecondaryColor,
			AccentColor:    req.AccentColor,
			FontFamily:     req.FontFamily,
			LogoLightURL:   req.LogoLightURL,
			LogoDarkURL:    req.LogoDarkURL,
			FaviconURL:     req.FaviconURL,
		})
		if err != nil {
			return errs.Internal(err).HTTPError(c)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Settings updated successfully",
	})
}
