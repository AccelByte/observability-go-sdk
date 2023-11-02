// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package trace

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
)

// LogTraceError logs the provided error and message to the logger obtained from the context,
// records the error in the trace span and sets the status of the span to Error.
//
// Parameters:
// ctx: Context in which the function operates, it must contain a valid span and logger.
// err: Error to be logged and recorded. If this is nil, only the error message will be logged.
// errMsg: Message to be logged and recorded. This message is also used as the span status description when error occurs.
//
// This function does not return any values.
func LogTraceError(ctx context.Context, err error, errMsg string, fields ...logrus.Fields) {
	span := SpanFromContext(ctx)
	log := LoggerFromContext(ctx).WithFields(mergeFields(fields...))
	span.SetStatus(codes.Error, errMsg)
	if err == nil {
		log.Error(errMsg)
		return
	}
	log.WithError(err).Error(errMsg)
	span.RecordError(err)
}

// TraceError record the current error in trace span without log message.
//
// Parameters:
// ctx: Context in which the function operates, it must contain a valid span.
// err: Error to be and recorded. If this is nil, this method does nothing.
// errMsg: Message to be recorded. This message is also used as the span status description when error occurs.
func TraceError(ctx context.Context, err error, errMsg string) {
	span := SpanFromContext(ctx)
	span.RecordError(err)
}

// Helper function to merge multiple logrus.Fields dictionaries into one
func mergeFields(fieldsSlice ...logrus.Fields) logrus.Fields {
	result := logrus.Fields{}
	for _, fields := range fieldsSlice {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}
