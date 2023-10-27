// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package trace

import (
	"github.com/AccelByte/go-restful-plugins/v4/pkg/auth/iam"
	"github.com/emicklei/go-restful/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// InstrumentCommonAttributes is a filter that will add span attributes for user id and flight id
func InstrumentCommonAttributes(tracerName string) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		r := req.Request
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		route := req.SelectedRoutePath()
		spanName := route

		flightID := req.HeaderParameter(FlightID)

		var tokenUserID string
		if val := req.Attribute(UserIDAttribute); val != nil {
			tokenUserID = val.(string)
		}

		if jwtClaims := iam.RetrieveJWTClaims(req); jwtClaims != nil {
			// if tokenNamespace, tokenUserID or tokenClientID is empty,
			// fallback get from jwt claims
			if tokenUserID == "" {
				tokenUserID = jwtClaims.Subject
			}
		}

		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(attribute.String("user.id", tokenUserID)),
			oteltrace.WithAttributes(attribute.String("flight.id", flightID)),
		}

		tracer := otel.GetTracerProvider().Tracer(
			tracerName,
		)

		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		req.Request = req.Request.WithContext(ctx)

		chain.ProcessFilter(req, resp)
	}
}
