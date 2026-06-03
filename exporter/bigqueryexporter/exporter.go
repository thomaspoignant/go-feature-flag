package bigqueryexporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	defaultTableName         = "feature_flag_evaluations"
	defaultTrackingTableName = "tracking_events"

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

	// TableName is the table receiving feature flag evaluation events.
	// Default: feature_flag_evaluations
	TableName string

	// TrackingTableName is the table receiving tracking events.
	// Default: tracking_events
	TrackingTableName string

	// GoogleCredentials is an optional Google credentials JSON.
	// If empty, Application Default Credentials are used.
	GoogleCredentials []byte

	// AutoMigrate creates the dataset and tables with the expected schema if they do not exist.
	AutoMigrate bool

	// newClientFunc is used only for unit testing purposes.
	newClientFunc func(context.Context, string, ...option.ClientOption) (bigQueryClient, error)

	// client is the initialized BigQuery client.
	client bigQueryClient
}

// Export streams feature and tracking events to their configured BigQuery tables.
func (e *Exporter) Export(
	ctx context.Context,
	_ *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	e.applyDefaults()

	featureRows, trackingRows, err := e.buildRows(events)
	if err != nil {
		return err
	}
	if len(featureRows) == 0 && len(trackingRows) == 0 {
		return nil
	}

	if e.client == nil {
		if err := e.initClient(ctx); err != nil {
			return err
		}
	}

	if len(featureRows) > 0 {
		if err := e.exportRows(ctx, e.TableName, featureEventSchema(), featureRows); err != nil {
			return err
		}
	}

	if len(trackingRows) > 0 {
		if err := e.exportRows(ctx, e.TrackingTableName, trackingEventSchema(), trackingRows); err != nil {
			return err
		}
	}

	return nil
}

// IsBulk returns true because BigQuery rows are flushed in batches.
func (e *Exporter) IsBulk() bool {
	return true
}

func (e *Exporter) applyDefaults() {
	if e.TableName == "" {
		e.TableName = defaultTableName
	}
	if e.TrackingTableName == "" {
		e.TrackingTableName = defaultTrackingTableName
	}
}

func (e *Exporter) initClient(ctx context.Context) error {
	if e.newClientFunc == nil {
		e.newClientFunc = newBigQueryClient
	}

	var opts []option.ClientOption
	if len(e.GoogleCredentials) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(credentialsType(e.GoogleCredentials), e.GoogleCredentials))
	}

	client, err := e.newClientFunc(ctx, e.ProjectID, opts...)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

func (e *Exporter) buildRows(events []exporter.ExportableEvent) ([]rowSaver, []rowSaver, error) {
	featureRows := make([]rowSaver, 0)
	trackingRows := make([]rowSaver, 0)

	for _, event := range events {
		switch e := event.(type) {
		case exporter.FeatureEvent:
			row, err := featureEventRow(e)
			if err != nil {
				return nil, nil, err
			}
			featureRows = append(featureRows, row)
		case exporter.TrackingEvent:
			row, err := trackingEventRow(e)
			if err != nil {
				return nil, nil, err
			}
			trackingRows = append(trackingRows, row)
		}
	}

	return featureRows, trackingRows, nil
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
