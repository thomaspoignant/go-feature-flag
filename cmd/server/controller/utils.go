package controller

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/server/model"
)

// assertRequest is the function which validates the request, if not valid an echo.HTTPError is return.
func assertRequest(u *model.AllFlagRequest) *echo.HTTPError {
	// nolint: staticcheck
	if u == nil || (u.User == nil && u.EvaluationContext == nil) {
		return echo.NewHTTPError(http.StatusBadRequest, "assertRequest: impossible to find user in request")
	}

	if u.EvaluationContext != nil {
		return assertContextKey(u.EvaluationContext.Key)
	}
	return assertContextKey(u.User.Key) // nolint: staticcheck
}

// assertContextKey is checking that the user key is valid, if not an echo.HTTPError is return.
func assertContextKey(key string) *echo.HTTPError {
	if len(key) == 0 {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "empty key for evaluation context, impossible to retrieve flags",
		}
	}
	return nil
}

func evaluationContextFromRequest(req *model.AllFlagRequest) (ffcontext.Context, error) {
	if req == nil {
		return ffcontext.EvaluationContext{},
			echo.NewHTTPError(http.StatusBadRequest, "evaluationContextFromRequest: impossible to convert the request, req nil")
	}
	if req.EvaluationContext != nil {
		u := req.EvaluationContext
		ctx := ffcontext.NewEvaluationContextBuilder(u.Key)
		for key, val := range u.Custom {
			ctx.AddCustom(key, val)
		}
		return ctx.Build(), nil
	}
	return userRequestToUser(req.User) // nolint: staticcheck
}

// userRequestToUser convert a user from the request model.AllFlagRequest to a go-feature-flag ffuser.User
// nolint: staticcheck
func userRequestToUser(u *model.UserRequest) (ffuser.User, error) {
	if u == nil {
		// nolint: staticcheck
		return ffuser.User{}, fmt.Errorf("userRequestToUser: impossible to convert user, userRequest nil")
	}

	uBuilder := ffuser.NewUserBuilder(u.Key).Anonymous(u.Anonymous) // nolint: staticcheck
	for key, val := range u.Custom {
		uBuilder.AddCustom(key, val)
	}
	user := uBuilder.Build()
	return user, nil
}
