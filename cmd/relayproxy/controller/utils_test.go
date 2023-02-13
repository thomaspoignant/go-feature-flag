package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"net/http"
	"testing"
)

func Test_assertRequest_nil_request(t *testing.T) {
	got := assertRequest(nil)
	want := echo.NewHTTPError(http.StatusBadRequest, "assertRequest: impossible to find user in request")
	assert.Equal(t, got, want)
}

func Test_assertRequest_request_without_user(t *testing.T) {
	got := assertRequest(&model.AllFlagRequest{User: nil})
	want := echo.NewHTTPError(http.StatusBadRequest, "assertRequest: impossible to find user in request")
	assert.Equal(t, want, got)
}

func Test_userRequestToUser_nil_user(t *testing.T) {
	_, got := userRequestToUser(nil)
	want := fmt.Errorf("userRequestToUser: impossible to convert user, userRequest nil")
	assert.Equal(t, want, got)
}
