package controller

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/internal/apikey"
)

type apiKeyCreate struct {
	s apikey.Storage
}

func NewAPIKeyCreate(s apikey.Storage) Controller {
	return &apiKeyCreate{
		s: s,
	}
}

// Handler is the entry point for the API key create endpoint
// @Summary      Create API key
// @Success      201  {object}   modeldocs.EvalFlagDoc "Success"
// @Failure 		 400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
func (h *apiKeyCreate) Handler(c echo.Context) error {
	// validateAuthHeader
	// validate user param
	// store and generate sha256 api key
	// return generated api key
	return echo.ErrNotImplemented
}

type apiKeyGet struct {
	s apikey.Storage
}

func NewAPIKeyGet(s apikey.Storage) Controller {
	return &apiKeyGet{
		s: s,
	}
}

// Handler is the entry point for the API key get endpoint
// @Summary      Get all API keys
// @Success      200 {object} modeldocs.EvalFlagDoc "Success"
// @Failure 		 400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
func (h *apiKeyGet) Handler(c echo.Context) error {
	// validateAuthHeader
	// return all authorized keys
	return echo.ErrNotImplemented
}

type apiKeyDelete struct {
	s apikey.Storage
}

func NewAPIKeyDelete(s apikey.Storage) Controller {
	// validateAuthHeader
	// validate whether key param exist in storage
	// delete key
	return &apiKeyDelete{
		s: s,
	}
}

// Handler is the entry point for the API key delete endpoint
// @Summary      Delete API key
// @Success      204 {object} modeldocs.EvalFlagDoc "Success"
// @Failure 		 400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
func (h *apiKeyDelete) Handler(c echo.Context) error {
	return echo.ErrNotImplemented
}

func validateAuthHeader(c echo.Context, s apikey.Storage) error {
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return errors.New("header should have 2 parts, bearer and key")
	}
	k := parts[1]
	u, ok := s.Read(k)
	if !ok {
		return errors.New("api key is not exist")
	}
	if isAdmin, ok := u.GetCustom()["adminapikey"].(bool); !ok || !isAdmin {
		return errors.New("user is not admin")
	}
	return nil
}
