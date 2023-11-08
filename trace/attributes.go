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
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// InstrumentCommonAttributes is a filter that will add span attributes for user id and flight id
//
// Parameters
// tracerName: tracer name
// excludeEndpoints: consists of blacklisted endpoint name to not send tracer
// eg. key: /healthz , key2: map of http.Method (GET, POST, DELETE, PUT) ,  value : boolean
func InstrumentCommonAttributes(tracerName string, excludeEndpoints map[string]map[string]bool) (filterFunc restful.FilterFunction) {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		r := req.Request
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		route := req.SelectedRoutePath()
		spanName := route

		// end process if route = blacklisted endpoints
		if _, exist := excludeEndpoints[route]; exist {
			if excludeEndpoints[route][r.Method] {
				chain.ProcessFilter(req, resp)
				return
			}
		}

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
			oteltrace.WithAttributes(HTTPServerRequest(r)...),
			oteltrace.WithAttributes(attribute.String("user.id", tokenUserID)),
			oteltrace.WithAttributes(attribute.String("flight.id", flightID)),
		}

		if route != "" {
			rAttr := semconv.HTTPRoute(route)
			opts = append(opts, oteltrace.WithAttributes(rAttr))
		}

		tracer := otel.GetTracerProvider().Tracer(
			tracerName,
		)

		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		req.Request = req.Request.WithContext(ctx)

		chain.ProcessFilter(req, resp)

		status := resp.StatusCode()
		span.SetStatus(HTTPServerStatus(status))
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
	}
}
