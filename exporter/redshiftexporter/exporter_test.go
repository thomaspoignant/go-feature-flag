package redshiftexporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestExporter_IsBulk(t *testing.T) {
	t.Parallel()

	e := &Exporter{}

	assert.True(t, e.IsBulk(), "Redshift exporter is a bulk one")
}

func TestExporter_ExportFeatureEvents(t *testing.T) {
	client := &mockRedshiftDataClient{}
	e := &Exporter{
		ClusterIdentifier: "my-cluster",
		Database:          "analytics",
		DbUser:            "goff_writer",
		TableName:         "feature_table",
		client:            client,
	}
	events := []exporter.ExportableEvent{
		exporter.FeatureEvent{
			Kind:         "feature",
			ContextKind:  "user",
			UserKey:      "user-key",
			CreationDate: 1617970547,
			Key:          "flag-key",
			Variation:    "variation-a",
			Value:        map[string]any{"enabled": true},
			Default:      false,
			Version:      "v1",
			Source:       "",
			Metadata:     map[string]any{"provider": "go"},
		},
	}

	err := e.Export(context.TODO(), nil, events)

	require.NoError(t, err)
	require.Len(t, client.executed, 1)
	input := client.executed[0]
	assert.Equal(t, "my-cluster", aws.ToString(input.ClusterIdentifier))
	assert.Equal(t, "analytics", aws.ToString(input.Database))
	assert.Equal(t, "goff_writer", aws.ToString(input.DbUser))
	assert.Nil(t, input.WorkgroupName)
	assert.Nil(t, input.SecretArn)
	assert.Equal(t, types.ExecutionModeTransaction, input.ExecutionMode)
	assert.Equal(t, []string{
		`INSERT INTO feature_table ("kind", "context_kind", "user_key", "creation_date", "key", "variation", ` +
			`"value", "default", "version", "source", "metadata") VALUES ` +
			`(:p0, :p1, :p2, 1617970547, :p3, :p4, JSON_PARSE(:p5), FALSE, :p6, '', JSON_PARSE(:p7))`,
	}, input.Sqls)
	assert.Equal(t, map[string]string{
		"p0": "feature",
		"p1": "user",
		"p2": "user-key",
		"p3": "flag-key",
		"p4": "variation-a",
		"p5": `{"enabled":true}`,
		"p6": "v1",
		"p7": `{"provider":"go"}`,
	}, paramsMap(input.Parameters))
	assert.Equal(t, []string{aws.ToString(client.lastStatementID())}, client.described)
}

func TestExporter_ExportTrackingEvents(t *testing.T) {
	client := &mockRedshiftDataClient{}
	e := &Exporter{
		WorkgroupName: "my-workgroup",
		Database:      "analytics",
		SecretArn:     "arn:aws:secretsmanager:us-east-1:123456789012:secret:goff",
		client:        client,
	}
	events := []exporter.ExportableEvent{
		exporter.TrackingEvent{
			Kind:              "tracking",
			ContextKind:       "user",
			UserKey:           "user-key",
			CreationDate:      1617970547,
			Key:               "checkout",
			TrackingDetails:   map[string]any{"amount": 42},
			EvaluationContext: map[string]any{"country": "FR"},
		},
	}

	err := e.Export(context.TODO(), nil, events)

	require.NoError(t, err)
	require.Len(t, client.executed, 1)
	input := client.executed[0]
	assert.Nil(t, input.ClusterIdentifier)
	assert.Nil(t, input.DbUser)
	assert.Equal(t, "my-workgroup", aws.ToString(input.WorkgroupName))
	assert.Equal(t, "arn:aws:secretsmanager:us-east-1:123456789012:secret:goff", aws.ToString(input.SecretArn))
	assert.Equal(t, []string{
		`INSERT INTO tracking_events ("kind", "context_kind", "user_key", "creation_date", "key", ` +
			`"tracking_details", "evaluation_context") VALUES ` +
			`(:p0, :p1, :p2, 1617970547, :p3, JSON_PARSE(:p4), JSON_PARSE(:p5))`,
	}, input.Sqls)
	assert.Equal(t, map[string]string{
		"p0": "tracking",
		"p1": "user",
		"p2": "user-key",
		"p3": "checkout",
		"p4": `{"amount":42}`,
		"p5": `{"country":"FR"}`,
	}, paramsMap(input.Parameters))
}

