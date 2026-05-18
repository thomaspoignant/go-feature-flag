package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

type VersionHandler struct{ svc *service.VersionService }

func NewVersionHandler(svc *service.VersionService) *VersionHandler { return &VersionHandler{svc: svc} }

// List godoc
// @Summary  List versions of a flag
// @Tags     versions
// @Produce  json
// @Param    id       path  string true  "flag id"
// @Param    page     query int    false "page" default(1)
// @Param    pageSize query int    false "items per page" default(50)
// @Success  200      {object} model.APIResponse{data=model.PaginatedResponse[model.FlagVersion]}
// @Router   /flags/{id}/versions [get]
func (h *VersionHandler) List(c echo.Context) error {
	page := parseInt(c.QueryParam("page"), 1)
	pageSize := parseInt(c.QueryParam("pageSize"), 50)
	items, total, err := h.svc.List(c.Request().Context(), c.Param("id"), page, pageSize)
	if err != nil {
		return writeServiceError(c, err, "failed to list versions")
	}
	return ok(c, model.NewPaginatedResponse(items, total, page, pageSize))
}

// Get godoc
// @Summary  Get a specific flag version
// @Tags     versions
// @Produce  json
// @Param    id path string true "flag id"
// @Param    n  path int    true "version number"
// @Success  200 {object} model.APIResponse{data=model.FlagVersion}
// @Failure  400 {object} model.APIResponse
// @Failure  404 {object} model.APIResponse
// @Router   /flags/{id}/versions/{n} [get]
func (h *VersionHandler) Get(c echo.Context) error {
	num, err := strconv.Atoi(c.Param("n"))
	if err != nil {
		return badRequest(c, "invalid version number")
	}
	v, err := h.svc.Get(c.Request().Context(), c.Param("id"), num)
	if err != nil {
		return writeServiceError(c, err, "failed to get version")
	}
	if v == nil {
		return notFound(c, "version not found")
	}
	return ok(c, v)
}

// Rollback godoc
// @Summary  Roll back to a previous version (creates a new current version)
// @Tags     versions
// @Accept   json
// @Produce  json
// @Param    id   path string                 true  "flag id"
// @Param    n    path int                    true  "version number to roll back to"
// @Param    body body model.RollbackRequest  false "optional comment"
// @Success  201  {object} model.APIResponse{data=model.FlagVersion}
// @Failure  404  {object} model.APIResponse
// @Router   /flags/{id}/versions/{n}/rollback [post]
func (h *VersionHandler) Rollback(c echo.Context) error {
	num, err := strconv.Atoi(c.Param("n"))
	if err != nil {
		return badRequest(c, "invalid version number")
	}
	var req model.RollbackRequest
	_ = c.Bind(&req)
	claims := middleware.MustClaims(c)
	v, err := h.svc.Rollback(c.Request().Context(), claims.UserID, c.Param("id"), num, req.Comment)
	if err != nil {
		return writeServiceError(c, err, "failed to rollback")
	}
	return created(c, v)
}
