// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

import (
	"time"

	"github.com/AccelByte/observability-go-sdk/metrics"
	"github.com/AccelByte/observability-go-sdk/sample/api"
)

const BASE_PATH = "/sampleservice"

func main() {
	totalSession := metrics.CounterVec(
		"ab_session_total_session",
		"The total number of available session",
		[]string{"namespace", "matchpool"},
	)

	metrics.Initialize("test_service", nil)

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
