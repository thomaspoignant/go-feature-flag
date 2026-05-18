package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

type FlagHandler struct{ svc *service.FlagService }

func NewFlagHandler(svc *service.FlagService) *FlagHandler { return &FlagHandler{svc: svc} }

func parseBoolPtr(s string) *bool {
	if s == "" {
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil
	}
	return &b
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// List godoc
// @Summary  List flags of a flagset
// @Tags     flags
// @Produce  json
// @Param    flagsetId path  string true  "flagset id"
// @Param    name      query string false "name contains filter"
// @Param    disabled  query bool   false "filter disabled flags"
// @Param    page      query int    false "page number"     default(1)
// @Param    pageSize  query int    false "items per page"  default(50)
// @Success  200       {object} model.APIResponse{data=model.PaginatedResponse[model.FlagListItem]}
// @Router   /flagsets/{flagsetId}/flags [get]
func (h *FlagHandler) List(c echo.Context) error {
	flagsetID := c.Param("flagsetId")
	var namePtr *string
	if n := c.QueryParam("name"); n != "" {
		namePtr = &n
	}
	filters := model.FlagFilters{
		Name:     namePtr,
		Disabled: parseBoolPtr(c.QueryParam("disabled")),
		Page:     parseInt(c.QueryParam("page"), 1),
		PageSize: parseInt(c.QueryParam("pageSize"), 50),
	}
	items, total, err := h.svc.List(c.Request().Context(), flagsetID, filters)
	if err != nil {
		return writeServiceError(c, err, "failed to list flags")
	}
	return ok(c, model.NewPaginatedResponse(items, total, filters.Page, filters.PageSize))
}

// Create godoc
// @Summary  Create a flag in a flagset (creates version 1)
// @Tags     flags
// @Accept   json
// @Produce  json
// @Param    flagsetId path string                  true "flagset id"
// @Param    body      body model.CreateFlagRequest true "flag payload (GOFF DTO)"
// @Success  201       {object} model.APIResponse{data=model.Flag}
// @Failure  400       {object} model.APIResponse
// @Failure  404       {object} model.APIResponse
// @Router   /flagsets/{flagsetId}/flags [post]
func (h *FlagHandler) Create(c echo.Context) error {
	var req model.CreateFlagRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	f, err := h.svc.Create(c.Request().Context(), claims.UserID, c.Param("flagsetId"), &req)
	if err != nil {
		return writeServiceError(c, err, "failed to create flag")
	}
	return created(c, f)
}

// Get godoc
// @Summary  Get a flag with its current version payload
// @Tags     flags
// @Produce  json
// @Param    id path string true "flag id"
// @Success  200 {object} model.APIResponse{data=model.Flag}
// @Failure  404 {object} model.APIResponse
// @Router   /flags/{id} [get]
func (h *FlagHandler) Get(c echo.Context) error {
	f, err := h.svc.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeServiceError(c, err, "failed to get flag")
	}
	if f == nil {
		return notFound(c, "flag not found")
	}
	return ok(c, f)
}

// Update godoc
// @Summary  Replace flag payload (creates a new version)
// @Tags     flags
// @Accept   json
// @Produce  json
// @Param    id   path string                  true "flag id"
// @Param    body body model.UpdateFlagRequest true "new payload"
// @Success  200  {object} model.APIResponse{data=model.Flag}
// @Failure  400  {object} model.APIResponse
// @Failure  404  {object} model.APIResponse
// @Router   /flags/{id} [put]
func (h *FlagHandler) Update(c echo.Context) error {
	var req model.UpdateFlagRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	f, err := h.svc.Update(c.Request().Context(), claims.UserID, c.Param("id"), &req)
	if err != nil {
		return writeServiceError(c, err, "failed to update flag")
	}
	return ok(c, f)
}

// Disable godoc
// @Summary  Toggle the disabled state of a flag (no new version)
// @Tags     flags
// @Accept   json
// @Produce  json
// @Param    id   path string                   true "flag id"
// @Param    body body model.DisableFlagRequest true "disabled state"
// @Success  200  {object} model.APIResponse
// @Failure  400  {object} model.APIResponse
// @Failure  404  {object} model.APIResponse
// @Router   /flags/{id}/disable [post]
func (h *FlagHandler) Disable(c echo.Context) error {
	var req model.DisableFlagRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	if err := h.svc.SetDisabled(c.Request().Context(), claims.UserID, c.Param("id"), req.Disabled); err != nil {
		return writeServiceError(c, err, "failed to set disabled")
	}
	return ok(c, nil)
}

// Delete godoc
// @Summary  Soft-delete a flag (audited; versions remain)
// @Tags     flags
// @Param    id path string true "flag id"
// @Success  200 {object} model.APIResponse
// @Failure  404 {object} model.APIResponse
// @Router   /flags/{id} [delete]
func (h *FlagHandler) Delete(c echo.Context) error {
	claims := middleware.MustClaims(c)
	if err := h.svc.Delete(c.Request().Context(), claims.UserID, c.Param("id")); err != nil {
		return writeServiceError(c, err, "failed to delete flag")
	}
	return ok(c, nil)
}
