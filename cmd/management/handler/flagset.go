package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

type FlagsetHandler struct{ svc *service.FlagsetService }

func NewFlagsetHandler(svc *service.FlagsetService) *FlagsetHandler { return &FlagsetHandler{svc: svc} }

// ListByTeam godoc
// @Summary  List flagsets of a team
// @Tags     flagsets
// @Produce  json
// @Param    teamId path string true "team id"
// @Success  200    {object} model.APIResponse{data=[]model.Flagset}
// @Router   /teams/{teamId}/flagsets [get]
func (h *FlagsetHandler) ListByTeam(c echo.Context) error {
	items, err := h.svc.ListByTeam(c.Request().Context(), c.Param("teamId"))
	if err != nil {
		return writeServiceError(c, err, "failed to list flagsets")
	}
	return ok(c, items)
}

// Create godoc
// @Summary  Create a flagset under a team
// @Tags     flagsets
// @Accept   json
// @Produce  json
// @Param    teamId path string                     true "team id"
// @Param    body   body model.CreateFlagsetRequest true "flagset"
// @Success  201    {object} model.APIResponse{data=model.Flagset}
// @Failure  400    {object} model.APIResponse
// @Failure  404    {object} model.APIResponse
// @Router   /teams/{teamId}/flagsets [post]
func (h *FlagsetHandler) Create(c echo.Context) error {
	var req model.CreateFlagsetRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	fs, err := h.svc.Create(c.Request().Context(), claims.UserID, c.Param("teamId"), &req)
	if err != nil {
		return writeServiceError(c, err, "failed to create flagset")
	}
	return created(c, fs)
}

// Get godoc
// @Summary  Get a flagset
// @Tags     flagsets
// @Produce  json
// @Param    id path string true "flagset id"
// @Success  200 {object} model.APIResponse{data=model.Flagset}
// @Failure  404 {object} model.APIResponse
// @Router   /flagsets/{id} [get]
func (h *FlagsetHandler) Get(c echo.Context) error {
	fs, err := h.svc.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeServiceError(c, err, "failed to get flagset")
	}
	if fs == nil {
		return notFound(c, "flagset not found")
	}
	return ok(c, fs)
}

// Update godoc
// @Summary  Update a flagset
// @Tags     flagsets
// @Accept   json
// @Produce  json
// @Param    id   path string                     true "flagset id"
// @Param    body body model.UpdateFlagsetRequest true "fields to update"
// @Success  200  {object} model.APIResponse{data=model.Flagset}
// @Failure  400  {object} model.APIResponse
// @Failure  404  {object} model.APIResponse
// @Router   /flagsets/{id} [patch]
func (h *FlagsetHandler) Update(c echo.Context) error {
	var req model.UpdateFlagsetRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	fs, err := h.svc.Update(c.Request().Context(), claims.UserID, c.Param("id"), &req)
	if err != nil {
		return writeServiceError(c, err, "failed to update flagset")
	}
	return ok(c, fs)
}

// Delete godoc
// @Summary  Delete a flagset
// @Tags     flagsets
// @Param    id path string true "flagset id"
// @Success  200 {object} model.APIResponse
// @Failure  404 {object} model.APIResponse
// @Router   /flagsets/{id} [delete]
func (h *FlagsetHandler) Delete(c echo.Context) error {
	claims := middleware.MustClaims(c)
	if err := h.svc.Delete(c.Request().Context(), claims.UserID, c.Param("id")); err != nil {
		return writeServiceError(c, err, "failed to delete flagset")
	}
	return ok(c, nil)
}

// CreateAPIKey godoc
// @Summary  Create a new API key for a flagset (raw key returned ONCE)
// @Tags     flagsets
// @Produce  json
// @Param    id path string true "flagset id"
// @Success  201 {object} model.APIResponse{data=model.CreateAPIKeyResponse}
// @Failure  404 {object} model.APIResponse
// @Router   /flagsets/{id}/api-keys [post]
func (h *FlagsetHandler) CreateAPIKey(c echo.Context) error {
	claims := middleware.MustClaims(c)
	out, err := h.svc.CreateAPIKey(c.Request().Context(), claims.UserID, c.Param("id"))
	if err != nil {
		return writeServiceError(c, err, "failed to create api key")
	}
	return created(c, out)
}

// DeleteAPIKey godoc
// @Summary  Revoke a flagset API key by hash
// @Tags     flagsets
// @Param    id      path string true "flagset id"
// @Param    keyHash path string true "key hash"
// @Success  200     {object} model.APIResponse
// @Failure  404     {object} model.APIResponse
// @Router   /flagsets/{id}/api-keys/{keyHash} [delete]
func (h *FlagsetHandler) DeleteAPIKey(c echo.Context) error {
	claims := middleware.MustClaims(c)
	if err := h.svc.DeleteAPIKey(c.Request().Context(), claims.UserID, c.Param("id"), c.Param("keyHash")); err != nil {
		return writeServiceError(c, err, "failed to delete api key")
	}
	return ok(c, nil)
}