func TestExporter_DefaultTableName(t *testing.T) {
	client := &mockRedshiftDataClient{}
	e := &Exporter{ClusterIdentifier: "my-cluster", Database: "analytics", client: client}

	err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{newFeatureEvent(0)})

	require.NoError(t, err)
	require.Len(t, client.executed, 1)
	assert.True(t, strings.HasPrefix(client.executed[0].Sqls[0], "INSERT INTO feature_flag_evaluations ("))
}

func TestExporter_ValuesAreSentAsParameters(t *testing.T) {
	client := &mockRedshiftDataClient{}
	e := &Exporter{ClusterIdentifier: "my-cluster", Database: "analytics", client: client}
	malicious := "x'); DROP TABLE feature_flag_evaluations; --"
	event := newFeatureEvent(0)
	event.UserKey = malicious
	event.Key = malicious

	err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{event})

	require.NoError(t, err)
	require.Len(t, client.executed, 1)
	assert.NotContains(t, client.executed[0].Sqls[0], "DROP TABLE")
	params := paramsMap(client.executed[0].Parameters)
	assert.Equal(t, malicious, params["p2"])
	assert.Equal(t, malicious, params["p3"])
}

func TestExporter_ExportChunksLargeBatches(t *testing.T) {
	tests := []struct {
		name           string
		events         []exporter.ExportableEvent
		wantStatements []int
	}{
		{
			name:           "row count limit",
			events:         newSmallEvents(maxRowsPerStatement*2 + 1),
			wantStatements: []int{3},
		},
		{
			name: "statement size limit",
			events: func() []exporter.ExportableEvent {
				events := make([]exporter.ExportableEvent, 0, 10)
				for i := 0; i < 10; i++ {
					event := newFeatureEvent(i)
					event.Metadata = map[string]any{"big": strings.Repeat("a", maxStatementBytes/4)}
					events = append(events, event)
				}
				return events
			}(),
			wantStatements: []int{4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockRedshiftDataClient{}
			e := &Exporter{ClusterIdentifier: "my-cluster", Database: "analytics", client: client}

			err := e.Export(context.TODO(), nil, tt.events)

			require.NoError(t, err)
			require.Len(t, client.executed, len(tt.wantStatements))
			rows := 0
			for i, input := range client.executed {
				assert.Len(t, input.Sqls, tt.wantStatements[i])
				assert.Equal(t, types.ExecutionModeTransaction, input.ExecutionMode)
				// parameters are shared by the batch: they are all unique and all used
				names := map[string]struct{}{}
				for _, p := range input.Parameters {
					names[":"+aws.ToString(p.Name)] = struct{}{}
				}
				assert.Len(t, names, len(input.Parameters))
				used := map[string]struct{}{}
				for _, placeholder := range paramRegex.FindAllString(strings.Join(input.Sqls, " "), -1) {
					used[placeholder] = struct{}{}
				}
				assert.Equal(t, names, used)
				for _, sql := range input.Sqls {
					assert.LessOrEqual(t, len(sql), maxStatementBytes)
					rows += strings.Count(sql, "JSON_PARSE(:p") / 2
				}
			}
			assert.Equal(t, len(tt.events), rows)
		})
	}
}

func TestBuildBatches_LimitsStatementsPerBatch(t *testing.T) {
	values := []string{"tracking", "", "", "0", "k", "null", "null"}
	rows := make([]row, 0, maxRowsPerStatement*maxStatementsPerBatch+1)
	for i := 0; i < cap(rows); i++ {
		// a tiny size, so only the statements limit applies
		rows = append(rows, row{key: "k", values: values, size: 1})
	}

	batches := buildBatches("tracking_events", trackingEventColumns, rows)

	require.Len(t, batches, 2)
	assert.Len(t, batches[0].sqls, maxStatementsPerBatch)
	assert.Len(t, batches[1].sqls, 1)
	assert.Len(t, batches[1].params, 4)
	assert.Equal(t, "p0", aws.ToString(batches[1].params[0].Name))
}

