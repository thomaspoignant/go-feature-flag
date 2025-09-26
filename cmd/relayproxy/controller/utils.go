package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
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

	if u.EvaluationContext != nil {
		return assertContextKey(u.EvaluationContext.Key)
	}
	return assertContextKey(u.User.Key) // nolint: staticcheck
}

// assertContextKey is checking that the user key is valid, if not an echo.HTTPError is return.
// Note: Empty keys are now allowed - the core evaluation logic will determine if a targeting key
// is required based on whether the flag needs bucketing (percentage-based rules, progressive rollouts).
func assertContextKey(key string) *echo.HTTPError {
	// No validation needed - let core evaluation logic handle targeting key requirements
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
	u.Custom["anonymous"] = u.Anonymous
	return utils.ConvertEvaluationCtxFromRequest(u.Key, u.Custom), nil
}
