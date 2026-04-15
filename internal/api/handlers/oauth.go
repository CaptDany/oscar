package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/config"
	"github.com/oscar/oscar/internal/domain/tenant"
	"github.com/oscar/oscar/internal/domain/user"
	"github.com/oscar/oscar/pkg/crypto"
	"github.com/oscar/oscar/pkg/errs"
)

type OAuthHandler struct {
	userRepo     user.Repository
	tenantRepo   tenant.Repository
	roleRepo     user.RoleRepository
	crypto       *crypto.Crypto
	tokenManager *crypto.TokenManager
	emailClient  *EmailSender
	baseURL      string
	oauthConfig  *config.OAuthConfig
}

type OAuthProvider string

const (
	ProviderGoogle OAuthProvider = "google"
	ProviderApple  OAuthProvider = "apple"
)

type OAuthUserInfo struct {
	Email string
	Name  string
	Sub   string
}

func NewOAuthHandler(
	userRepo user.Repository,
	tenantRepo tenant.Repository,
	roleRepo user.RoleRepository,
	cryptoSvc *crypto.Crypto,
	tokenManager *crypto.TokenManager,
	emailClient *EmailSender,
	baseURL string,
	oauthConfig *config.OAuthConfig,
) *OAuthHandler {
	return &OAuthHandler{
		userRepo:     userRepo,
		tenantRepo:   tenantRepo,
		roleRepo:     roleRepo,
		crypto:       cryptoSvc,
		tokenManager: tokenManager,
		emailClient:  emailClient,
		baseURL:      baseURL,
		oauthConfig:  oauthConfig,
	}
}

func (h *OAuthHandler) GoogleLogin(c echo.Context) error {
	state, err := generateState()
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	redirectURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile&state=%s",
		h.oauthConfig.GoogleClientID,
		url.QueryEscape(fmt.Sprintf("%s/api/v1/auth/oauth/google/callback", h.baseURL)),
		state,
	)

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *OAuthHandler) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	storedState, _ := c.Cookie("oauth_state")

	if storedState == nil || storedState.Value != state {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid state parameter",
		})
	}

	c.SetCookie(&http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	tokenRes, err := h.exchangeGoogleToken(code)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	userInfo, err := h.getGoogleUserInfo(tokenRes.AccessToken)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return h.handleOAuthLogin(c, ProviderGoogle, userInfo)
}

func (h *OAuthHandler) AppleLogin(c echo.Context) error {
	state, err := generateState()
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	redirectURL := fmt.Sprintf(
		"https://appleid.apple.com/auth/authorize?client_id=%s&redirect_uri=%s&response_type=code%%20id_token&scope=email%%20name&state=%s",
		h.oauthConfig.AppleClientID,
		url.QueryEscape(fmt.Sprintf("%s/api/v1/auth/oauth/apple/callback", h.baseURL)),
		state,
	)

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *OAuthHandler) AppleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	storedState, _ := c.Cookie("oauth_state")

	if storedState == nil || storedState.Value != state {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid state parameter",
		})
	}

	c.SetCookie(&http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	userInfo, err := h.verifyAppleIDToken(c.QueryParam("id_token"))
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if code != "" {
		_, err := h.exchangeAppleToken(code)
		if err != nil {
			return errs.Internal(err).HTTPError(c)
		}
	}

	return h.handleOAuthLogin(c, ProviderApple, userInfo)
}

type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

func (h *OAuthHandler) exchangeGoogleToken(code string) (*googleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", h.oauthConfig.GoogleClientID)
	data.Set("client_secret", h.oauthConfig.GoogleClientSecret)
	data.Set("redirect_uri", fmt.Sprintf("%s/api/v1/auth/oauth/google/callback", h.baseURL))
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google token exchange failed: %s", string(body))
	}

	var tokenRes googleTokenResponse
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return nil, err
	}

	return &tokenRes, nil
}

func (h *OAuthHandler) getGoogleUserInfo(accessToken string) (*OAuthUserInfo, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo failed: %s", string(body))
	}

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		Email: userInfo.Email,
		Name:  userInfo.Name,
		Sub:   userInfo.ID,
	}, nil
}

func (h *OAuthHandler) exchangeAppleToken(code string) (map[string]interface{}, error) {
	clientSecret := h.generateAppleClientSecret()

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", h.oauthConfig.AppleClientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", fmt.Sprintf("%s/api/v1/auth/oauth/apple/callback", h.baseURL))
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://appleid.apple.com/auth/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apple token exchange failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (h *OAuthHandler) verifyAppleIDToken(idToken string) (*OAuthUserInfo, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims struct {
		Email string `json:"email"`
		Sub   string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		Email: claims.Email,
		Sub:   claims.Sub,
	}, nil
}

