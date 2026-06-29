package bigqueryexporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	defaultTableName = "feature_flag_evaluations"

	featureEventType  = "feature"
	trackingEventType = "tracking"

	// maxRowsPerInsert bounds the number of rows sent in a single streaming
	// insert request. BigQuery limits a streaming insert to 50,000 rows (and
	// 10MB) per request, so we chunk well below that ceiling. The relay proxy's
	// default MaxEventInMemory (100,000) can exceed the limit in a single flush.
	maxRowsPerInsert = 10000

	// Column names shared between the row builders and the table schemas.
	colKind              = "kind"
	colContextKind       = "context_kind"
	colUserKey           = "user_key"
	colCreationDate      = "creation_date"
	colKey               = "key"
	colVariation         = "variation"
	colValue             = "value"
	colDefault           = "default"
	colVersion           = "version"
	colSource            = "source"
	colMetadata          = "metadata"
	colTrackingDetails   = "tracking_details"
	colEvaluationContext = "evaluation_context"
)

var _ exporter.Exporter = &Exporter{}

type bigQueryClient interface {
	Inserter(dataset, table string) inserter
	EnsureTable(ctx context.Context, dataset, table string, schema bigquery.Schema) error
	Close() error
}

type inserter interface {
	Put(ctx context.Context, src any) error
}

type rowSaver struct {
	values   map[string]bigquery.Value
	insertID string
}

func (r rowSaver) Save() (map[string]bigquery.Value, string, error) {
	// Returning a deterministic insertID keeps BigQuery's best-effort
	// de-duplication enabled: if Put retries a request that BigQuery already
	// accepted, rows sharing the same insertID are dropped rather than
	// duplicated. An empty insertID would disable de-duplication.
	return r.values, r.insertID, nil
}

// Exporter streams events to Google BigQuery.
type Exporter struct {
	// ProjectID is the ID of the GCP project containing the BigQuery dataset.
	ProjectID string

	// DatasetID is the BigQuery dataset receiving the events.
	DatasetID string

	// TableName is the table receiving events for this exporter instance.
	// Default: feature_flag_evaluations
	TableName string

	// GoogleCredentials is an optional Google credentials JSON.
	// If empty, Application Default Credentials are used.
	GoogleCredentials []byte

	// AutoMigrate creates the dataset and tables with the expected schema if they do not exist.
	AutoMigrate bool

	// newClientFunc is used only for unit testing purposes.
	newClientFunc func(context.Context, string, ...option.ClientOption) (bigQueryClient, error)

	// client is the initialized BigQuery client.
	client bigQueryClient

	initOnce sync.Once
	initErr  error
}

// Export streams events to the configured BigQuery table.
func (e *Exporter) Export(
	ctx context.Context,
	_ *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	rows, schema, err := e.buildRows(events)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	if err := e.initClient(ctx); err != nil {
		return err
	}

	return e.exportRows(ctx, e.tableName(), schema, rows)
}

// IsBulk returns true because BigQuery rows are flushed in batches.
func (e *Exporter) IsBulk() bool {
	return true
}

func (e *Exporter) tableName() string {
	if e.TableName == "" {
		return defaultTableName
	}
	return e.TableName
}

func (e *Exporter) initClient(ctx context.Context) error {
	e.initOnce.Do(func() {
		if e.client != nil {
			return
		}

		newClientFunc := e.newClientFunc
		if newClientFunc == nil {
			newClientFunc = newBigQueryClient
		}

		var opts []option.ClientOption
		if len(e.GoogleCredentials) > 0 {
			opts = append(opts, option.WithAuthCredentialsJSON(credentialsType(e.GoogleCredentials), e.GoogleCredentials))
		}

		e.client, e.initErr = newClientFunc(ctx, e.ProjectID, opts...)
	})
	return e.initErr
}

func (e *Exporter) buildRows(events []exporter.ExportableEvent) ([]rowSaver, bigquery.Schema, error) {
	rows := make([]rowSaver, 0, len(events))
	eventType := ""

	for _, event := range events {
		row, rowType, err := buildRow(event)
		if err != nil {
			return nil, nil, err
		}
		if rowType == "" {
			// Skip events this exporter does not handle.
			continue
		}
		if eventType == "" {
			eventType = rowType
		}
		if eventType != rowType {
			return nil, nil, errors.New("bigquery exporter received mixed event types")
		}
		rows = append(rows, row)
	}

	return rows, schemaForEventType(eventType), nil
}

// buildRow converts a single exportable event into a BigQuery row and reports
// the event type it belongs to. The returned event type is empty for events
// this exporter does not handle.
func buildRow(event exporter.ExportableEvent) (rowSaver, string, error) {
	switch evt := event.(type) {
	case exporter.FeatureEvent:
		row, err := featureEventRow(evt)
		return row, featureEventType, err
	case exporter.TrackingEvent:
		row, err := trackingEventRow(evt)
		return row, trackingEventType, err
	default:
		return rowSaver{}, "", nil
	}
}

// schemaForEventType returns the BigQuery schema matching the resolved event
// type, or nil when no rows were produced.
func schemaForEventType(eventType string) bigquery.Schema {
	switch eventType {
	case featureEventType:
		return featureEventSchema()
	case trackingEventType:
		return trackingEventSchema()
	default:
		return nil
	}
}

