package metric

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

type fakeFlagsetSource struct {
	flagsets map[string]*ffclient.GoFeatureFlag
	err      error
}

func (f fakeFlagsetSource) AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	return f.flagsets, f.err
}

func newFileClient(t *testing.T, content string) (*ffclient.GoFeatureFlag, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "flags.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	client, err := ffclient.New(ffclient.Config{
		PollingInterval: time.Hour,
		Retriever:       &fileretriever.Retriever{Path: path},
	})
	require.NoError(t, err)
	t.Cleanup(client.Close)
	return client, path
}

const flagsYAML = `
flag-on:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: enabled
flag-off:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: enabled
  disable: true
`

func TestFlagLoaded_ExposesOneSeriesPerFlagPerFlagset(t *testing.T) {
	teamA, _ := newFileClient(t, flagsYAML)
	teamB, _ := newFileClient(t, `
other-flag:
  variations:
    enabled: true
  defaultRule:
    variation: enabled
`)
	m, err := NewMetrics()
	require.NoError(t, err)
	require.NoError(t, m.RegisterFlagLoaded(fakeFlagsetSource{
		flagsets: map[string]*ffclient.GoFeatureFlag{"team-a": teamA, "team-b": teamB},
	}))

	expected := `
# HELP gofeatureflag_flag_loaded Flags currently loaded in the relay proxy cache, one series per flag with value 1.
# TYPE gofeatureflag_flag_loaded gauge
gofeatureflag_flag_loaded{disabled="false",flag_name="flag-on",flagset="team-a"} 1
gofeatureflag_flag_loaded{disabled="true",flag_name="flag-off",flagset="team-a"} 1
gofeatureflag_flag_loaded{disabled="false",flag_name="other-flag",flagset="team-b"} 1
`
	assert.NoError(t, testutil.GatherAndCompare(m.Registry, strings.NewReader(expected), "gofeatureflag_flag_loaded"))
}

func TestFlagLoaded_DisabledReflectsScheduledStepsAndExperimentation(t *testing.T) {
	client, _ := newFileClient(t, `
scheduled-off:
  variations:
    enabled: true
  defaultRule:
    variation: enabled
  scheduledRollout:
    - date: 2020-01-01T00:00:00Z
      disable: true
scheduled-back-on:
  variations:
    enabled: true
  defaultRule:
    variation: enabled
  disable: true
  scheduledRollout:
    - date: 2020-01-01T00:00:00Z
      disable: false
future-off:
  variations:
    enabled: true
  defaultRule:
    variation: enabled
  scheduledRollout:
    - date: 2999-01-01T00:00:00Z
      disable: true
experiment-over:
  variations:
    enabled: true
  defaultRule:
    variation: enabled
  experimentation:
    start: 2020-01-01T00:00:00Z
    end: 2020-02-01T00:00:00Z
percentage-with-bucketing-key:
  bucketingKey: teamId
  variations:
    enabled: true
    disabled: false
  defaultRule:
    percentage:
      enabled: 50
      disabled: 50
disabled-with-nested-bucketing-key:
  bucketingKey: company.id
  variations:
    enabled: true
  defaultRule:
    variation: enabled
  disable: true
percentage-with-nested-bucketing-key:
  bucketingKey: company.id
  variations:
    enabled: true
    disabled: false
  defaultRule:
    percentage:
      enabled: 50
      disabled: 50
scheduled-off-with-bucketing-key:
  bucketingKey: teamId
  variations:
    enabled: true
  defaultRule:
    variation: enabled
  scheduledRollout:
    - date: 2020-01-01T00:00:00Z
      disable: true
experiment-over-with-nested-bucketing-key:
  bucketingKey: company.id
  variations:
    enabled: true
  defaultRule:
    variation: enabled
  experimentation:
    start: 2020-01-01T00:00:00Z
    end: 2020-02-01T00:00:00Z
`)
	m, err := NewMetrics()
	require.NoError(t, err)
	require.NoError(t, m.RegisterFlagLoaded(fakeFlagsetSource{
		flagsets: map[string]*ffclient.GoFeatureFlag{"default": client},
	}))

	expected := `
# HELP gofeatureflag_flag_loaded Flags currently loaded in the relay proxy cache, one series per flag with value 1.
# TYPE gofeatureflag_flag_loaded gauge
gofeatureflag_flag_loaded{disabled="true",flag_name="scheduled-off",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="false",flag_name="scheduled-back-on",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="false",flag_name="future-off",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="true",flag_name="experiment-over",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="false",flag_name="percentage-with-bucketing-key",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="true",flag_name="disabled-with-nested-bucketing-key",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="false",flag_name="percentage-with-nested-bucketing-key",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="true",flag_name="scheduled-off-with-bucketing-key",flagset="default"} 1
gofeatureflag_flag_loaded{disabled="true",flag_name="experiment-over-with-nested-bucketing-key",flagset="default"} 1
`
	assert.NoError(t, testutil.GatherAndCompare(m.Registry, strings.NewReader(expected), "gofeatureflag_flag_loaded"))
}