func TestExporter_ExportLimitsBatchBytes(t *testing.T) {
	client := &mockRedshiftDataClient{}
	e := &Exporter{ClusterIdentifier: "my-cluster", Database: "analytics", client: client}
	events := make([]exporter.ExportableEvent, 0, 60)
	for i := 0; i < 60; i++ {
		event := newFeatureEvent(i)
		event.Metadata = map[string]any{"big": strings.Repeat("a", maxStatementBytes/4)}
		events = append(events, event)
	}

	err := e.Export(context.TODO(), nil, events)

	require.NoError(t, err)
	require.Greater(t, len(client.executed), 1)
	rows := 0
	for _, input := range client.executed {
		assert.Less(t, len(input.Sqls), maxStatementsPerBatch)
		size := 0
		for _, sql := range input.Sqls {
			size += len(sql)
			rows += strings.Count(sql, "JSON_PARSE(:p") / 2
		}
		for _, p := range input.Parameters {
			size += len(aws.ToString(p.Name)) + len(aws.ToString(p.Value))
		}
		assert.LessOrEqual(t, size, maxBatchBytes)
	}
	assert.Equal(t, len(events), rows)
}

func TestExporter_TruncatesTextValues(t *testing.T) {
	client := &mockRedshiftDataClient{}
	e := &Exporter{ClusterIdentifier: "my-cluster", Database: "analytics", client: client}
	event := newFeatureEvent(0)
	event.UserKey = strings.Repeat("a", 511) + "é" // 513 bytes, "é" is 2 bytes
	event.Key = strings.Repeat("k", 300)           // 300 bytes
	event.Variation = strings.Repeat("😀", 70)      // 280 bytes, 4 bytes per rune
	event.Version = "v1\xff"                       // invalid UTF-8
	event.Default = true

	err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{event})

	require.NoError(t, err)
	require.Len(t, client.executed, 1)
	params := paramsMap(client.executed[0].Parameters)
	assert.Equal(t, strings.Repeat("a", 511), params["p2"])
	assert.Equal(t, strings.Repeat("k", 256), params["p3"])
	assert.Equal(t, strings.Repeat("😀", 64), params["p4"])
	assert.Equal(t, "v1\uFFFD", params["p6"])
	assert.Contains(t, client.executed[0].Sqls[0], "JSON_PARSE(:p5), TRUE, :p6")
}

func TestTruncateUTF8(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		maxBytes int
		want     string
	}{
		{name: "shorter than max", value: "abc", maxBytes: 5, want: "abc"},
		{name: "exact size", value: "abcde", maxBytes: 5, want: "abcde"},
		{name: "ascii", value: "abcdef", maxBytes: 5, want: "abcde"},
		{name: "cut in a 2 bytes rune", value: "abcdé", maxBytes: 5, want: "abcd"},
		{name: "cut in a 4 bytes rune", value: "a😀", maxBytes: 4, want: "a"},
		{name: "multibyte runes fit", value: "éé", maxBytes: 4, want: "éé"},
		{name: "invalid UTF-8", value: "a\xffb", maxBytes: 10, want: "a\uFFFDb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateUTF8(tt.value, tt.maxBytes)

			assert.Equal(t, tt.want, got)
			assert.LessOrEqual(t, len(got), tt.maxBytes)
		})
	}
}

func TestExporter_SkipsOversizedRows(t *testing.T) {
	tests := []struct {
		name     string
		oversize func(*exporter.FeatureEvent)
	}{
		{
			name: "JSON value larger than a VARCHAR",
			oversize: func(e *exporter.FeatureEvent) {
				e.Value = strings.Repeat("a", maxSuperBytes)
			},
		},
		{
			name: "row larger than a statement",
			oversize: func(e *exporter.FeatureEvent) {
				e.Value = strings.Repeat("a", 60000)
				e.Metadata = map[string]any{"big": strings.Repeat("b", 60000)}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockRedshiftDataClient{}
			e := &Exporter{ClusterIdentifier: "my-cluster", Database: "analytics", client: client}
			var logs bytes.Buffer
			logger := &fflog.FFLogger{LeveledLogger: slog.New(slog.NewTextHandler(&logs, nil))}
			oversized := newFeatureEvent(1)
			oversized.Key = "big-flag"
			oversized.UserKey = "secret-user"
			tt.oversize(&oversized)

			err := e.Export(context.TODO(), logger,
				[]exporter.ExportableEvent{newFeatureEvent(0), oversized, newFeatureEvent(2)})

			require.NoError(t, err)
			require.Len(t, client.executed, 1)
			assert.Equal(t, 2, strings.Count(client.executed[0].Sqls[0], "JSON_PARSE(:p")/2)
			assert.NotContains(t, strings.Join(client.executed[0].Sqls, ""), "big-flag")
			for _, p := range client.executed[0].Parameters {
				assert.NotEqual(t, "big-flag", aws.ToString(p.Value))
			}
			assert.Contains(t, logs.String(), "level=WARN")
			assert.Contains(t, logs.String(), "count=1")
			assert.Contains(t, logs.String(), "big-flag")
			assert.NotContains(t, logs.String(), "secret-user")
		})
	}
}

