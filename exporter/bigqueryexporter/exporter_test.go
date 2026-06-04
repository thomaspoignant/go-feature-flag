package bigqueryexporter

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"text/template"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"google.golang.org/api/googleapi"
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

func TestExporter_ExportSkipsEmptyBatches(t *testing.T) {
	tests := []struct {
		name   string
		events []exporter.ExportableEvent
	}{
		{
			name:   "empty event list",
			events: nil,
		},
		{
			name: "unsupported event type",
			events: []exporter.ExportableEvent{
				unsupportedEvent{key: "ignored"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initCalled := false
			e := &Exporter{
				ProjectID: "fake-project",
				DatasetID: "fake-dataset",
				newClientFunc: func(
					context.Context,
					string,
					...option.ClientOption,
				) (bigQueryClient, error) {
					initCalled = true
					return newMockBigQueryClient(), nil
				},
			}

			err := e.Export(context.TODO(), nil, tt.events)

			assert.NoError(t, err)
			assert.False(t, initCalled)
			assert.Nil(t, e.client)
		})
	}
}

func TestExporter_ExportErrors(t *testing.T) {
	clientErr := errors.New("client init failed")
	ensureErr := errors.New("ensure table failed")
	putErr := errors.New("put rows failed")

	tests := []struct {
		name                string
		events              []exporter.ExportableEvent
		client              *mockBigQueryClient
		newClientErr        error
		autoMigrate         bool
		wantErrContains     string
		wantNewClientCalled bool
	}{
		{
			name: "client initialization error",
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Key: "flag-key", Value: true},
			},
			newClientErr:        clientErr,
			wantErrContains:     "client init failed",
			wantNewClientCalled: true,
		},
		{
			name: "feature value marshal error",
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Value: func() {}},
			},
			wantErrContains: "unsupported type: func",
		},
		{
			name: "feature metadata marshal error",
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind:     "feature",
					Value:    true,
					Metadata: map[string]any{"bad": func() {}},
				},
			},
			wantErrContains: "unsupported type: func",
		},
		{
			name: "tracking details marshal error",
			events: []exporter.ExportableEvent{
				exporter.TrackingEvent{
					Kind:            "tracking",
					TrackingDetails: map[string]any{"bad": func() {}},
				},
			},
			wantErrContains: "unsupported type: func",
		},
		{
			name: "tracking evaluation context marshal error",
			events: []exporter.ExportableEvent{
				exporter.TrackingEvent{
					Kind:              "tracking",
					TrackingDetails:   map[string]any{"amount": 42},
					EvaluationContext: map[string]any{"bad": func() {}},
				},
			},
			wantErrContains: "unsupported type: func",
		},
		{
			name:        "auto migrate returns error",
			autoMigrate: true,
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Key: "flag-key", Value: true},
			},
			client:              &mockBigQueryClient{ensureTableErr: ensureErr},
			wantErrContains:     "ensure table failed",
			wantNewClientCalled: true,
		},
		{
			name: "feature insert returns error",
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Key: "flag-key", Value: true},
			},
			client:              &mockBigQueryClient{putErr: putErr},
			wantErrContains:     "put rows failed",
			wantNewClientCalled: true,
		},
		{
			name: "tracking insert returns error after feature insert succeeds",
			events: []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Key: "flag-key", Value: true},
				exporter.TrackingEvent{
					Kind:              "tracking",
					Key:               "checkout",
					TrackingDetails:   map[string]any{"amount": 42},
					EvaluationContext: map[string]any{"country": "FR"},
				},
			},
			client: &mockBigQueryClient{
				putErrByTable: map[string]error{"fake-dataset.tracking-table": putErr},
			},
			wantErrContains:     "put rows failed",
			wantNewClientCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.client
			if client == nil {
				client = newMockBigQueryClient()
			}
			client.ensureInitialized()
			newClientCalled := false
			e := &Exporter{
				ProjectID:         "fake-project",
				DatasetID:         "fake-dataset",
				TableName:         "feature-table",
				TrackingTableName: "tracking-table",
				AutoMigrate:       tt.autoMigrate,
				newClientFunc: func(
					context.Context,
					string,
					...option.ClientOption,
				) (bigQueryClient, error) {
					newClientCalled = true
					if tt.newClientErr != nil {
						return nil, tt.newClientErr
					}
					return client, nil
				},
			}

			err := e.Export(context.TODO(), nil, tt.events)

			assert.ErrorContains(t, err, tt.wantErrContains)
			assert.Equal(t, tt.wantNewClientCalled, newClientCalled)
		})
	}
}

func TestExporter_ExportChunksRows(t *testing.T) {
	client := newMockBigQueryClient()
	e := &Exporter{
		ProjectID:         "fake-project",
		DatasetID:         "fake-dataset",
		TableName:         "feature-table",
		TrackingTableName: "tracking-table",
		newClientFunc:     mockNewClientFunc(client),
	}
	events := make([]exporter.ExportableEvent, 0, maxRowsPerInsert+3)
	for i := 0; i < maxRowsPerInsert+3; i++ {
		events = append(events, exporter.FeatureEvent{
			Kind:         "feature",
			UserKey:      fmt.Sprintf("user-%d", i),
			CreationDate: int64(i),
			Key:          "flag-key",
			Value:        true,
		})
	}

	err := e.Export(context.TODO(), nil, events)

	assert.NoError(t, err)
	puts := client.inserters["fake-dataset.feature-table"].puts
	assert.Len(t, puts, 2)
	assert.Len(t, puts[0].([]rowSaver), maxRowsPerInsert)
	assert.Len(t, puts[1].([]rowSaver), 3)
}

