package bigqueryexporter

import (
	"context"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"google.golang.org/api/option"
)

func TestExporter_IsBulk(t *testing.T) {
	t.Parallel()

	e := &Exporter{}

	assert.True(t, e.IsBulk(), "BigQuery exporter is a bulk one")
}

func TestExporter_ExportFeatureEvents(t *testing.T) {
	ctx := context.TODO()
	client := newMockBigQueryClient()
	e := &Exporter{
		ProjectID:         "fake-project",
		DatasetID:         "fake-dataset",
		TableName:         "feature-table",
		TrackingTableName: "tracking-table",
		newClientFunc:     mockNewClientFunc(client),
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
			Source:       "SERVER",
			Metadata:     map[string]any{"provider": "go"},
		},
	}

	err := e.Export(ctx, nil, events)

	assert.NoError(t, err)
	assert.Empty(t, client.ensureTableCalls)
	assert.Len(t, client.inserters, 1)
	assert.Len(t, client.inserters["fake-dataset.feature-table"].puts, 1)

	rows := client.inserters["fake-dataset.feature-table"].puts[0].([]rowSaver)
	assert.Len(t, rows, 1)
	assert.Equal(t, map[string]bigquery.Value{
		"kind":          "feature",
		"context_kind":  "user",
		"user_key":      "user-key",
		"creation_date": int64(1617970547),
		"key":           "flag-key",
		"variation":     "variation-a",
		"value":         `{"enabled":true}`,
		"default":       false,
		"version":       "v1",
		"source":        "SERVER",
		"metadata":      `{"provider":"go"}`,
	}, rows[0].values)
	assert.Equal(t, "flag-key|user-key|1617970547", rows[0].insertID)
}

func TestExporter_ExportTrackingEvents(t *testing.T) {
	ctx := context.TODO()
	client := newMockBigQueryClient()
	e := &Exporter{
		ProjectID:         "fake-project",
		DatasetID:         "fake-dataset",
		TableName:         "feature-table",
		TrackingTableName: "tracking-table",
		newClientFunc:     mockNewClientFunc(client),
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

	err := e.Export(ctx, nil, events)

	assert.NoError(t, err)
	assert.Empty(t, client.ensureTableCalls)
	assert.Len(t, client.inserters, 1)
	assert.Len(t, client.inserters["fake-dataset.tracking-table"].puts, 1)

	rows := client.inserters["fake-dataset.tracking-table"].puts[0].([]rowSaver)
	assert.Len(t, rows, 1)
	assert.Equal(t, map[string]bigquery.Value{
		"kind":               "tracking",
		"context_kind":       "user",
		"user_key":           "user-key",
		"creation_date":      int64(1617970547),
		"key":                "checkout",
		"tracking_details":   `{"amount":42}`,
		"evaluation_context": `{"country":"FR"}`,
	}, rows[0].values)
	assert.Equal(t, "checkout|user-key|1617970547", rows[0].insertID)
}

func TestExporter_ExportMixedBatch(t *testing.T) {
	ctx := context.TODO()
	client := newMockBigQueryClient()
	e := &Exporter{
		ProjectID:         "fake-project",
		DatasetID:         "fake-dataset",
		TableName:         "feature-table",
		TrackingTableName: "tracking-table",
		newClientFunc:     mockNewClientFunc(client),
	}
	events := []exporter.ExportableEvent{
		exporter.FeatureEvent{
			Kind: "feature", ContextKind: "user", UserKey: "user-key", CreationDate: 1617970547, Key: "flag-key",
			Variation: "variation-a", Value: true, Default: false, Version: "v1", Source: "SERVER",
		},
		exporter.TrackingEvent{
			Kind: "tracking", ContextKind: "user", UserKey: "user-key", CreationDate: 1617970548, Key: "checkout",
			TrackingDetails: map[string]any{"amount": 42}, EvaluationContext: map[string]any{"country": "FR"},
		},
		exporter.FeatureEvent{
			Kind: "feature", ContextKind: "user", UserKey: "user-key-2", CreationDate: 1617970549, Key: "flag-key-2",
			Variation: "variation-b", Value: false, Default: true, Version: "v2", Source: "SERVER",
		},
	}

	err := e.Export(ctx, nil, events)

	assert.NoError(t, err)
	assert.Len(t, client.inserters["fake-dataset.feature-table"].puts, 1)
	assert.Len(t, client.inserters["fake-dataset.tracking-table"].puts, 1)
	assert.Len(t, client.inserters["fake-dataset.feature-table"].puts[0].([]rowSaver), 2)
	assert.Len(t, client.inserters["fake-dataset.tracking-table"].puts[0].([]rowSaver), 1)
}

func TestExporter_AutoMigrate(t *testing.T) {
	tests := []struct {
		name                 string
		autoMigrate          bool
		events               []exporter.ExportableEvent
		wantEnsureTableCalls []ensureTableCall
	}{
		{
			name:        "AutoMigrate true ensures written tables",
			autoMigrate: true,
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Value: true},
				exporter.TrackingEvent{Kind: "tracking"},
			},
			wantEnsureTableCalls: []ensureTableCall{
				{dataset: "fake-dataset", table: "feature-table", schema: featureEventSchema()},
				{dataset: "fake-dataset", table: "tracking-table", schema: trackingEventSchema()},
			},
		},
		{
			name:        "AutoMigrate false does not ensure tables",
			autoMigrate: false,
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Value: true},
				exporter.TrackingEvent{Kind: "tracking"},
			},
			wantEnsureTableCalls: nil,
		},
		{
			name:        "AutoMigrate true only ensures tables being written",
			autoMigrate: true,
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Value: true},
			},
			wantEnsureTableCalls: []ensureTableCall{
				{dataset: "fake-dataset", table: "feature-table", schema: featureEventSchema()},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			client := newMockBigQueryClient()
			e := &Exporter{
				ProjectID:         "fake-project",
				DatasetID:         "fake-dataset",
				TableName:         "feature-table",
				TrackingTableName: "tracking-table",
				AutoMigrate:       tt.autoMigrate,
				newClientFunc:     mockNewClientFunc(client),
			}

			err := e.Export(ctx, nil, tt.events)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantEnsureTableCalls, client.ensureTableCalls)
		})
	}
}

