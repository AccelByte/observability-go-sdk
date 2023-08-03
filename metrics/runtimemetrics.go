// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"runtime/metrics"
	"time"
)

var (
	runtimeMetricsGaugeMap  map[string]GaugeMetric
	runtimeMetricsHistogram map[string]ObserverMetric
)

func sendRuntimeMetrics() {
	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:

			descs := metrics.All()

			samples := make([]metrics.Sample, len(descs))
			for i := range samples {
				samples[i].Name = descs[i].Name
			}

			metrics.Read(samples)

			for _, sample := range samples {
				name, value := sample.Name, sample.Value

				switch value.Kind() {
				case metrics.KindUint64:
					runtimeMetricsGaugeMap[name].Set(float64(value.Uint64()))

				case metrics.KindFloat64:
					runtimeMetricsGaugeMap[name].Set(value.Float64())

				case metrics.KindFloat64Histogram:
					for _, i := range value.Float64Histogram().Buckets {
						runtimeMetricsHistogram[name].Observe(i)
					}
				}
			}

		case <-quit:
			return
		}
	}
}
