package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

type OnboardingHandler struct{ svc *service.OnboardingService }

func NewOnboardingHandler(svc *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{svc: svc}
}

// CreateTeam godoc
// @Summary Create a team during onboarding (user with no memberships)
// @Tags    onboarding
// @Accept  json
// @Param   body body model.CreateTeamRequest true "team"
// @Success 201 {object} model.APIResponse{data=model.Team}
// @Failure 403 {object} model.APIResponse "user already belongs to a team"
// @Router  /onboarding/team [post]
func (h *OnboardingHandler) CreateTeam(c echo.Context) error {
	var req model.CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid body")
	}
	claims := middleware.MustClaims(c)
	t, err := h.svc.CreateTeamForNewUser(c.Request().Context(), claims.UserID, &req)
	if err != nil {
		return writeServiceError(c, err, "failed to create team")
	}
	return created(c, t)
}
