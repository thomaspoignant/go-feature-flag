package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

// assertRequest is the function which validates the request, if not valid an echo.HTTPError is return.
func assertRequest(u *model.AllFlagRequest) *echo.HTTPError {
	// nolint: staticcheck
	if u == nil || (u.User == nil && u.EvaluationContext == nil) {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"assertRequest: impossible to find user in request",
		)
	}
	return nil
}

func evaluationContextFromRequest(req *model.AllFlagRequest) (ffcontext.Context, error) {
	if req == nil {
		return ffcontext.EvaluationContext{},
			echo.NewHTTPError(
				http.StatusBadRequest,
				"evaluationContextFromRequest: impossible to convert the request, req nil",
			)
	}
	if req.EvaluationContext != nil {
		u := req.EvaluationContext
		return utils.ConvertEvaluationCtxFromRequest(u.Key, u.Custom), nil
	}
	return userRequestToUser(req.User) // nolint: staticcheck
}

// userRequestToUser convert a user from the request model.AllFlagRequest to a go-feature-flag ffuser.User
// nolint: staticcheck
func userRequestToUser(u *model.UserRequest) (ffcontext.Context, error) {
	if u == nil {
		return ffcontext.EvaluationContext{}, fmt.Errorf(
			"userRequestToUser: impossible to convert user, userRequest nil",
		)
	}
	custom := u.Custom
	if custom == nil {
		custom = make(map[string]any)
	}

	custom["anonymous"] = u.Anonymous
	return utils.ConvertEvaluationCtxFromRequest(u.Key, custom), nil
}
