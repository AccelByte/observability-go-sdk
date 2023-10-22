// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package trace

import (
	"context"

	"github.com/sirupsen/logrus"
)

// logKey is the key to store the logger in the context
type logKey struct{}

// LoggerFromContext extracts the logger from the context. If it's not found, a logger with default settings is returned
// For web API endpoints, the Logger go-restful filter is usually used to add the logger in the request context.
func LoggerFromContext(ctx context.Context) *logrus.Entry {
	val, ok := ctx.Value(logKey{}).(*logrus.Entry)
	if !ok {
		le := logrus.NewEntry(logrus.StandardLogger())
		le.Debug("log not found in context, using default")
		return le.WithField(LogFieldTraceID, TraceIDFromContext(ctx))
	}

	return val
}

// LoggerAddField extracts the logger in the context and adds a field with the given key and value
func LoggerAddField(ctx context.Context, key string, value interface{}) context.Context {
	log := LoggerFromContext(ctx)
	le := log.WithField(key, value)
	le.Level = log.Level
	return context.WithValue(ctx, logKey{}, le)
}

// ContextWithLogger inserts a log entry from l into ctx and returns the updated context. Note that this
// will over-ride any existing log entry in the context. For go-restful endpoints, consider using the
// Logger function instead
func ContextWithLogger(ctx context.Context, l *logrus.Logger) context.Context {
	le := logrus.NewEntry(l)
	le.Level = l.Level
	return context.WithValue(ctx, logKey{}, le)
}

// NewLogEntryContext sets up a log entry with the given format and level and injects it into ctx, which are
// then returned. For any invalid value including empty string, it will fall back to using the defaults
// of logrus
func NewLogEntryContext(ctx context.Context, format, level string) (context.Context, *logrus.Entry) {
	logger := NewLogger(format, level)

	ctx = ContextWithLogger(ctx, logger)
	return ctx, LoggerFromContext(ctx)
}

// NewLogger returns a logger with the format and level set with the given values. For any invalid value including empty string,
// it will fall back to using the defaults of logrus
func NewLogger(format, level string) *logrus.Logger {
	logger := logrus.New()
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.WithError(err).Info("failed to parse level")
	} else {
		logger.SetLevel(lvl)
		logrus.SetLevel(lvl)
	}

	setLogFormat(logger, format)
	return logger
}

func setLogFormat(logger *logrus.Logger, format string) {
	switch format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		break
	}
}