func (e *Exporter) exportRows(
	ctx context.Context,
	table string,
	schema bigquery.Schema,
	rows []rowSaver,
) error {
	if e.AutoMigrate {
		if err := e.client.EnsureTable(ctx, e.DatasetID, table, schema); err != nil {
			return err
		}
	}

	ins := e.client.Inserter(e.DatasetID, table)
	// Chunk rows to stay under BigQuery's per-request streaming-insert limits.
	for start := 0; start < len(rows); start += maxRowsPerInsert {
		end := start + maxRowsPerInsert
		if end > len(rows) {
			end = len(rows)
		}
		if err := ins.Put(ctx, rows[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// insertID builds a deterministic best-effort de-duplication ID for a row so
// that BigQuery drops duplicates when Put retries an already-accepted request.
func insertID(event exporter.ExportableEvent) string {
	return fmt.Sprintf("%s|%s|%d", event.GetKey(), event.GetUserKey(), event.GetCreationDate())
}

func featureEventRow(event exporter.FeatureEvent) (rowSaver, error) {
	value, err := jsonString(event.Value)
	if err != nil {
		return rowSaver{}, err
	}
	metadata, err := jsonString(event.Metadata)
	if err != nil {
		return rowSaver{}, err
	}

	return rowSaver{
		values: map[string]bigquery.Value{
			colKind:         event.Kind,
			colContextKind:  event.ContextKind,
			colUserKey:      event.UserKey,
			colCreationDate: event.CreationDate,
			colKey:          event.Key,
			colVariation:    event.Variation,
			colValue:        value,
			colDefault:      event.Default,
			colVersion:      event.Version,
			colSource:       event.Source,
			colMetadata:     metadata,
		},
		insertID: insertID(event),
	}, nil
}

func trackingEventRow(event exporter.TrackingEvent) (rowSaver, error) {
	trackingDetails, err := jsonString(event.TrackingDetails)
	if err != nil {
		return rowSaver{}, err
	}
	evaluationContext, err := jsonString(event.EvaluationContext)
	if err != nil {
		return rowSaver{}, err
	}

	return rowSaver{
		values: map[string]bigquery.Value{
			colKind:              event.Kind,
			colContextKind:       event.ContextKind,
			colUserKey:           event.UserKey,
			colCreationDate:      event.CreationDate,
			colKey:               event.Key,
			colTrackingDetails:   trackingDetails,
			colEvaluationContext: evaluationContext,
		},
		insertID: insertID(event),
	}, nil
}

func jsonString(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func featureEventSchema() bigquery.Schema {
	return bigquery.Schema{
		{Name: colKind, Type: bigquery.StringFieldType},
		{Name: colContextKind, Type: bigquery.StringFieldType},
		{Name: colUserKey, Type: bigquery.StringFieldType},
		{Name: colCreationDate, Type: bigquery.IntegerFieldType},
		{Name: colKey, Type: bigquery.StringFieldType},
		{Name: colVariation, Type: bigquery.StringFieldType},
		{Name: colValue, Type: bigquery.JSONFieldType},
		{Name: colDefault, Type: bigquery.BooleanFieldType},
		{Name: colVersion, Type: bigquery.StringFieldType},
		{Name: colSource, Type: bigquery.StringFieldType},
		{Name: colMetadata, Type: bigquery.JSONFieldType},
	}
}

func trackingEventSchema() bigquery.Schema {
	return bigquery.Schema{
		{Name: colKind, Type: bigquery.StringFieldType},
		{Name: colContextKind, Type: bigquery.StringFieldType},
		{Name: colUserKey, Type: bigquery.StringFieldType},
		{Name: colCreationDate, Type: bigquery.IntegerFieldType},
		{Name: colKey, Type: bigquery.StringFieldType},
		{Name: colTrackingDetails, Type: bigquery.JSONFieldType},
		{Name: colEvaluationContext, Type: bigquery.JSONFieldType},
	}
}

// credentialsType inspects the "type" field of a Google credentials JSON blob
// and maps it to the option.CredentialsType expected by
// option.WithAuthCredentialsJSON. It defaults to ServiceAccount, which is the
// most common credential shape, when the type is missing or unrecognized.
func credentialsType(creds []byte) option.CredentialsType {
	var parsed struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(creds, &parsed); err != nil {
		return option.ServiceAccount
	}
	switch parsed.Type {
	case "authorized_user":
		return option.AuthorizedUser
	case "impersonated_service_account":
		return option.ImpersonatedServiceAccount
	case "external_account":
		return option.ExternalAccount
	default:
		return option.ServiceAccount
	}
}

func newBigQueryClient(
	ctx context.Context,
	projectID string,
	opts ...option.ClientOption,
) (bigQueryClient, error) {
	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}
	return &realBigQueryClient{client: client}, nil
}

type realBigQueryClient struct {
	client *bigquery.Client
}

func (c *realBigQueryClient) Inserter(dataset, table string) inserter {
	return c.client.Dataset(dataset).Table(table).Inserter()
}

func (c *realBigQueryClient) EnsureTable(
	ctx context.Context,
	dataset string,
	table string,
	schema bigquery.Schema,
) error {
	bqDataset := c.client.Dataset(dataset)
	if err := bqDataset.Create(ctx, nil); err != nil && !isAlreadyExists(err) {
		return err
	}

	err := bqDataset.Table(table).Create(ctx, &bigquery.TableMetadata{Schema: schema})
	if err != nil && !isAlreadyExists(err) {
		return err
	}
	return nil
}

func (c *realBigQueryClient) Close() error {
	return c.client.Close()
}

func isAlreadyExists(err error) bool {
	var googleErr *googleapi.Error
	if errors.As(err, &googleErr) {
		return googleErr.Code == 409
	}
	return false
}
