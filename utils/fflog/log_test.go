package fflog_test

import (
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestFFLogger_Error(t *testing.T) {
	type fields struct {
		msg           string
		keysAndValues []any
	}
	tests := []struct {
		name        string
		logger      *fflog.FFLogger
		fields      fields
		expectedLog string
	}{
		{
			name: "Test Happy Path - slog",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Error message",
				keysAndValues: nil,
			},
			expectedLog: "ERROR Error message" + "\n",
		},
		{
			name: "Test Happy Path - slog with keys and values",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Error message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "ERROR Error message test=toto toto=test" + "\n",
		},
		{
			name: "Test Happy Path - legacy",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Error message",
				keysAndValues: nil,
			},
			expectedLog: "ERROR Error message" + "\n",
		},
		{
			name: "Test Happy Path - legacy with keys and values",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Error message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "ERROR Error message test=toto toto=test" + "\n",
		},
		{
			name:   "FFLogger nil",
			logger: nil,
			fields: fields{
				msg:           "Error message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
		{
			name:   "FFLogger no logger configured",
			logger: &fflog.FFLogger{},
			fields: fields{
				msg:           "Error message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = os.Remove(file.Name())
			}()

			log.SetOutput(file)
			if tt.logger != nil && tt.logger.LegacyLogger != nil {
				tt.logger.LegacyLogger.SetOutput(file)
			}

			if tt.fields.keysAndValues != nil {
				tt.logger.Error(tt.fields.msg, tt.fields.keysAndValues...)
			} else {
				tt.logger.Error(tt.fields.msg)
			}

			content, err := os.ReadFile(file.Name())
			assert.NoError(t, err)
			if len(string(content)) >= 21 {
				actualWithoutTimestamp := string(content)[20:]
				assert.Equal(t, tt.expectedLog, actualWithoutTimestamp)
			} else {
				assert.Equal(t, tt.expectedLog, string(content))
			}
		})
	}
}

func TestFFLogger_Warn(t *testing.T) {
	type fields struct {
		msg           string
		keysAndValues []any
	}
	tests := []struct {
		name        string
		logger      *fflog.FFLogger
		fields      fields
		expectedLog string
	}{
		{
			name: "Test Happy Path - slog",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Warn message",
				keysAndValues: nil,
			},
			expectedLog: "WARN Warn message" + "\n",
		},
		{
			name: "Test Happy Path - slog with keys and values",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Warn message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "WARN Warn message test=toto toto=test" + "\n",
		},
		{
			name: "Test Happy Path - legacy",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Warn message",
				keysAndValues: nil,
			},
			expectedLog: "WARN Warn message" + "\n",
		},
		{
			name: "Test Happy Path - legacy with keys and values",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Warn message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "WARN Warn message test=toto toto=test" + "\n",
		},
		{
			name:   "FFLogger nil",
			logger: nil,
			fields: fields{
				msg:           "Warn message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
		{
			name:   "FFLogger no logger configured",
			logger: &fflog.FFLogger{},
			fields: fields{
				msg:           "Warn message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = os.Remove(file.Name())
			}()

			log.SetOutput(file)
			if tt.logger != nil && tt.logger.LegacyLogger != nil {
				tt.logger.LegacyLogger.SetOutput(file)
			}

			if tt.fields.keysAndValues != nil {
				tt.logger.Warn(tt.fields.msg, tt.fields.keysAndValues...)
			} else {
				tt.logger.Warn(tt.fields.msg)
			}

			content, err := os.ReadFile(file.Name())
			assert.NoError(t, err)
			if len(string(content)) >= 21 {
				actualWithoutTimestamp := string(content)[20:]
				assert.Equal(t, tt.expectedLog, actualWithoutTimestamp)
			} else {
				assert.Equal(t, tt.expectedLog, string(content))
			}
		})
	}
}

