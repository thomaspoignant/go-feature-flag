package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

type TeamHandler struct{ svc *service.TeamService }

func NewTeamHandler(svc *service.TeamService) *TeamHandler { return &TeamHandler{svc: svc} }

// List godoc
// @Summary List teams visible to the user
// @Tags    teams
// @Success 200 {object} model.APIResponse{data=[]model.Team}
// @Router  /teams [get]
func (h *TeamHandler) List(c echo.Context) error {
	claims := middleware.MustClaims(c)
	items, err := h.svc.List(c.Request().Context(), claims.UserID, claims.IsSuperAdmin)
	if err != nil {
		return writeServiceError(c, err, "failed to list teams")
	}
	return ok(c, items)
}

// Create godoc
// @Summary Create a team (super admin only)
// @Tags    teams
// @Accept  json
// @Param   body body model.CreateTeamRequest true "team"
// @Success 201 {object} model.APIResponse{data=model.Team}
// @Router  /teams [post]
func (h *TeamHandler) Create(c echo.Context) error {
	var req model.CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	t, err := h.svc.Create(c.Request().Context(), claims.UserID, &req)
	if err != nil {
		return writeServiceError(c, err, "failed to create team")
	}
	return created(c, t)
}

// Get godoc
// @Summary Get team
// @Tags    teams
// @Param   id path string true "team id"
// @Success 200 {object} model.APIResponse{data=model.Team}
// @Router  /teams/{id} [get]
func (h *TeamHandler) Get(c echo.Context) error {
	t, err := h.svc.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeServiceError(c, err, "failed to get team")
	}
	if t == nil {
		return notFound(c, "team not found")
	}
	return ok(c, t)
}

// Update godoc
// @Summary Update a team
// @Tags    teams
// @Accept  json
// @Produce json
// @Param   id   path string                  true  "team id"
// @Param   body body model.UpdateTeamRequest true  "fields to update"
// @Success 200  {object} model.APIResponse{data=model.Team}
// @Failure 400  {object} model.APIResponse
// @Failure 404  {object} model.APIResponse
// @Router  /teams/{id} [patch]
func (h *TeamHandler) Update(c echo.Context) error {
	var req model.UpdateTeamRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	t, err := h.svc.Update(c.Request().Context(), claims.UserID, c.Param("id"), &req)
	if err != nil {
		return writeServiceError(c, err, "failed to update team")
	}
	return ok(c, t)
}

// Delete godoc
// @Summary Delete a team (super admin only)
// @Tags    teams
// @Param   id path string true "team id"
// @Success 200 {object} model.APIResponse
// @Failure 404 {object} model.APIResponse
// @Router  /teams/{id} [delete]
func (h *TeamHandler) Delete(c echo.Context) error {
	claims := middleware.MustClaims(c)
	if err := h.svc.Delete(c.Request().Context(), claims.UserID, c.Param("id")); err != nil {
		return writeServiceError(c, err, "failed to delete team")
	}
	return ok(c, nil)
}

// ListMembers godoc
// @Summary List team members
// @Tags    teams
// @Produce json
// @Param   id path string true "team id"
// @Success 200 {object} model.APIResponse{data=[]model.TeamMember}
// @Router  /teams/{id}/members [get]
func (h *TeamHandler) ListMembers(c echo.Context) error {
	items, err := h.svc.ListMembers(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeServiceError(c, err, "failed to list members")
	}
	return ok(c, items)
}

// AddMember godoc
// @Summary Add a member to a team
// @Tags    teams
// @Accept  json
// @Produce json
// @Param   id   path string                 true "team id"
// @Param   body body model.AddMemberRequest true "member to add"
// @Success 201  {object} model.APIResponse{data=model.TeamMember}
// @Failure 400  {object} model.APIResponse
// @Failure 404  {object} model.APIResponse
// @Router  /teams/{id}/members [post]
func (h *TeamHandler) AddMember(c echo.Context) error {
	var req model.AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	m, err := h.svc.AddMember(c.Request().Context(), claims.UserID, c.Param("id"), &req)
	if err != nil {
		return writeServiceError(c, err, "failed to add member")
	}
	return created(c, m)
}

// UpdateMember godoc
// @Summary Update a team member's role
// @Tags    teams
// @Accept  json
// @Produce json
// @Param   id   path string                    true "team id"
// @Param   uid  path string                    true "user id"
// @Param   body body model.UpdateMemberRequest true "new role"
// @Success 200  {object} model.APIResponse
// @Failure 400  {object} model.APIResponse
// @Failure 404  {object} model.APIResponse
// @Router  /teams/{id}/members/{uid} [patch]
func (h *TeamHandler) UpdateMember(c echo.Context) error {
	var req model.UpdateMemberRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	if err := h.svc.UpdateMember(c.Request().Context(), claims.UserID, c.Param("id"), c.Param("uid"), req.Role); err != nil {
		return writeServiceError(c, err, "failed to update member")
	}
	return ok(c, nil)
}

// RemoveMember godoc
// @Summary Remove a member from a team
// @Tags    teams
// @Param   id  path string true "team id"
// @Param   uid path string true "user id"
// @Success 200 {object} model.APIResponse
// @Failure 404 {object} model.APIResponse
// @Router  /teams/{id}/members/{uid} [delete]
func (h *TeamHandler) RemoveMember(c echo.Context) error {
	claims := middleware.MustClaims(c)
	if err := h.svc.RemoveMember(c.Request().Context(), claims.UserID, c.Param("id"), c.Param("uid")); err != nil {
		return writeServiceError(c, err, "failed to remove member")
	}
	return ok(c, nil)
}