func TestExporter_SkipsAllRowsWithoutLogger(t *testing.T) {
	client := &mockRedshiftDataClient{}
	e := &Exporter{ClusterIdentifier: "my-cluster", Database: "analytics", AutoMigrate: true, client: client}
	event := newFeatureEvent(0)
	event.Value = strings.Repeat("a", maxSuperBytes)

	err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{event})

	assert.NoError(t, err)
	assert.Empty(t, client.executed)
}

func TestExporter_AutoMigrate(t *testing.T) {
	tests := []struct {
		name        string
		autoMigrate bool
		events      []exporter.ExportableEvent
		wantCreate  string
	}{
		{
			name:        "feature table",
			autoMigrate: true,
			events:      []exporter.ExportableEvent{newFeatureEvent(0)},
			wantCreate: `CREATE TABLE IF NOT EXISTS feature_flag_evaluations ("kind" VARCHAR(50), ` +
				`"context_kind" VARCHAR(50), "user_key" VARCHAR(512), "creation_date" BIGINT, "key" VARCHAR(256), ` +
				`"variation" VARCHAR(256), "value" SUPER, "default" BOOLEAN, "version" VARCHAR(64), ` +
				`"source" VARCHAR(64), "metadata" SUPER)`,
		},
		{
			name:        "tracking table",
			autoMigrate: true,
			events:      []exporter.ExportableEvent{exporter.TrackingEvent{Kind: "tracking", Key: "checkout"}},
			wantCreate: `CREATE TABLE IF NOT EXISTS tracking_events ("kind" VARCHAR(50), ` +
				`"context_kind" VARCHAR(50), "user_key" VARCHAR(512), "creation_date" BIGINT, "key" VARCHAR(256), ` +
				`"tracking_details" SUPER, "evaluation_context" SUPER)`,
		},
		{
			name:        "disabled",
			autoMigrate: false,
			events:      []exporter.ExportableEvent{newFeatureEvent(0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockRedshiftDataClient{}
			e := &Exporter{
				ClusterIdentifier: "my-cluster",
				Database:          "analytics",
				AutoMigrate:       tt.autoMigrate,
				client:            client,
			}

			require.NoError(t, e.Export(context.TODO(), nil, tt.events))
			require.NoError(t, e.Export(context.TODO(), nil, tt.events))

			var creates []string
			for _, input := range client.executed {
				if strings.HasPrefix(input.Sqls[0], "CREATE TABLE") {
					creates = append(creates, input.Sqls...)
					assert.Nil(t, input.Parameters)
				}
			}
			if tt.wantCreate == "" {
				assert.Empty(t, creates)
				return
			}
			// the table is created only once per exporter instance
			assert.Equal(t, []string{tt.wantCreate}, creates)
			assert.Equal(t, []string{tt.wantCreate}, client.executed[0].Sqls)
		})
	}
}

func TestExporter_AutoMigrateRetriesAfterFailure(t *testing.T) {
	client := &mockRedshiftDataClient{
		executeErrs: []error{errors.New("create failed")},
	}
	e := &Exporter{
		ClusterIdentifier: "my-cluster",
		Database:          "analytics",
		AutoMigrate:       true,
		client:            client,
	}
	events := []exporter.ExportableEvent{newFeatureEvent(0)}

	assert.ErrorContains(t, e.Export(context.TODO(), nil, events), "create failed")
	require.NoError(t, e.Export(context.TODO(), nil, events))

	require.Len(t, client.executed, 3)
	assert.True(t, strings.HasPrefix(client.executed[1].Sqls[0], "CREATE TABLE IF NOT EXISTS"))
	assert.True(t, strings.HasPrefix(client.executed[2].Sqls[0], "INSERT INTO"))
}

func TestExporter_WaitsForStatement(t *testing.T) {
	tests := []struct {
		name              string
		client            *mockRedshiftDataClient
		timeout           time.Duration
		wantErr           []string
		wantErrIs         error
		wantPolls         int
		wantCancel        bool
		wantDescribeInput bool
	}{
		{
			name: "polls until finished",
			client: &mockRedshiftDataClient{statuses: []types.StatusString{
				types.StatusStringSubmitted, types.StatusStringStarted, types.StatusStringFinished,
			}},
			wantPolls:         3,
			wantDescribeInput: true,
		},
		{
			name: "failed statement",
			client: &mockRedshiftDataClient{
				statuses:    []types.StatusString{types.StatusStringPicked, types.StatusStringFailed},
				statusError: "ERROR: relation \"feature_flag_evaluations\" does not exist",
			},
			wantErr:   []string{"FAILED: ERROR: relation \"feature_flag_evaluations\" does not exist"},
			wantPolls: 2,
		},
		{
			name: "failed sub statement",
			client: &mockRedshiftDataClient{
				statuses:          []types.StatusString{types.StatusStringFailed},
				subStatementError: "ERROR: value too long",
			},
			wantErr:   []string{"FAILED: ERROR: value too long"},
			wantPolls: 1,
		},
		{
			name:      "aborted statement",
			client:    &mockRedshiftDataClient{statuses: []types.StatusString{types.StatusStringAborted}},
			wantErr:   []string{"ABORTED"},
			wantPolls: 1,
		},
		{
			name:      "transient describe errors",
			client:    &mockRedshiftDataClient{describeErrs: []error{errors.New("throttled"), errors.New("throttled")}},
			wantPolls: 3,
		},
		{
			name:       "describe errors until the deadline",
			client:     &mockRedshiftDataClient{describeErr: errors.New("throttled")},
			timeout:    30 * time.Millisecond,
			wantErr:    []string{"did not complete", "last describe statement error: throttled"},
			wantErrIs:  context.DeadlineExceeded,
			wantCancel: true,
		},
		{
			name: "timeout then cancelled",
			client: &mockRedshiftDataClient{
				statuses:    []types.StatusString{types.StatusStringStarted},
				afterCancel: []types.StatusString{types.StatusStringStarted, types.StatusStringAborted},
			},
			timeout:    30 * time.Millisecond,
			wantErr:    []string{"did not complete"},
			wantErrIs:  context.DeadlineExceeded,
			wantCancel: true,
		},
		{
			name: "timeout and cancel error",
			client: &mockRedshiftDataClient{
				statuses:  []types.StatusString{types.StatusStringStarted},
				cancelErr: errors.New("cancel denied"),
			},
			timeout:    30 * time.Millisecond,
			wantErr:    []string{"did not complete", "cancel statement: cancel denied"},
			wantErrIs:  context.DeadlineExceeded,
			wantCancel: true,
		},
		{
			name: "timeout then finished while cancelling",
			client: &mockRedshiftDataClient{
				statuses: []types.StatusString{types.StatusStringStarted},
				afterCancel: []types.StatusString{
					types.StatusStringStarted, types.StatusStringStarted, types.StatusStringFinished,
				},
			},
			timeout:    30 * time.Millisecond,
			wantCancel: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Exporter{
				ClusterIdentifier: "my-cluster",
				Database:          "analytics",
				client:            tt.client,
				pollInterval:      time.Millisecond,
				statementTimeout:  tt.timeout,
				cancelTimeout:     20 * time.Millisecond,
			}

			err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{newFeatureEvent(0)})

			if len(tt.wantErr) == 0 {
				assert.NoError(t, err)
			}
			for _, want := range tt.wantErr {
				assert.ErrorContains(t, err, want)
			}
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}
			if tt.wantPolls > 0 {
				assert.Len(t, tt.client.described, tt.wantPolls)
			}
			if tt.wantCancel {
				assert.Equal(t, []string{aws.ToString(tt.client.lastStatementID())}, tt.client.cancelled)
			} else {
				assert.Empty(t, tt.client.cancelled)
			}
			if tt.wantDescribeInput {
				assert.Equal(t, int32(describeWaitSeconds), aws.ToInt32(tt.client.lastDescribeInput.WaitTimeSeconds))
			}
		})
	}
}