func (h *OAuthHandler) generateAppleClientSecret() string {
	return h.oauthConfig.AppleClientSecret
}

func (h *OAuthHandler) handleOAuthLogin(c echo.Context, provider OAuthProvider, userInfo *OAuthUserInfo) error {
	ctx := c.Request().Context()

	oauthUser, err := h.userRepo.GetByOAuthProvider(ctx, string(provider), userInfo.Sub)
	if err != nil && err != user.ErrOAuthUserNotFound {
		return errs.Internal(err).HTTPError(c)
	}

	var u *user.User
	if oauthUser != nil {
		u, err = h.userRepo.GetByID(ctx, oauthUser.UserID)
		if err != nil {
			return errs.Internal(err).HTTPError(c)
		}
	} else {
		u, err = h.userRepo.GetByEmail(ctx, uuid.Nil, userInfo.Email)
		if err != nil && err != user.ErrUserNotFound {
			return errs.Internal(err).HTTPError(c)
		}

		if u != nil {
			err = h.userRepo.LinkOAuth(ctx, u.ID, string(provider), userInfo.Sub)
			if err != nil {
				return errs.Internal(err).HTTPError(c)
			}
		} else {
			u, err = h.handleNewOAuthUser(ctx, provider, userInfo)
			if err != nil {
				return errs.Internal(err).HTTPError(c)
			}
		}
	}

	roleNames, _ := h.roleRepo.GetUserRoleNames(ctx, u.ID)

	payload := crypto.TokenPayload{
		UserID:   u.ID.String(),
		TenantID: u.TenantID.String(),
		Email:    u.Email,
		Roles:    roleNames,
	}

	tokens, err := h.tokenManager.GenerateTokenPair(payload, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	c.SetCookie(&http.Cookie{
		Name:     "oscar_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   int(tokens.ExpiresAt - time.Now().Unix()),
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	c.SetCookie(&http.Cookie{
		Name:     "oscar_refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"token":         tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user": map[string]interface{}{
			"id":         u.ID,
			"tenant_id":  u.TenantID,
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"roles":      roleNames,
		},
	})
}

func (h *OAuthHandler) handleNewOAuthUser(ctx context.Context, provider OAuthProvider, userInfo *OAuthUserInfo) (*user.User, error) {
	emailDomain := ""
	if idx := strings.Index(userInfo.Email, "@"); idx != -1 {
		emailDomain = userInfo.Email[idx+1:]
	}
	domain := strings.ReplaceAll(emailDomain, ".", "-")

	t, err := h.tenantRepo.GetBySlug(ctx, domain)
	if err != nil {
		createTenantReq := &tenant.CreateTenantRequest{
			Slug: domain,
			Name: domain,
		}
		t, err = h.tenantRepo.Create(ctx, createTenantReq)
		if err != nil {
			return nil, err
		}

		if err := h.tenantRepo.SeedRoles(ctx, t.ID); err != nil {
			return nil, err
		}
		if err := h.tenantRepo.SeedPipeline(ctx, t.ID); err != nil {
			return nil, err
		}
	}

	names := strings.SplitN(userInfo.Name, " ", 2)
	firstName := names[0]
	lastName := ""
	if len(names) > 1 {
		lastName = names[1]
	}

	now := time.Now()
	u, err := h.userRepo.CreateOAuthUser(ctx, &user.CreateOAuthUserRequest{
		TenantID:  t.ID,
		Email:     userInfo.Email,
		FirstName: firstName,
		LastName:  lastName,
	}, string(provider), userInfo.Sub)
	if err != nil {
		return nil, err
	}

	_ = u

	defaultRole, err := h.roleRepo.GetByName(ctx, t.ID, "Read Only")
	if err != nil {
		defaultRole, err = h.roleRepo.GetByName(ctx, t.ID, "Owner")
		if err != nil {
			return nil, err
		}
	}

	_ = h.roleRepo.AssignToUser(ctx, u.ID, []uuid.UUID{defaultRole.ID})

	if provider == ProviderGoogle || provider == ProviderApple {
		_ = h.userRepo.VerifyEmail(ctx, u.ID)
	}

	_ = now

	return u, nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type CreateOAuthUserRequest struct {
	Email     string
	FirstName string
	LastName  string
}
