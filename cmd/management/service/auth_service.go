package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/repository"
)

type AuthService struct {
	cfg         config.Config
	users       *repository.UserRepo
	provider    *oidc.Provider
	oauth       *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	adminEmails map[string]struct{}
}

func NewAuthService(ctx context.Context, cfg config.Config, users *repository.UserRepo) (*AuthService, error) {
	provider, err := oidc.NewProvider(ctx, cfg.OIDC.Issuer)
	if err != nil {
		return nil, fmt.Errorf("init oidc provider: %w", err)
	}
	o := &oauth2.Config{
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.OIDC.RedirectURL,
		Scopes:       cfg.OIDC.Scopes,
	}
	v := provider.Verifier(&oidc.Config{ClientID: cfg.OIDC.ClientID})
	adm := make(map[string]struct{}, len(cfg.Auth.AdminEmails))
	for _, e := range cfg.Auth.AdminEmails {
		adm[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	return &AuthService{cfg: cfg, users: users, provider: provider, oauth: o, verifier: v, adminEmails: adm}, nil
}

func (s *AuthService) AuthURL(state string) string {
	return s.oauth.AuthCodeURL(state)
}

func (s *AuthService) RandomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type Claims struct {
	UserID       string `json:"uid"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	IsSuperAdmin bool   `json:"sa"`
	jwt.RegisteredClaims
}

func (s *AuthService) Exchange(ctx context.Context, code string) (*model.User, string, error) {
	tok, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("oauth exchange: %w", err)
	}
	rawID, ok := tok.Extra("id_token").(string)
	if !ok || rawID == "" {
		return nil, "", errors.New("id_token missing from token response")
	}
	idt, err := s.verifier.Verify(ctx, rawID)
	if err != nil {
		return nil, "", fmt.Errorf("verify id_token: %w", err)
	}
	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}
	if err := idt.Claims(&claims); err != nil {
		return nil, "", fmt.Errorf("decode claims: %w", err)
	}
	if claims.Email == "" {
		return nil, "", errors.New("email claim missing")
	}

	_, isAdmin := s.adminEmails[strings.ToLower(claims.Email)]

	u, err := s.users.UpsertFromOIDC(ctx, nil, claims.Email, claims.Name, claims.Sub, isAdmin)
	if err != nil {
		return nil, "", err
	}

	signed, err := s.IssueJWT(u)
	if err != nil {
		return nil, "", err
	}
	return u, signed, nil
}

func (s *AuthService) IssueJWT(u *model.User) (string, error) {
	now := time.Now()
	c := Claims{
		UserID:       u.ID,
		Email:        u.Email,
		Name:         u.Name,
		IsSuperAdmin: u.IsSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.Auth.SessionMaxAge)),
			Subject:   u.ID,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return tok.SignedString([]byte(s.cfg.Auth.JWTSecret))
}

func (s *AuthService) ParseJWT(token string) (*Claims, error) {
	c := &Claims{}
	_, err := jwt.ParseWithClaims(token, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.Auth.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