func TestExporter_StatementTimeoutIsPerTransaction(t *testing.T) {
	// every transaction takes 100ms, the export takes longer than the timeout of a single transaction
	client := &mockRedshiftDataClient{finishAfter: 100 * time.Millisecond}
	e := &Exporter{
		ClusterIdentifier: "my-cluster",
		Database:          "analytics",
		AutoMigrate:       true,
		client:            client,
		pollInterval:      time.Millisecond,
		statementTimeout:  150 * time.Millisecond,
		cancelTimeout:     20 * time.Millisecond,
	}

	err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{newFeatureEvent(0)})

	require.NoError(t, err)
	require.Len(t, client.executed, 2)
	assert.True(t, strings.HasPrefix(client.executed[0].Sqls[0], "CREATE TABLE"))
	assert.True(t, strings.HasPrefix(client.executed[1].Sqls[0], "INSERT INTO"))
	assert.Empty(t, client.cancelled)
}

func TestExporter_StatementTimeoutCancelsTheTransaction(t *testing.T) {
	client := &mockRedshiftDataClient{
		statuses:    []types.StatusString{types.StatusStringStarted},
		afterCancel: []types.StatusString{types.StatusStringAborted},
	}
	e := &Exporter{
		ClusterIdentifier: "my-cluster",
		Database:          "analytics",
		AutoMigrate:       true,
		client:            client,
		pollInterval:      time.Millisecond,
		statementTimeout:  30 * time.Millisecond,
	}

	err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{newFeatureEvent(0)})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	// the migration has been cancelled, no insert has been sent
	require.Len(t, client.executed, 1)
	assert.True(t, strings.HasPrefix(client.executed[0].Sqls[0], "CREATE TABLE"))
	assert.Len(t, client.cancelled, 1)
}