func TestFlagLoaded_DeletedFlagDisappearsAfterRefresh(t *testing.T) {
	client, path := newFileClient(t, flagsYAML)
	m, err := NewMetrics()
	require.NoError(t, err)
	require.NoError(t, m.RegisterFlagLoaded(fakeFlagsetSource{
		flagsets: map[string]*ffclient.GoFeatureFlag{"default": client},
	}))
	count, err := testutil.GatherAndCount(m.Registry, "gofeatureflag_flag_loaded")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	require.NoError(t, os.WriteFile(path, []byte(`
flag-on:
  variations:
    enabled: true
  defaultRule:
    variation: enabled
`), 0o600))
	require.True(t, client.ForceRefresh())

	expected := `
# HELP gofeatureflag_flag_loaded Flags currently loaded in the relay proxy cache, one series per flag with value 1.
# TYPE gofeatureflag_flag_loaded gauge
gofeatureflag_flag_loaded{disabled="false",flag_name="flag-on",flagset="default"} 1
`
	assert.NoError(t, testutil.GatherAndCompare(m.Registry, strings.NewReader(expected), "gofeatureflag_flag_loaded"))
}

func TestFlagLoaded_NoSeriesWhenFlagsetsUnavailable(t *testing.T) {
	tests := []struct {
		name   string
		source FlagsetSource
	}{
		{name: "source error", source: fakeFlagsetSource{err: errors.New("no flagsets configured")}},
		{name: "nil flagset", source: fakeFlagsetSource{flagsets: map[string]*ffclient.GoFeatureFlag{"default": nil}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetrics()
			require.NoError(t, err)
			require.NoError(t, m.RegisterFlagLoaded(tt.source))
			count, err := testutil.GatherAndCount(m.Registry, "gofeatureflag_flag_loaded")
			require.NoError(t, err)
			assert.Equal(t, 0, count)
		})
	}
}

func TestFlagLoaded_RegisterIsNoopWithoutRegistryOrSource(t *testing.T) {
	assert.NoError(t, (&Metrics{}).RegisterFlagLoaded(fakeFlagsetSource{}))
	m, err := NewMetrics()
	require.NoError(t, err)
	assert.NoError(t, m.RegisterFlagLoaded(nil))
}

func TestFlagLoaded_SkipsSeriesWithInvalidLabelValues(t *testing.T) {
	valid, _ := newFileClient(t, flagsYAML)
	invalid, _ := newFileClient(t, flagsYAML)
	m, err := NewMetrics()
	require.NoError(t, err)
	require.NoError(t, m.RegisterFlagLoaded(fakeFlagsetSource{
		flagsets: map[string]*ffclient.GoFeatureFlag{"team-a": valid, "bad-\xff": invalid},
	}))

	expected := `
# HELP gofeatureflag_flag_loaded Flags currently loaded in the relay proxy cache, one series per flag with value 1.
# TYPE gofeatureflag_flag_loaded gauge
gofeatureflag_flag_loaded{disabled="false",flag_name="flag-on",flagset="team-a"} 1
gofeatureflag_flag_loaded{disabled="true",flag_name="flag-off",flagset="team-a"} 1
`
	assert.NotPanics(t, func() {
		assert.NoError(t, testutil.GatherAndCompare(m.Registry, strings.NewReader(expected), "gofeatureflag_flag_loaded"))
	})
}
