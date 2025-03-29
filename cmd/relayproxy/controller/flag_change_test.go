package controller_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"golang.org/x/net/context"
)

func TestPIFlagChange_WithConfigChange(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()

	content, err := os.ReadFile("../testdata/controller/config_flags.yaml")
	assert.NoError(t, err)

	errWF := os.WriteFile(file.Name(), content, 0644)
	assert.NoError(t, errWF)
	file.Close()

	goFF, _ := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: file.Name(),
		},
	})
	defer goFF.Close()
	ctrl := controller.NewAPIFlagChange(goFF, metric.Metrics{})

	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.GET, "/v1/flag/change", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	c.SetPath("/v1/flag/change")
	handlerErr := ctrl.Handler(c)
	assert.NoError(t, handlerErr)

	want, _ := os.ReadFile("../testdata/controller/flag_change/flag_change_with_config_change.json")
	assert.JSONEq(t, string(want), rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)

	content, err = os.ReadFile("../testdata/controller/config_flags_v2.yaml")
	assert.NoError(t, err)

	errWF = os.WriteFile(file.Name(), content, 0644)
	assert.NoError(t, errWF)

	time.Sleep(1500 * time.Millisecond)

	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req, rec2)
	c2.SetPath("/v1/flag/change")
	handlerErr2 := ctrl.Handler(c2)
	assert.NoError(t, handlerErr2)
	assert.NotEqual(t, want, rec2.Body.String())
}

func TestPIFlagChange_WithoutConfigChange(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()

	content, err := os.ReadFile("../testdata/controller/config_flags.yaml")
	assert.NoError(t, err)

	errWF := os.WriteFile(file.Name(), content, 0644)
	assert.NoError(t, errWF)
	file.Close()

	goFF, _ := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: file.Name(),
		},
	})
	defer goFF.Close()
	ctrl := controller.NewAPIFlagChange(goFF, metric.Metrics{})

	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.GET, "/v1/flag/change", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	c.SetPath("/v1/flag/change")
	handlerErr := ctrl.Handler(c)
	assert.NoError(t, handlerErr)

	want, _ := os.ReadFile(
		"../testdata/controller/flag_change/flag_change_without_config_change.json",
	)
	assert.JSONEq(t, string(want), rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(1500 * time.Millisecond)

	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req, rec2)
	c2.SetPath("/v1/flag/change")
	handlerErr2 := ctrl.Handler(c2)
	assert.NoError(t, handlerErr2)
	assert.JSONEq(t, string(want), rec2.Body.String())
}