func TestExporter_ExportStopsOnContextCancellation(t *testing.T) {
	client := &mockRedshiftDataClient{statuses: []types.StatusString{types.StatusStringStarted}}
	e := &Exporter{
		ClusterIdentifier: "my-cluster",
		Database:          "analytics",
		client:            client,
		pollInterval:      time.Hour,
		cancelTimeout:     20 * time.Millisecond,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := e.Export(ctx, nil, []exporter.ExportableEvent{newFeatureEvent(0)})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Len(t, client.cancelled, 1)
}

func TestExporter_ExportSkipsEmptyBatches(t *testing.T) {
	tests := []struct {
		name   string
		events []exporter.ExportableEvent
	}{
		{name: "empty event list", events: []exporter.ExportableEvent{}},
		{name: "unsupported event type", events: []exporter.ExportableEvent{unsupportedEvent{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockRedshiftDataClient{}
			// no database configured: an empty batch must not even validate the config
			e := &Exporter{client: client}

			assert.NoError(t, e.Export(context.TODO(), nil, tt.events))
			assert.Empty(t, client.executed)
		})
	}
}

func TestExporter_ExportErrors(t *testing.T) {
	featureEvents := []exporter.ExportableEvent{newFeatureEvent(0)}
	tests := []struct {
		name     string
		exporter *Exporter
		events   []exporter.ExportableEvent
		wantErr  string
	}{
		{
			name:     "missing database",
			exporter: &Exporter{ClusterIdentifier: "my-cluster"},
			events:   featureEvents,
			wantErr:  "redshift exporter: Database is required",
		},
		{
			name:     "missing cluster and workgroup",
			exporter: &Exporter{Database: "analytics"},
			events:   featureEvents,
			wantErr:  "redshift exporter: one of ClusterIdentifier or WorkgroupName is required",
		},
		{
			name:     "both cluster and workgroup",
			exporter: &Exporter{Database: "analytics", ClusterIdentifier: "my-cluster", WorkgroupName: "my-workgroup"},
			events:   featureEvents,
			wantErr:  "redshift exporter: ClusterIdentifier and WorkgroupName are mutually exclusive",
		},
		{
			name: "invalid table name",
			exporter: &Exporter{
				Database: "analytics", ClusterIdentifier: "my-cluster", TableName: "events; DROP TABLE users",
			},
			events:  featureEvents,
			wantErr: `redshift exporter: invalid table name "events; DROP TABLE users"`,
		},
		{
			name:     "mixed event types",
			exporter: &Exporter{Database: "analytics", ClusterIdentifier: "my-cluster"},
			events:   []exporter.ExportableEvent{newFeatureEvent(0), exporter.TrackingEvent{Key: "checkout"}},
			wantErr:  "redshift exporter received mixed event types",
		},
		{
			name:     "feature value marshal error",
			exporter: &Exporter{Database: "analytics", ClusterIdentifier: "my-cluster"},
			events:   []exporter.ExportableEvent{exporter.FeatureEvent{Key: "flag", Value: make(chan int)}},
			wantErr:  "unsupported type",
		},
		{
			name:     "feature metadata marshal error",
			exporter: &Exporter{Database: "analytics", ClusterIdentifier: "my-cluster"},
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Key: "flag", Metadata: map[string]any{"bad": make(chan int)}},
			},
			wantErr: "unsupported type",
		},
		{
			name:     "tracking details marshal error",
			exporter: &Exporter{Database: "analytics", ClusterIdentifier: "my-cluster"},
			events: []exporter.ExportableEvent{
				exporter.TrackingEvent{Key: "checkout", TrackingDetails: map[string]any{"bad": make(chan int)}},
			},
			wantErr: "unsupported type",
		},
		{
			name:     "tracking evaluation context marshal error",
			exporter: &Exporter{Database: "analytics", ClusterIdentifier: "my-cluster"},
			events: []exporter.ExportableEvent{
				exporter.TrackingEvent{Key: "checkout", EvaluationContext: map[string]any{"bad": make(chan int)}},
			},
			wantErr: "unsupported type",
		},
		{
			name: "execute statement error",
			exporter: &Exporter{
				Database: "analytics", ClusterIdentifier: "my-cluster",
				client: &mockRedshiftDataClient{executeErrs: []error{errors.New("access denied")}},
			},
			events:  featureEvents,
			wantErr: "redshift exporter: execute statement: access denied",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exporter.client == nil {
				tt.exporter.client = &mockRedshiftDataClient{}
			}

			err := tt.exporter.Export(context.TODO(), nil, tt.events)

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestIsValidTableName(t *testing.T) {
	tests := []struct {
		name  string
		table string
		want  bool
	}{
		{name: "empty", table: "", want: true},
		{name: "table", table: "feature_flag_evaluations", want: true},
		{name: "schema and table", table: "analytics.goff_events$1", want: true},
		{name: "starts with a digit", table: "1events", want: false},
		{name: "too many parts", table: "db.schema.table", want: false},
		{name: "injection", table: "events; DROP TABLE users", want: false},
		{name: "quoted", table: `"events"`, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTableName(tt.table))
		})
	}
}

func TestExporter_ErrorsDoNotLeakCredentials(t *testing.T) {
	client := &mockRedshiftDataClient{
		statuses:    []types.StatusString{types.StatusStringFailed},
		statusError: "permission denied",
	}
	e := &Exporter{
		ClusterIdentifier: "my-cluster",
		Database:          "analytics",
		DbUser:            "goff_writer",
		SecretArn:         "arn:aws:secretsmanager:us-east-1:123456789012:secret:goff",
		client:            client,
	}

	err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{newFeatureEvent(0)})

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "goff_writer")
	assert.NotContains(t, err.Error(), "secretsmanager")
}

