package redshiftexporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	defaultFeatureTableName  = "feature_flag_evaluations"
	defaultTrackingTableName = "tracking_events"

	featureEventType  = "feature"
	trackingEventType = "tracking"

	// maxStatementBytes bounds the size of a single INSERT (SQL text + parameters).
	// The Redshift Data API rejects statements larger than 200KB, we keep a safe margin.
	maxStatementBytes = 100 * 1024
	// maxRowsPerStatement bounds the number of rows (and parameters) of a single INSERT.
	maxRowsPerStatement = 500
	// maxStatementsPerBatch is the maximum number of SQL statements of a BatchExecuteStatement call.
	maxStatementsPerBatch = 40
	// maxBatchBytes bounds the size of a BatchExecuteStatement call (SQL text + parameters).
	maxBatchBytes = 1 << 20
	// maxSuperBytes is the maximum size of a JSON value: it is sent as a VARCHAR parameter to JSON_PARSE,
	// and a Redshift VARCHAR cannot be longer than 65535 bytes.
	maxSuperBytes = 65535
	// paramIndexUpperBound is used to compute an upper bound of the size of a row,
	// whatever the index of its parameters in the batch.
	paramIndexUpperBound = 9999999

	defaultPollInterval = 500 * time.Millisecond
	// defaultStatementTimeout bounds the time spent waiting for a transaction.
	defaultStatementTimeout = 5 * time.Minute
	// describeWaitSeconds makes DescribeStatement wait (long polling) for the statement to complete.
	describeWaitSeconds = 10
	// defaultCancelTimeout bounds the time spent cancelling a statement that did not finish in time.
	defaultCancelTimeout = 10 * time.Second
)

