package crypto

import (
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
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt   int64  `json:"expires_at"`
	TokenType   string `json:"token_type"`
}

type TokenManager struct {
	paseto       *paseto.PASETO
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
		ExpiresAt:   time.Now().Add(accessTTL).Unix(),
		TokenType:   "Bearer",
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
		Expiration:  expiration,
		NotBefore:   now,
	}

	for _, role := range payload.Roles {
		jsonToken.Set(role, true)
	}

	return tm.paseto.Encrypt(tm.symmetricKey, jsonToken, payload, nil)
}

func (tm *TokenManager) ValidateToken(token string) (*TokenPayload, error) {
	var payload TokenPayload
	if err := tm.paseto.Decrypt(token, tm.symmetricKey, &payload, nil); err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	if payload.TokenType == TokenTypeAccess && time.Now().After(time.Unix(payload.Exp, 0)) {
		return nil, fmt.Errorf("token expired")
	}

	return &payload, nil
}

func (tm *TokenManager) ValidateRefreshToken(token string) (*TokenPayload, error) {
	var payload TokenPayload
	if err := tm.paseto.Decrypt(token, tm.symmetricKey, &payload, nil); err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	if payload.TokenType != TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token type: expected refresh")
	}

	if time.Now().After(time.Unix(payload.Exp, 0)) {
		return nil, fmt.Errorf("token expired")
	}

	return &payload, nil
}

func (tm *TokenManager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	payload, err := tm.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	payload.TokenType = TokenTypeAccess
	return tm.GenerateTokenPair(*payload, 15*time.Minute, 7*24*time.Hour)
}