func TestExporter_InitClient(t *testing.T) {
	tests := []struct {
		name      string
		awsConfig *aws.Config
		profile   string
		wantErr   string
	}{
		{name: "provided AWS config", awsConfig: &aws.Config{Region: "us-east-1"}},
		{name: "default AWS config"},
		{
			name:    "default AWS config error",
			profile: "goff-redshift-exporter-missing-profile",
			wantErr: "redshift exporter: impossible to load AWS config",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emptyFile := t.TempDir() + "/empty"
			require.NoError(t, os.WriteFile(emptyFile, nil, 0o600))
			t.Setenv("AWS_CONFIG_FILE", emptyFile)
			t.Setenv("AWS_SHARED_CREDENTIALS_FILE", emptyFile)
			t.Setenv("AWS_PROFILE", tt.profile)
			e := &Exporter{AwsConfig: tt.awsConfig}

			err := e.initClient(context.TODO())

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.ErrorContains(t,
					(&Exporter{AwsConfig: tt.awsConfig, ClusterIdentifier: "c", Database: "d"}).
						Export(context.TODO(), nil, []exporter.ExportableEvent{newFeatureEvent(0)}),
					tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.IsType(t, &redshiftdata.Client{}, e.client)
		})
	}
}

type mockRedshiftDataClient struct {
	mu sync.Mutex
	// executeErrs are returned by the successive BatchExecuteStatement calls, nil means success.
	executeErrs []error
	// describeErrs are returned by the first DescribeStatement calls.
	describeErrs []error
	// describeErr is returned by every DescribeStatement call.
	describeErr error
	// statuses are returned by the successive DescribeStatement calls, the last one is repeated.
	// FINISHED is returned when empty.
	statuses          []types.StatusString
	statusError       string
	subStatementError string
	// afterCancel are returned by the DescribeStatement calls once the statement has been cancelled,
	// the last one is repeated.
	afterCancel []types.StatusString
	// finishAfter makes a statement STARTED until this duration elapsed since its execution.
	finishAfter time.Duration
	executedAt  time.Time
	cancelErr   error

	executed          []*redshiftdata.BatchExecuteStatementInput
	described         []string
	lastDescribeInput *redshiftdata.DescribeStatementInput
	cancelled         []string
}