// tableNameRegex accepts an unquoted Redshift identifier, optionally prefixed by a schema.
// The table name cannot be sent as a parameter, so we validate it before using it in the SQL.
var tableNameRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$]*(\.[A-Za-z_][A-Za-z0-9_$]*)?$`)

// IsValidTableName returns true if the name can be used as TableName.
// An empty name is valid, the default table name is used in that case.
func IsValidTableName(name string) bool {
	return name == "" || tableNameRegex.MatchString(name)
}

var _ exporter.Exporter = &Exporter{}

// redshiftDataClient is the subset of the Redshift Data API client used by the exporter.
type redshiftDataClient interface {
	BatchExecuteStatement(
		ctx context.Context,
		params *redshiftdata.BatchExecuteStatementInput,
		optFns ...func(*redshiftdata.Options),
	) (*redshiftdata.BatchExecuteStatementOutput, error)
	DescribeStatement(
		ctx context.Context,
		params *redshiftdata.DescribeStatementInput,
		optFns ...func(*redshiftdata.Options),
	) (*redshiftdata.DescribeStatementOutput, error)
	CancelStatement(
		ctx context.Context,
		params *redshiftdata.CancelStatementInput,
		optFns ...func(*redshiftdata.Options),
	) (*redshiftdata.CancelStatementOutput, error)
}

type columnType int

const (
	textColumn columnType = iota
	bigintColumn
	booleanColumn
	superColumn
)

type column struct {
	name    string
	sqlType string
	kind    columnType
	// maxBytes is the size of a text column, VARCHAR lengths are in bytes.
	maxBytes int
}

const (
	sqlTypeBigint  = "BIGINT"
	sqlTypeBoolean = "BOOLEAN"
	sqlTypeSuper   = "SUPER"
)

func textCol(name string, maxBytes int) column {
	return column{name: name, sqlType: fmt.Sprintf("VARCHAR(%d)", maxBytes), kind: textColumn, maxBytes: maxBytes}
}

var featureEventColumns = []column{
	textCol("kind", 50),
	textCol("context_kind", 50),
	textCol("user_key", 512),
	{name: "creation_date", sqlType: sqlTypeBigint, kind: bigintColumn},
	textCol("key", 256),
	textCol("variation", 256),
	{name: "value", sqlType: sqlTypeSuper, kind: superColumn},
	{name: "default", sqlType: sqlTypeBoolean, kind: booleanColumn},
	textCol("version", 64),
	textCol("source", 64),
	{name: "metadata", sqlType: sqlTypeSuper, kind: superColumn},
}

var trackingEventColumns = []column{
	textCol("kind", 50),
	textCol("context_kind", 50),
	textCol("user_key", 512),
	{name: "creation_date", sqlType: sqlTypeBigint, kind: bigintColumn},
	textCol("key", 256),
	{name: "tracking_details", sqlType: sqlTypeSuper, kind: superColumn},
	{name: "evaluation_context", sqlType: sqlTypeSuper, kind: superColumn},
}

// Exporter inserts events in an AWS Redshift table using the Redshift Data API.
type Exporter struct {
	// ClusterIdentifier is the identifier of a provisioned Redshift cluster.
	// Set either ClusterIdentifier or WorkgroupName.
	ClusterIdentifier string

	// WorkgroupName is the name or ARN of a Redshift Serverless workgroup.
	// Set either ClusterIdentifier or WorkgroupName.
	WorkgroupName string

	// Database is the name of the database receiving the events.
	Database string

	// DbUser is the database user, used with ClusterIdentifier to get temporary credentials.
	// If DbUser and SecretArn are empty, the database user is derived from the IAM identity.
	DbUser string

	// SecretArn is the ARN of an AWS Secrets Manager secret containing the database credentials.
	SecretArn string

	// TableName is the table receiving events for this exporter instance, it can be prefixed by a schema.
	// Default: feature_flag_evaluations for feature events, tracking_events for tracking events.
	TableName string

	// AwsConfig is the AWS SDK configuration used to call the Redshift Data API.
	// If nil, the default AWS configuration is loaded.
	AwsConfig *aws.Config

	// AutoMigrate creates the table with the expected schema if it does not exist.
	AutoMigrate bool

	// client is the Redshift Data API client, it can be set for testing purposes.
	client redshiftDataClient

	initOnce sync.Once
	initErr  error

	migrateMutex sync.Mutex
	migrated     bool

	// pollInterval, statementTimeout and cancelTimeout can be overridden for testing purposes.
	pollInterval     time.Duration
	statementTimeout time.Duration
	cancelTimeout    time.Duration
}

// row is a row ready to be inserted, values are in the order of the table columns.
type row struct {
	key    string
	values []string
	// size is an upper bound of the size of the row in an INSERT (SQL + parameters).
	size int
}

// chunk is a group of rows inserted by a single INSERT statement.
type chunk struct {
	rows []row
	// size is an upper bound of the size of the statement (SQL + parameters).
	size int
}

type batch struct {
	sqls   []string
	params []types.SqlParameter
}

// Export inserts the events in the configured Redshift table.
// Each group of up to maxStatementsPerBatch INSERT statements (and maxBatchBytes) runs in a single transaction.
func (e *Exporter) Export(
	ctx context.Context,
	logger *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	rows, eventType, err := buildRows(events)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	if err := e.validate(); err != nil {
		return err
	}
	table, columns := e.tableName(eventType), columnsForEventType(eventType)
	if !IsValidTableName(table) {
		return fmt.Errorf("redshift exporter: invalid table name %q", table)
	}

	rows, skippedKeys := prepareRows(columns, rows, len(insertPrefix(table, columns)))
	if len(skippedKeys) > 0 {
		logger.Warn("redshift exporter: events skipped because they are too large for Redshift",
			"count", len(skippedKeys), "keys", uniqueKeys(skippedKeys))
	}
	if len(rows) == 0 {
		return nil
	}

	if err := e.initClient(ctx); err != nil {
		return err
	}
	if err := e.migrate(ctx, table, columns); err != nil {
		return err
	}
	for _, b := range buildBatches(table, columns, rows) {
		if err := e.run(ctx, b.sqls, b.params); err != nil {
			return err
		}
	}
	return nil
}

// IsBulk returns true because events are inserted in batches.
func (e *Exporter) IsBulk() bool {
	return true
}

func (e *Exporter) validate() error {
	if e.Database == "" {
		return errors.New("redshift exporter: Database is required")
	}
	if e.ClusterIdentifier == "" && e.WorkgroupName == "" {
		return errors.New("redshift exporter: one of ClusterIdentifier or WorkgroupName is required")
	}
	if e.ClusterIdentifier != "" && e.WorkgroupName != "" {
		return errors.New("redshift exporter: ClusterIdentifier and WorkgroupName are mutually exclusive")
	}
	return nil
}

func (e *Exporter) tableName(eventType string) string {
	if e.TableName != "" {
		return e.TableName
	}
	if eventType == trackingEventType {
		return defaultTrackingTableName
	}
	return defaultFeatureTableName
}

func (e *Exporter) initClient(ctx context.Context) error {
	e.initOnce.Do(func() {
		if e.client != nil {
			return
		}
		if e.AwsConfig == nil {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				e.initErr = fmt.Errorf("redshift exporter: impossible to load AWS config: %w", err)
				return
			}
			e.AwsConfig = &cfg
		}
		e.client = redshiftdata.NewFromConfig(*e.AwsConfig)
	})
	return e.initErr
}

// migrate creates the table if AutoMigrate is enabled. It runs once per exporter
// instance and is retried on the next export if it fails.
func (e *Exporter) migrate(ctx context.Context, table string, columns []column) error {
	if !e.AutoMigrate {
		return nil
	}
	e.migrateMutex.Lock()
	defer e.migrateMutex.Unlock()
	if e.migrated {
		return nil
	}

	definitions := make([]string, 0, len(columns))
	for _, c := range columns {
		definitions = append(definitions, quoteIdentifier(c.name)+" "+c.sqlType)
	}
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table, strings.Join(definitions, ", "))
	if err := e.run(ctx, []string{sql}, nil); err != nil {
		return err
	}
	e.migrated = true
	return nil
}

// run executes the statements in a single transaction and waits for their completion,
// the Redshift Data API being asynchronous. The transaction is cancelled if it does not
// complete before the statement timeout.
func (e *Exporter) run(ctx context.Context, sqls []string, params []types.SqlParameter) error {
	timeout := e.statementTimeout
	if timeout <= 0 {
		timeout = defaultStatementTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	input := &redshiftdata.BatchExecuteStatementInput{
		Database:      aws.String(e.Database),
		Sqls:          sqls,
		Parameters:    params,
		ExecutionMode: types.ExecutionModeTransaction,
	}
	if e.ClusterIdentifier != "" {
		input.ClusterIdentifier = aws.String(e.ClusterIdentifier)
	}
	if e.WorkgroupName != "" {
		input.WorkgroupName = aws.String(e.WorkgroupName)
	}
	if e.DbUser != "" {
		input.DbUser = aws.String(e.DbUser)
	}
	if e.SecretArn != "" {
		input.SecretArn = aws.String(e.SecretArn)
	}

	output, err := e.client.BatchExecuteStatement(ctx, input)
	if err != nil {
		return fmt.Errorf("redshift exporter: execute statement: %w", err)
	}
	return e.waitForStatement(ctx, output.Id)
}

// waitForStatement polls the statement until it completes. DescribeStatement errors are
// considered transient and retried until the context is done, the statement is then cancelled.
func (e *Exporter) waitForStatement(ctx context.Context, id *string) error {
	ticker := time.NewTicker(e.interval())
	defer ticker.Stop()

	var lastErr error
	for {
		output, err := e.client.DescribeStatement(ctx, &redshiftdata.DescribeStatementInput{
			Id:              id,
			WaitTimeSeconds: aws.Int32(describeWaitSeconds),
		})
		if err != nil {
			lastErr = err
		} else if done, err := statementResult(id, output); done {
			return err
		}

		select {
		case <-ctx.Done():
			return e.cancelStatement(ctx, id, lastErr)
		case <-ticker.C:
		}
	}
}

// cancelStatement cancels a statement that did not complete before the context was done.
// The cancellation is asynchronous, the statement is polled until it completes: it returns nil
// if the statement finished in the meantime.
func (e *Exporter) cancelStatement(ctx context.Context, id *string, lastErr error) error {
	timeout := e.cancelTimeout
	if timeout <= 0 {
		timeout = defaultCancelTimeout
	}
	cancelCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	_, cancelErr := e.client.CancelStatement(cancelCtx, &redshiftdata.CancelStatementInput{Id: id})
	if e.finishedAfterCancel(cancelCtx, id) {
		return nil
	}

	errs := []error{fmt.Errorf("redshift exporter: statement %s did not complete: %w", aws.ToString(id), ctx.Err())}
	if lastErr != nil {
		errs = append(errs, fmt.Errorf("last describe statement error: %w", lastErr))
	}
	if cancelErr != nil {
		errs = append(errs, fmt.Errorf("cancel statement: %w", cancelErr))
	}
	return errors.Join(errs...)
}

// finishedAfterCancel polls a cancelled statement until it completes and reports if it finished.
func (e *Exporter) finishedAfterCancel(ctx context.Context, id *string) bool {
	ticker := time.NewTicker(e.interval())
	defer ticker.Stop()
	for {
		output, err := e.client.DescribeStatement(ctx, &redshiftdata.DescribeStatementInput{Id: id})
		if err == nil {
			if done, _ := statementResult(id, output); done {
				return output.Status == types.StatusStringFinished
			}
		}
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
		}
	}
}

func (e *Exporter) interval() time.Duration {
	if e.pollInterval <= 0 {
		return defaultPollInterval
	}
	return e.pollInterval
}

// statementResult reports if the statement is completed and its error if it did not succeed.
func statementResult(id *string, output *redshiftdata.DescribeStatementOutput) (bool, error) {
	switch output.Status {
	case types.StatusStringFinished:
		return true, nil
	case types.StatusStringFailed, types.StatusStringAborted:
		msg := fmt.Sprintf("redshift exporter: statement %s %s", aws.ToString(id), output.Status)
		if detail := statementError(output); detail != "" {
			msg += ": " + detail
		}
		return true, errors.New(msg)
	default:
		return false, nil
	}
}

// statementError returns the error of a statement, or the one of the sub-statement that failed.
func statementError(output *redshiftdata.DescribeStatementOutput) string {
	if aws.ToString(output.Error) != "" {
		return aws.ToString(output.Error)
	}
	for _, sub := range output.SubStatements {
		if aws.ToString(sub.Error) != "" {
			return aws.ToString(sub.Error)
		}
	}
	return ""
}

// buildRows converts the events into rows and reports the event type of the batch.
// Events this exporter does not handle are skipped.
func buildRows(events []exporter.ExportableEvent) ([]row, string, error) {
	rows := make([]row, 0, len(events))
	eventType := ""
	for _, event := range events {
		var values []string
		var rowType string
		var err error
		switch evt := event.(type) {
		case exporter.FeatureEvent:
			values, err = featureEventRow(evt)
			rowType = featureEventType
		case exporter.TrackingEvent:
			values, err = trackingEventRow(evt)
			rowType = trackingEventType
		default:
			continue
		}
		if err != nil {
			return nil, "", err
		}
		if eventType == "" {
			eventType = rowType
		}
		if eventType != rowType {
			return nil, "", errors.New("redshift exporter received mixed event types")
		}
		rows = append(rows, row{key: event.GetKey(), values: values})
	}
	return rows, eventType, nil
}

func columnsForEventType(eventType string) []column {
	if eventType == trackingEventType {
		return trackingEventColumns
	}
	return featureEventColumns
}

func featureEventRow(event exporter.FeatureEvent) ([]string, error) {
	value, err := json.Marshal(event.Value)
	if err != nil {
		return nil, err
	}
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return nil, err
	}
	return []string{
		event.Kind,
		event.ContextKind,
		event.UserKey,
		strconv.FormatInt(event.CreationDate, 10),
		event.Key,
		event.Variation,
		string(value),
		strconv.FormatBool(event.Default),
		event.Version,
		event.Source,
		string(metadata),
	}, nil
}

func trackingEventRow(event exporter.TrackingEvent) ([]string, error) {
	trackingDetails, err := json.Marshal(event.TrackingDetails)
	if err != nil {
		return nil, err
	}
	evaluationContext, err := json.Marshal(event.EvaluationContext)
	if err != nil {
		return nil, err
	}
	return []string{
		event.Kind,
		event.ContextKind,
		event.UserKey,
		strconv.FormatInt(event.CreationDate, 10),
		event.Key,
		string(trackingDetails),
		string(evaluationContext),
	}, nil
}

// prepareRows truncates the text values to the size of their column and skips the rows that
// Redshift would reject: a failed export is retried with the same events, so a single invalid
// row would block the exporter forever. It returns the kept rows and the keys of the skipped ones.
func prepareRows(columns []column, rows []row, prefixSize int) ([]row, []string) {
	kept := make([]row, 0, len(rows))
	var skippedKeys []string
	for _, r := range rows {
		tooLarge := false
		for i, c := range columns {
			switch c.kind {
			case textColumn:
				r.values[i] = truncateUTF8(r.values[i], c.maxBytes)
			case superColumn:
				tooLarge = tooLarge || len(r.values[i]) > maxSuperBytes
			}
		}
		_, _, r.size = buildRowValues(columns, r.values, paramIndexUpperBound)
		if tooLarge || prefixSize+r.size > maxStatementBytes {
			skippedKeys = append(skippedKeys, r.key)
			continue
		}
		kept = append(kept, r)
	}
	return kept, skippedKeys
}

// truncateUTF8 makes the value valid UTF-8 and truncates it to maxBytes without
// splitting a multibyte character.
func truncateUTF8(value string, maxBytes int) string {
	value = strings.ToValidUTF8(value, "\uFFFD")
	if len(value) <= maxBytes {
		return value
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(value[cut]) {
		cut--
	}
	return value[:cut]
}

func uniqueKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	unique := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			unique = append(unique, k)
		}
	}
	return unique
}

func insertPrefix(table string, columns []column) string {
	quoted := make([]string, 0, len(columns))
	for _, c := range columns {
		quoted = append(quoted, quoteIdentifier(c.name))
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, strings.Join(quoted, ", "))
}

// buildBatches builds multi-rows INSERT statements and groups them in batches of up to
// maxStatementsPerBatch statements and maxBatchBytes. Parameters are shared by all the
// statements of a batch, so they are numbered across the whole batch.
func buildBatches(table string, columns []column, rows []row) []batch {
	prefix := insertPrefix(table, columns)
	var batches []batch
	var b batch
	batchBytes := 0
	for _, c := range chunkRows(rows, len(prefix)) {
		if len(b.sqls) > 0 && (len(b.sqls) == maxStatementsPerBatch || batchBytes+c.size > maxBatchBytes) {
			batches = append(batches, b)
			b, batchBytes = batch{}, 0
		}
		var sql strings.Builder
		sql.WriteString(prefix)
		for i, r := range c.rows {
			if i > 0 {
				sql.WriteString(", ")
			}
			values, params, _ := buildRowValues(columns, r.values, len(b.params))
			sql.WriteString(values)
			b.params = append(b.params, params...)
		}
		b.sqls = append(b.sqls, sql.String())
		batchBytes += c.size
	}
	if len(b.sqls) > 0 {
		batches = append(batches, b)
	}
	return batches
}

// chunkRows splits the rows to stay under the Redshift Data API limits for a single statement.
func chunkRows(rows []row, prefixSize int) []chunk {
	var chunks []chunk
	start, size := 0, prefixSize
	for i, r := range rows {
		nbRows := i - start
		if nbRows > 0 && (nbRows == maxRowsPerStatement || size+len(", ")+r.size > maxStatementBytes) {
			chunks = append(chunks, chunk{rows: rows[start:i], size: size})
			start, size = i, prefixSize
		}
		if i > start {
			size += len(", ")
		}
		size += r.size
	}
	if start < len(rows) {
		chunks = append(chunks, chunk{rows: rows[start:], size: size})
	}
	return chunks
}

// buildRowValues returns the "(...)" SQL fragment of a row, its parameters and its size
// (SQL fragment + parameters names and values).
func buildRowValues(columns []column, values []string, firstParam int) (string, []types.SqlParameter, int) {
	placeholders := make([]string, 0, len(columns))
	params := make([]types.SqlParameter, 0, len(columns))
	size := 0
	for i, c := range columns {
		value := values[i]
		switch {
		case c.kind == bigintColumn:
			// The value has been formatted from an int64 by strconv, it is safe to inline.
			placeholders = append(placeholders, value)
		case c.kind == booleanColumn:
			placeholders = append(placeholders, booleanLiteral(value))
		case value == "":
			// The Redshift Data API rejects empty parameter values, we use an empty literal instead.
			placeholders = append(placeholders, "''")
		default:
			name := "p" + strconv.Itoa(firstParam+len(params))
			params = append(params, types.SqlParameter{Name: aws.String(name), Value: aws.String(value)})
			size += len(name) + len(value)
			placeholder := ":" + name
			if c.kind == superColumn {
				placeholder = "JSON_PARSE(" + placeholder + ")"
			}
			placeholders = append(placeholders, placeholder)
		}
	}
	sql := "(" + strings.Join(placeholders, ", ") + ")"
	return sql, params, size + len(sql)
}

func booleanLiteral(value string) string {
	if value == strconv.FormatBool(true) {
		return "TRUE"
	}
	return "FALSE"
}

// quoteIdentifier quotes a column name, some of them (e.g. "default") are reserved words.
func quoteIdentifier(name string) string {
	return `"` + name + `"`
}
