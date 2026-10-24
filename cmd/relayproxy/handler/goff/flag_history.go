package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/thomaspoignant/go-feature-flag/internal/flaghistory"
)

type flagHistory struct {
	history *flaghistory.History
}

// NewFlagHistory initialize the controller for the /admin/v1/flag-history endpoints
func NewFlagHistory(history *flaghistory.History) Controller {
	return &flagHistory{history: history}
}

// Handler returns the modifications made on the flags, most recent first.
// @Summary      List the modifications made on the flags.
// @Tags Admin API to manage GO Feature Flag
// @Description  Every modification (added, updated, deleted) of a flag is recorded with its author and
// @Description  the list of the fields that changed. The result is paginated with a cursor.
// @Security     ApiKeyAuth
// @Security     XApiKeyAuth
// @Produce      json
// @Param        flagKey path  string false "Only the modifications of this flag (only on /admin/v1/flag-history/{flagKey})"
// @Param        flagset query string false "Only the modifications of this flagset"
// @Param        action  query string false "Only this kind of modification" Enums(added, updated, deleted)
// @Param        author  query string false "Only the modifications made by this author"
// @Param        from    query string false "Only the modifications after this date (RFC 3339)"
// @Param        to      query string false "Only the modifications before this date (RFC 3339)"
// @Param        limit   query int    false "Page size (default 50, max 500)"
// @Param        cursor  query int    false "nextCursor returned by the previous page"
// @Success      200  {object} flaghistory.Page "Success"
// @Failure      400  {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      503  {object} modeldocs.HTTPErrorDoc "The history database is not reachable"
// @Router       /admin/v1/flag-history [get]
// @Router       /admin/v1/flag-history/{flagKey} [get]
func (h *flagHistory) Handler(c *echo.Context) error {
	filter, err := parseHistoryFilter(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	store, err := h.history.Store()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, err.Error())
	}
	page, err := store.List(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "impossible to read the flag history")
	}
	return c.JSON(http.StatusOK, page)
}

func parseHistoryFilter(c *echo.Context) (flaghistory.Filter, error) {
	q := c.QueryParams()
	f := flaghistory.Filter{
		Flagset: q.Get("flagset"),
		FlagKey: c.Param("flagKey"),
		Author:  q.Get("author"),
	}

	switch a := flaghistory.Action(q.Get("action")); a {
	case "", flaghistory.ActionAdded, flaghistory.ActionUpdated, flaghistory.ActionDeleted:
		f.Action = a
	default:
		return f, fmt.Errorf("invalid action %q, expected added, updated or deleted", a)
	}

	var err error
	if f.From, err = parseTime(q.Get("from"), "from"); err != nil {
		return f, err
	}
	if f.To, err = parseTime(q.Get("to"), "to"); err != nil {
		return f, err
	}
	if !f.From.IsZero() && !f.To.IsZero() && f.From.After(f.To) {
		return f, errors.New("from must be before to")
	}

	if v := q.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil || limit < 1 {
			return f, fmt.Errorf("invalid limit %q, expected a positive integer", v)
		}
		f.Limit = limit
	}
	if v := q.Get("cursor"); v != "" {
		cursor, err := strconv.ParseInt(v, 10, 64)
		if err != nil || cursor < 0 {
			return f, fmt.Errorf("invalid cursor %q", v)
		}
		f.Cursor = cursor
	}
	return f, nil
}

func parseTime(value, name string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s %q, expected a RFC 3339 date", name, value)
	}
	return t, nil
}