func (m *mockRedshiftDataClient) BatchExecuteStatement(
	_ context.Context,
	params *redshiftdata.BatchExecuteStatementInput,
	_ ...func(*redshiftdata.Options),
) (*redshiftdata.BatchExecuteStatementOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executed = append(m.executed, params)
	m.executedAt = time.Now()
	if len(m.executeErrs) > 0 {
		err := m.executeErrs[0]
		m.executeErrs = m.executeErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	return &redshiftdata.BatchExecuteStatementOutput{Id: m.statementID(len(m.executed))}, nil
}

func (m *mockRedshiftDataClient) DescribeStatement(
	_ context.Context,
	params *redshiftdata.DescribeStatementInput,
	_ ...func(*redshiftdata.Options),
) (*redshiftdata.DescribeStatementOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.described = append(m.described, aws.ToString(params.Id))
	m.lastDescribeInput = params
	if len(m.describeErrs) > 0 {
		err := m.describeErrs[0]
		m.describeErrs = m.describeErrs[1:]
		return nil, err
	}
	if m.describeErr != nil {
		return nil, m.describeErr
	}
	status := types.StatusStringFinished
	if len(m.statuses) > 0 {
		status = m.statuses[0]
		if len(m.statuses) > 1 {
			m.statuses = m.statuses[1:]
		}
	}
	if m.finishAfter > 0 && time.Since(m.executedAt) < m.finishAfter {
		status = types.StatusStringStarted
	}
	if len(m.cancelled) > 0 && len(m.afterCancel) > 0 {
		status = m.afterCancel[0]
		if len(m.afterCancel) > 1 {
			m.afterCancel = m.afterCancel[1:]
		}
	}
	out := &redshiftdata.DescribeStatementOutput{Id: params.Id, Status: status}
	if m.statusError != "" {
		out.Error = aws.String(m.statusError)
	}
	if m.subStatementError != "" {
		out.SubStatements = []types.SubStatementData{{}, {Error: aws.String(m.subStatementError)}}
	}
	return out, nil
}

func (m *mockRedshiftDataClient) CancelStatement(
	_ context.Context,
	params *redshiftdata.CancelStatementInput,
	_ ...func(*redshiftdata.Options),
) (*redshiftdata.CancelStatementOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelled = append(m.cancelled, aws.ToString(params.Id))
	if m.cancelErr != nil {
		return nil, m.cancelErr
	}
	return &redshiftdata.CancelStatementOutput{Status: aws.Bool(true)}, nil
}

func (m *mockRedshiftDataClient) statementID(n int) *string {
	return aws.String(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
}

func (m *mockRedshiftDataClient) lastStatementID() *string {
	return m.statementID(len(m.executed))
}

var paramRegex = regexp.MustCompile(`:p[0-9]+`)

func paramsMap(params []types.SqlParameter) map[string]string {
	res := make(map[string]string, len(params))
	for _, p := range params {
		res[aws.ToString(p.Name)] = aws.ToString(p.Value)
	}
	return res
}

func newFeatureEvent(i int) exporter.FeatureEvent {
	return exporter.FeatureEvent{
		Kind:         "feature",
		ContextKind:  "user",
		UserKey:      fmt.Sprintf("user-%d", i),
		CreationDate: 1617970547,
		Key:          "flag-key",
		Variation:    "variation-a",
		Value:        true,
		Version:      "v1",
		Source:       "SERVER",
	}
}

// newSmallEvents returns events small enough to hit the rows limit before the size limit.
func newSmallEvents(n int) []exporter.ExportableEvent {
	events := make([]exporter.ExportableEvent, 0, n)
	for i := 0; i < n; i++ {
		events = append(events, exporter.TrackingEvent{Key: "k"})
	}
	return events
}

type unsupportedEvent struct{}

func (unsupportedEvent) GetUserKey() string     { return "" }
func (unsupportedEvent) GetKey() string         { return "" }
func (unsupportedEvent) GetCreationDate() int64 { return 0 }
func (unsupportedEvent) FormatInCSV(_ *template.Template) ([]byte, error) {
	return nil, nil
}
func (unsupportedEvent) FormatInJSON() ([]byte, error) { return nil, nil }
