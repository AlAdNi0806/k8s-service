package utils

import (
	"context"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

// HelperLogger is a wrapper around the global OpenTelemetry logger.
type HelperLogger struct {
	logger log.Logger
}

// NewHelperLogger creates a new logger instance with a specific name.
func NewHelperLogger(name string) HelperLogger {
	return HelperLogger{
		logger: global.Logger(name),
	}
}

// LogInfo emits an INFO level log event.
func (h HelperLogger) LogInfo(ctx context.Context, message string, attributes ...log.KeyValue) {
	h.emit(ctx, log.SeverityInfo, message, attributes)
}

// LogWarn emits a WARN level log event.
func (h HelperLogger) LogWarn(ctx context.Context, message string, attributes ...log.KeyValue) {
	h.emit(ctx, log.SeverityWarn, message, attributes)
}

// LogError emits an ERROR level log event.
func (h HelperLogger) LogError(ctx context.Context, message string, err error, attributes ...log.KeyValue) {
	// Add the error object itself as the final attribute
	if err != nil {
		attributes = append(attributes, log.KeyValue{Key: "error", Value: log.StringValue(err.Error())})
	}
	h.emit(ctx, log.SeverityError, message, attributes)
}

// emit creates and emits the log.Record. This is the core helper logic.
func (h HelperLogger) emit(ctx context.Context, severity log.Severity, message string, attributes []log.KeyValue) {
	record := log.Record{}
	record.SetBody(log.StringValue(message))
	record.SetSeverity(severity)

	// Add time for completeness, though the exporter usually handles this
	record.SetTimestamp(record.Timestamp())

	// Add attributes
	record.AddAttributes(attributes...)

	h.logger.Emit(ctx, record)
}
