package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

// assertRequest is the function which validates the request, if not valid an echo.HTTPError is return.
func assertRequest(u *model.RelayProxyRequest) *echo.HTTPError {
	if u == nil || u.User == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "assertRequest: impossible to find user in request")
	}
	return assertUserKey(u.User.Key)
}

// assertUserKey is checking that the user key is valid, if not an echo.HTTPError is return.
func assertUserKey(userKey string) *echo.HTTPError {
	if len(userKey) == 0 {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "empty key for user, impossible to retrieve flags"}
	}
	return nil
}

// userRequestToUser convert a user from the request model.RelayProxyRequest to a go-feature-flag ffuser.User
func userRequestToUser(u *model.UserRequest) (ffuser.User, error) {
	if u == nil {
		return ffuser.User{}, fmt.Errorf("userRequestToUser: impossible to convert user, userRequest nil")
	}
	uBuilder := ffuser.NewUserBuilder(u.Key).Anonymous(u.Anonymous)
	for key, val := range u.Custom {
		uBuilder.AddCustom(key, val)
	}
	user := uBuilder.Build()
	return user, nil
}