func TestExporter_InitClientPassesCredentialsOptions(t *testing.T) {
	tests := []struct {
		name              string
		credentials       []byte
		wantClientOptions int
	}{
		{
			name:              "application default credentials",
			wantClientOptions: 0,
		},
		{
			name:              "explicit credentials",
			credentials:       []byte(`{"type":"authorized_user"}`),
			wantClientOptions: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotProjectID string
			var gotClientOptions int
			client := newMockBigQueryClient()
			e := &Exporter{
				ProjectID:         "fake-project",
				DatasetID:         "fake-dataset",
				TableName:         "feature-table",
				TrackingTableName: "tracking-table",
				GoogleCredentials: tt.credentials,
				newClientFunc: func(
					_ context.Context,
					projectID string,
					opts ...option.ClientOption,
				) (bigQueryClient, error) {
					gotProjectID = projectID
					gotClientOptions = len(opts)
					return client, nil
				},
			}

			err := e.Export(context.TODO(), nil, []exporter.ExportableEvent{
				exporter.FeatureEvent{Kind: "feature", Value: true},
			})

			assert.NoError(t, err)
			assert.Equal(t, "fake-project", gotProjectID)
			assert.Equal(t, tt.wantClientOptions, gotClientOptions)
		})
	}
}

func TestCredentialsType(t *testing.T) {
	tests := []struct {
		name  string
		creds []byte
		want  option.CredentialsType
	}{
		{
			name:  "authorized user",
			creds: []byte(`{"type":"authorized_user"}`),
			want:  option.AuthorizedUser,
		},
		{
			name:  "impersonated service account",
			creds: []byte(`{"type":"impersonated_service_account"}`),
			want:  option.ImpersonatedServiceAccount,
		},
		{
			name:  "external account",
			creds: []byte(`{"type":"external_account"}`),
			want:  option.ExternalAccount,
		},
		{
			name:  "service account",
			creds: []byte(`{"type":"service_account"}`),
			want:  option.ServiceAccount,
		},
		{
			name:  "missing type",
			creds: []byte(`{"client_email":"test@example.com"}`),
			want:  option.ServiceAccount,
		},
		{
			name:  "invalid json",
			creds: []byte(`{`),
			want:  option.ServiceAccount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, credentialsType(tt.creds))
		})
	}
}

func TestRowSaver_Save(t *testing.T) {
	row := rowSaver{
		values:   map[string]bigquery.Value{colKey: "flag-key"},
		insertID: "flag-key|user-key|123",
	}

	values, insertID, err := row.Save()

	assert.NoError(t, err)
	assert.Equal(t, row.values, values)
	assert.Equal(t, row.insertID, insertID)
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "google already exists error",
			err:  &googleapi.Error{Code: 409},
			want: true,
		},
		{
			name: "wrapped google already exists error",
			err:  fmt.Errorf("wrapped: %w", &googleapi.Error{Code: 409}),
			want: true,
		},
		{
			name: "different google error",
			err:  &googleapi.Error{Code: 403},
			want: false,
		},
		{
			name: "non google error",
			err:  errors.New("boom"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAlreadyExists(tt.err))
		})
	}
}

type unsupportedEvent struct {
	key string
}

func (u unsupportedEvent) GetUserKey() string {
	return "user-key"
}

func (u unsupportedEvent) GetKey() string {
	return u.key
}

func (u unsupportedEvent) GetCreationDate() int64 {
	return 123
}

func (u unsupportedEvent) FormatInCSV(_ *template.Template) ([]byte, error) {
	return nil, nil
}

func (u unsupportedEvent) FormatInJSON() ([]byte, error) {
	return nil, nil
}

type ensureTableCall struct {
	dataset string
	table   string
	schema  bigquery.Schema
}

type mockBigQueryClient struct {
	inserters        map[string]*mockInserter
	ensureTableCalls []ensureTableCall
	ensureTableErr   error
	putErr           error
	putErrByTable    map[string]error
}

func newMockBigQueryClient() *mockBigQueryClient {
	client := &mockBigQueryClient{}
	client.ensureInitialized()
	return client
}

func (m *mockBigQueryClient) ensureInitialized() {
	if m.inserters == nil {
		m.inserters = make(map[string]*mockInserter)
	}
}

func (m *mockBigQueryClient) Inserter(dataset, table string) inserter {
	m.ensureInitialized()
	key := dataset + "." + table
	if m.inserters[key] == nil {
		m.inserters[key] = &mockInserter{err: m.putErr}
		if m.putErrByTable != nil {
			m.inserters[key].err = m.putErrByTable[key]
		}
	}
	return m.inserters[key]
}

func (m *mockBigQueryClient) EnsureTable(
	_ context.Context,
	dataset string,
	table string,
	schema bigquery.Schema,
) error {
	if m.ensureTableErr != nil {
		return m.ensureTableErr
	}
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
	err  error
}

func (m *mockInserter) Put(_ context.Context, src any) error {
	if m.err != nil {
		return m.err
	}
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