func TestFFLogger_Info(t *testing.T) {
	type fields struct {
		msg           string
		keysAndValues []any
	}
	tests := []struct {
		name        string
		logger      *fflog.FFLogger
		fields      fields
		expectedLog string
	}{
		{
			name: "Test Happy Path - slog",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Info message",
				keysAndValues: nil,
			},
			expectedLog: "INFO Info message" + "\n",
		},
		{
			name: "Test Happy Path - slog with keys and values",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Info message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "INFO Info message test=toto toto=test" + "\n",
		},
		{
			name: "Test Happy Path - legacy",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Info message",
				keysAndValues: nil,
			},
			expectedLog: "INFO Info message" + "\n",
		},
		{
			name: "Test Happy Path - legacy with keys and values",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Info message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "INFO Info message test=toto toto=test" + "\n",
		},
		{
			name:   "FFLogger nil",
			logger: nil,
			fields: fields{
				msg:           "Info message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
		{
			name:   "FFLogger no logger configured",
			logger: &fflog.FFLogger{},
			fields: fields{
				msg:           "Info message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = os.Remove(file.Name())
			}()

			log.SetOutput(file)
			if tt.logger != nil && tt.logger.LegacyLogger != nil {
				tt.logger.LegacyLogger.SetOutput(file)
			}

			if tt.fields.keysAndValues != nil {
				tt.logger.Info(tt.fields.msg, tt.fields.keysAndValues...)
			} else {
				tt.logger.Info(tt.fields.msg)
			}

			content, err := os.ReadFile(file.Name())
			assert.NoError(t, err)
			if len(string(content)) >= 21 {
				actualWithoutTimestamp := string(content)[20:]
				assert.Equal(t, tt.expectedLog, actualWithoutTimestamp)
			} else {
				assert.Equal(t, tt.expectedLog, string(content))
			}
		})
	}
}

func TestFFLogger_Debug(t *testing.T) {
	type fields struct {
		msg           string
		keysAndValues []any
	}
	tests := []struct {
		name        string
		logger      *fflog.FFLogger
		fields      fields
		expectedLog string
	}{
		{
			name: "Test Happy Path - slog",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Debug message",
				keysAndValues: nil,
			},
			expectedLog: "DEBUG Debug message" + "\n",
		},
		{
			name: "Test Happy Path - slog with keys and values",
			logger: &fflog.FFLogger{
				LeveledLogger: slog.Default(),
				LegacyLogger:  log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Debug message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "DEBUG Debug message test=toto toto=test" + "\n",
		},
		{
			name: "Test Happy Path - legacy",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg:           "Debug message",
				keysAndValues: nil,
			},
			expectedLog: "DEBUG Debug message" + "\n",
		},
		{
			name: "Test Happy Path - legacy with keys and values",
			logger: &fflog.FFLogger{
				LegacyLogger: log.New(os.Stdout, "", 0),
			},
			fields: fields{
				msg: "Debug message",
				keysAndValues: []any{
					slog.String("test", "toto"),
					slog.String("toto", "test"),
				},
			},
			expectedLog: "DEBUG Debug message test=toto toto=test" + "\n",
		},
		{
			name:   "FFLogger nil",
			logger: nil,
			fields: fields{
				msg:           "Debug message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
		{
			name:   "FFLogger no logger configured",
			logger: &fflog.FFLogger{},
			fields: fields{
				msg:           "Debug message",
				keysAndValues: nil,
			},
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slog.SetLogLoggerLevel(slog.LevelDebug)
			file, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = os.Remove(file.Name())
			}()

			log.SetOutput(file)
			if tt.logger != nil && tt.logger.LegacyLogger != nil {
				tt.logger.LegacyLogger.SetOutput(file)
			}

			if tt.fields.keysAndValues != nil {
				tt.logger.Debug(tt.fields.msg, tt.fields.keysAndValues...)
			} else {
				tt.logger.Debug(tt.fields.msg)
			}

			content, err := os.ReadFile(file.Name())
			assert.NoError(t, err)
			if len(string(content)) >= 21 {
				actualWithoutTimestamp := string(content)[20:]
				assert.Equal(t, tt.expectedLog, actualWithoutTimestamp)
			} else {
				assert.Equal(t, tt.expectedLog, string(content))
			}
		})
	}
}

func TestConvertToFFLogger(t *testing.T) {
	l := log.New(os.Stdout, "", 0)
	ffl := fflog.ConvertToFFLogger(l)
	assert.Equal(t, ffl.GetLogLogger(slog.LevelInfo), l)
}

func TestGetLogLogger(t *testing.T) {
	l := log.New(os.Stdout, "", 0)
	ffl := &fflog.FFLogger{
		LeveledLogger: slog.Default(),
		LegacyLogger:  l,
	}

	ffl2 := &fflog.FFLogger{
		LegacyLogger: l,
	}

	assert.NotEqual(t, ffl.GetLogLogger(slog.LevelInfo), l)
	assert.Equal(t, ffl2.GetLogLogger(slog.LevelInfo), l)
}
