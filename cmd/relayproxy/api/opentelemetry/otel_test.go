package opentelemetry

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitSampler(t *testing.T) {
	t.Run("OTEL_TRACES_SAMPLER unset", func(t *testing.T) {
		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		sampler, err := initSampler("test", c)
		require.NoError(t, err)
		assert.Equal(t, sdktrace.AlwaysSample(), sampler)
	})

	t.Run("OTEL_TRACES_SAMPLER set to non-jaeger_remote", func(t *testing.T) {
		t.Setenv("OTEL_TRACES_SAMPLER", "always_on")
		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		sampler, err := initSampler("test", c)
		require.NoError(t, err)
		assert.Nil(t, sampler)
	})

	t.Run("OTEL_TRACES_SAMPLER set to jaeger_remote", func(t *testing.T) {
		t.Setenv("OTEL_TRACES_SAMPLER", "jaeger_remote")
		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		sampler, err := initSampler("test", c)
		require.NoError(t, err)

		// not really any way to assert on the sampler other than calling
		// Description()...
		assert.Equal(t, "JaegerRemoteSampler{}", sampler.Description())
	})
}

func TestJaegerRemoteSamplerOpts(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		url, refreshInterval, maxOperations, err := jaegerRemoteSamplerOpts(c)
		require.NoError(t, err)
		assert.Equal(t, defaultSamplerURL, url)
		assert.Equal(t, defaultSamplingRefreshInterval, refreshInterval)
		assert.Equal(t, defaultSamplingMaxOperations, maxOperations)
	})

	t.Run("JAEGER_SAMPLER_MANAGER_HOST_PORT set", func(t *testing.T) {
		expected := "http://example.com:1234"
		t.Setenv("JAEGER_SAMPLER_MANAGER_HOST_PORT", expected)

		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		url, _, _, err := jaegerRemoteSamplerOpts(c)
		require.NoError(t, err)
		assert.Equal(t, expected, url)
	})

	t.Run("JAEGER_SAMPLER_REFRESH_INTERVAL set", func(t *testing.T) {
		expected := 42 * time.Second
		t.Setenv("JAEGER_SAMPLER_REFRESH_INTERVAL", expected.String())

		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		_, refreshInterval, _, err := jaegerRemoteSamplerOpts(c)
		require.NoError(t, err)
		assert.Equal(t, expected, refreshInterval)
	})

	t.Run("JAEGER_SAMPLER_MAX_OPERATIONS set", func(t *testing.T) {
		expected := 42
		t.Setenv("JAEGER_SAMPLER_MAX_OPERATIONS", strconv.Itoa(expected))

		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		_, _, maxOperations, err := jaegerRemoteSamplerOpts(c)
		require.NoError(t, err)
		assert.Equal(t, expected, maxOperations)
	})

	t.Run("invalid JAEGER_SAMPLER_REFRESH_INTERVAL", func(t *testing.T) {
		t.Setenv("JAEGER_SAMPLER_REFRESH_INTERVAL", "bogus")

		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)

		_, _, _, err := jaegerRemoteSamplerOpts(c)
		require.Error(t, err)
	})

	t.Run("invalid JAEGER_SAMPLER_MAX_OPERATIONS", func(t *testing.T) {
		t.Setenv("JAEGER_SAMPLER_MAX_OPERATIONS", "bogus")

		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		_, errC := config.New(f, zap.L(), "1.X.X")
		require.Error(t, errC)
	})
}

func TestInitResource(t *testing.T) {
	t.Run("defaults, no env", func(t *testing.T) {
		res, err := initResource(context.Background(), "test", "1.2.3", nil)
		require.NoError(t, err)

		rmap := map[string]attribute.Value{}
		for _, attr := range res.Attributes() {
			rmap[string(attr.Key)] = attr.Value
		}

		// just spot-check a few things
		assert.Equal(t, "test", rmap["service.name"].AsString())
		assert.Equal(t, "1.2.3", rmap["service.version"].AsString())
		assert.Equal(t, "go", rmap["process.runtime.name"].AsString())
	})

	t.Run("with config map set", func(t *testing.T) {
		attribs := map[string]string{
			"key1": "val1",
			"key2": "val2",
		}

		res, err := initResource(context.Background(), "test", "1.2.3", attribs)
		require.NoError(t, err)

		rmap := map[string]attribute.Value{}
		for _, attr := range res.Attributes() {
			rmap[string(attr.Key)] = attr.Value
		}

		assert.Equal(t, "val1", rmap["key1"].AsString())
		assert.Equal(t, "val2", rmap["key2"].AsString())
	})
}

func TestInit(t *testing.T) {
	logger := log.InitLogger().ZapLogger

	svc := NewOtelService()

	t.Run("no config", func(t *testing.T) {
		err := svc.Init(context.Background(), logger, &config.Config{})
		require.NoError(t, err)
		defer func() { _ = svc.Stop(context.Background()) }()
		assert.NotNil(t, otel.GetTracerProvider())
	})

	t.Run("disabled", func(t *testing.T) {
		t.Setenv("OTEL_SDK_DISABLED", "true")
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://example.com:4318")
		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)
		err := svc.Init(context.Background(), logger, c)
		require.NoError(t, err)
		defer func() { _ = svc.Stop(context.Background()) }()
		assert.Equal(t, noop.NewTracerProvider(), otel.GetTracerProvider())
	})

	t.Run("support openTelemetryOtlpEndpoint", func(t *testing.T) {
		err := svc.Init(context.Background(), logger, &config.Config{
			OpenTelemetryOtlpEndpoint: "https://example.com:4318",
		})
		require.NoError(t, err)
		defer func() { _ = svc.Stop(context.Background()) }()
		assert.Equal(t, "https://example.com:4318", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	})

	t.Run("OTEL_EXPORTER_OTLP_ENDPOINT takes precedence", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://example.com:4318")
		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)
		c.OpenTelemetryOtlpEndpoint = "https://bogus.com:4317"
		err := svc.Init(context.Background(), logger, c)
		require.NoError(t, err)
		defer func() { _ = svc.Stop(context.Background()) }()
		assert.Equal(t, "https://example.com:4318", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	})

	t.Run("error handler logs to zap", func(t *testing.T) {
		obs, logs := observer.New(zap.InfoLevel)
		testLogger := zap.New(obs)

		expectedErr := errors.New("test error")

		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://example.com:4318")
		f := pflag.NewFlagSet("config", pflag.ContinueOnError)
		c, errC := config.New(f, zap.L(), "1.X.X")
		require.NoError(t, errC)
		c.OpenTelemetryOtlpEndpoint = "https://bogus.com:4317"
		err := svc.Init(context.Background(), testLogger, c)
		require.NoError(t, err)
		defer func() { _ = svc.Stop(context.Background()) }()
		otel.GetErrorHandler().Handle(expectedErr)

		require.Len(t, logs.All(), 1)

		want := []observer.LoggedEntry{{
			Entry:   zapcore.Entry{Level: zap.ErrorLevel, Message: "OTel error"},
			Context: []zapcore.Field{zap.Error(expectedErr)},
		}}

		assert.Equal(t, want, logs.AllUntimed())
	})
}
