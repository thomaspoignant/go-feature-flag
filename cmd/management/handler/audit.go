package handler

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

type AuditHandler struct{ svc *service.AuditService }

func NewAuditHandler(svc *service.AuditService) *AuditHandler { return &AuditHandler{svc: svc} }

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

// List godoc
// @Summary  List audit log entries with filters
// @Tags     audit
// @Produce  json
// @Param    teamId    query string false "filter by team id"
// @Param    flagsetId query string false "filter by flagset id"
// @Param    flagId    query string false "filter by flag id"
// @Param    actorId   query string false "filter by actor user id"
// @Param    action    query string false "filter by action (e.g. flag.create)"
// @Param    from      query string false "ISO-8601 lower bound (inclusive)"
// @Param    to        query string false "ISO-8601 upper bound (inclusive)"
// @Param    page      query int    false "page" default(1)
// @Param    pageSize  query int    false "items per page" default(50)
// @Success  200       {object} model.APIResponse{data=model.PaginatedResponse[model.AuditEntry]}
// @Router   /audit [get]
func (h *AuditHandler) List(c echo.Context) error {
	f := model.AuditFilters{
		TeamID:    strPtr(c.QueryParam("teamId")),
		FlagsetID: strPtr(c.QueryParam("flagsetId")),
		FlagID:    strPtr(c.QueryParam("flagId")),
		ActorID:   strPtr(c.QueryParam("actorId")),
		Action:    strPtr(c.QueryParam("action")),
		From:      timePtr(c.QueryParam("from")),
		To:        timePtr(c.QueryParam("to")),
		Page:      parseInt(c.QueryParam("page"), 1),
		PageSize:  parseInt(c.QueryParam("pageSize"), 50),
	}
	items, total, err := h.svc.List(c.Request().Context(), f)
	if err != nil {
		return writeServiceError(c, err, "failed to list audit log")
	}
	return ok(c, model.NewPaginatedResponse(items, total, f.Page, f.PageSize))
}
