package crypto

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/o1egl/paseto"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type TokenPayload struct {
	TokenType TokenType `json:"t"`
	UserID    string    `json:"uid"`
	TenantID  string    `json:"tid"`
	Email     string    `json:"em"`
	Roles     []string  `json:"roles"`
	Exp       int64     `json:"exp"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

type TokenManager struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewTokenManager(secret []byte) *TokenManager {
	return &TokenManager{
		paseto:       paseto.NewV2(),
		symmetricKey: secret,
	}
}

func (tm *TokenManager) GenerateTokenPair(payload TokenPayload, accessTTL, refreshTTL time.Duration) (*TokenPair, error) {
	accessToken, err := tm.generateToken(payload, accessTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	payload.TokenType = TokenTypeRefresh
	refreshToken, err := tm.generateToken(payload, refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(accessTTL).Unix(),
		TokenType:    "Bearer",
	}, nil
}

func (tm *TokenManager) GenerateAccessToken(payload TokenPayload, ttl time.Duration) (string, error) {
	return tm.generateToken(payload, ttl)
}

func (tm *TokenManager) generateToken(payload TokenPayload, ttl time.Duration) (string, error) {
	now := time.Now()
	expiration := now.Add(ttl)

	jsonToken := paseto.JSONToken{
		IssuedAt:   now,
		Expiration: expiration,
		NotBefore:  now,
	}

	jsonToken.Set("uid", payload.UserID)
	jsonToken.Set("tid", payload.TenantID)
	jsonToken.Set("em", payload.Email)
	rolesJSON, _ := json.Marshal(payload.Roles)
	jsonToken.Set("roles", string(rolesJSON))

	return tm.paseto.Encrypt(tm.symmetricKey, jsonToken, nil)
}

func (tm *TokenManager) ValidateToken(token string) (*TokenPayload, error) {
	var jsonToken paseto.JSONToken
	if err := tm.paseto.Decrypt(token, tm.symmetricKey, &jsonToken, nil); err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	if err := jsonToken.Validate(); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	payload := &TokenPayload{
		UserID:   jsonToken.Get("uid"),
		TenantID: jsonToken.Get("tid"),
		Email:    jsonToken.Get("em"),
		Exp:      jsonToken.Expiration.Unix(),
	}
	if err := json.Unmarshal([]byte(jsonToken.Get("roles")), &payload.Roles); err != nil {
		payload.Roles = []string{}
	}

	if payload.UserID == "" {
		return nil, fmt.Errorf("invalid token payload")
	}

	return payload, nil
}

func (tm *TokenManager) ValidateRefreshToken(token string) (*TokenPayload, error) {
	payload, err := tm.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if payload.TokenType != TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token type: expected refresh")
	}

	return payload, nil
}

func (tm *TokenManager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	payload, err := tm.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	payload.TokenType = TokenTypeAccess
	return tm.GenerateTokenPair(*payload, 15*time.Minute, 7*24*time.Hour)
}
