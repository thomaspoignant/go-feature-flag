//go:build docker

package metric_test

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	promversion "github.com/prometheus/common/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
)

func TestMetrics_BuildInfo(t *testing.T) {
	version := "9.9.9"
	promversion.Version = version

	metricSrv, err := metric.NewMetrics()
	assert.NoError(t, err)

	expected := fmt.Sprintf(`
		# HELP gofeatureflag_build_info A metric with a constant '1' value labeled by version, revision, branch, goversion from which gofeatureflag was built, and the goos and goarch for the build.
		# TYPE gofeatureflag_build_info gauge
		gofeatureflag_build_info{branch="",goarch="%s",goos="%s",goversion="%s",revision="unknown",tags="docker",version="%s"} 1
	`, runtime.GOARCH, runtime.GOOS, runtime.Version(), version)

	require.NoError(t, testutil.GatherAndCompare(metricSrv.Registry, bytes.NewReader([]byte(expected)), "gofeatureflag_build_info"))
}
