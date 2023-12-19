// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

import (
	"context"
	"log"
	"time"

	"github.com/gremlinflat/observability-go-sdk/metrics"
	"github.com/gremlinflat/observability-go-sdk/sample/api"
	"github.com/gremlinflat/observability-go-sdk/trace"
)

const (
	BASE_PATH               = "/sampleservice"
	OTEL_COLLECTOR_ENDPOINT = "127.0.0.1:5555"
	OTEL_COLLECTOR_TIMEOUT  = time.Second * 5
)

func main() {
	totalSession := metrics.CounterVec(
		"ab_session_total_session",
		"The total number of available session",
		[]string{"namespace", "matchpool"},
	)

	metrics.SetProvider(metrics.NewPrometheusProvider(metrics.PrometheusProviderOpts{
		DisableGoCollector:      true, // disable default go collector
		DisableProcessCollector: true, // disable default process collector
	}))

	metrics.Initialize("test_service", metrics.BuildInfo{
		RevisionID:         "a41133",
		BuildDate:          time.Now().String(),
		Version:            "1.1.0",
		GitHash:            "a41133",
		RoleSeedingVersion: "1.0.0",
	}, &metrics.Opts{
		EnableRuntimeMetrics: false, // set this to true or leave it empty to enable go runtime metrics
	})

	trace.Initialize("test_service", "observability-go-sdk")

	clean, err := trace.SetUpTracer(context.Background(), OTEL_COLLECTOR_ENDPOINT, OTEL_COLLECTOR_TIMEOUT)
	if err != nil {
		log.Fatalf("error set up otel tracer : %v", err.Error())
	}
	defer clean()

	go sendCustomPeriodically(totalSession)
	api.InitWebService(BASE_PATH).Serve()
}

func sendCustomPeriodically(totalSession metrics.CounterVecMetric) {
	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			totalSession.With(map[string]string{"namespace": "test", "matchpool": "asdf"}).Add(float64(5))

		case <-quit:
			return
		}
	}
}