func TestExporter_DefaultTableNames(t *testing.T) {
	ctx := context.TODO()
	client := newMockBigQueryClient()
	e := &Exporter{
		ProjectID:     "fake-project",
		DatasetID:     "fake-dataset",
		newClientFunc: mockNewClientFunc(client),
	}
	events := []exporter.ExportableEvent{
		exporter.FeatureEvent{Kind: "feature", Value: true},
		exporter.TrackingEvent{Kind: "tracking"},
	}

	err := e.Export(ctx, nil, events)

	assert.NoError(t, err)
	assert.Equal(t, defaultTableName, e.TableName)
	assert.Equal(t, defaultTrackingTableName, e.TrackingTableName)
	assert.Len(t, client.inserters["fake-dataset."+defaultTableName].puts, 1)
	assert.Len(t, client.inserters["fake-dataset."+defaultTrackingTableName].puts, 1)
}

type ensureTableCall struct {
	dataset string
	table   string
	schema  bigquery.Schema
}

type mockBigQueryClient struct {
	inserters        map[string]*mockInserter
	ensureTableCalls []ensureTableCall
}

func newMockBigQueryClient() *mockBigQueryClient {
	return &mockBigQueryClient{
		inserters: make(map[string]*mockInserter),
	}
}

func (m *mockBigQueryClient) Inserter(dataset, table string) inserter {
	key := dataset + "." + table
	if m.inserters[key] == nil {
		m.inserters[key] = &mockInserter{}
	}
	return m.inserters[key]
}

func (m *mockBigQueryClient) EnsureTable(
	_ context.Context,
	dataset string,
	table string,
	schema bigquery.Schema,
) error {
	m.ensureTableCalls = append(m.ensureTableCalls, ensureTableCall{
		dataset: dataset,
		table:   table,
		schema:  schema,
	})
	return nil
}

func (m *mockBigQueryClient) Close() error {
	return nil
}

type mockInserter struct {
	puts []any
}

func (m *mockInserter) Put(_ context.Context, src any) error {
	m.puts = append(m.puts, src)
	return nil
}

func mockNewClientFunc(
	client bigQueryClient,
) func(context.Context, string, ...option.ClientOption) (bigQueryClient, error) {
	return func(_ context.Context, _ string, _ ...option.ClientOption) (bigQueryClient, error) {
		return client, nil
	}
}
