// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	dbCallLabelAction = "action"
	dbCallLabelResult = "result"

	dbCallResultSuccess = "success"
	dbCallResultError   = "error"
)

type DBMetrics struct {
	dbName          string
	metricsProvider Provider
	latencyMetrics  *ObserverVecMetric
}

// NewDBMetrics returns new DB metrics.
func NewDBMetrics(metricsProvider Provider, dbName string, labels ...string) *DBMetrics {
	l := []string{dbCallLabelAction, dbCallLabelResult}
	if len(labels) > 0 {
		l = append(l, labels...)
	}
	latencyMetrics := metricsProvider.NewHistogram(generateDBMetricsName(dbName),
		fmt.Sprintf("Latency of %s in seconds", dbName), prometheus.DefBuckets, l...)
	return &DBMetrics{metricsProvider: metricsProvider, latencyMetrics: &latencyMetrics}
}

func generateDBMetricsName(dbName string) string {
	return generateMetricsName(serviceName, fmt.Sprintf("%s_db_latency_seconds", dbName))
}

// NewCall returns a new DB call metrics and start it.
func (d *DBMetrics) NewCall(action string) *dbCallMetrics {
	dbCall := &dbCallMetrics{
		action:         action,
		startTime:      time.Time{},
		endTime:        time.Time{},
		labelsMap:      map[string]string{},
		latencyMetrics: d.latencyMetrics,
	}
	dbCall.start()
	return dbCall
}

type dbCallMetrics struct {
	action         string
	isError        bool
	startTime      time.Time
	endTime        time.Time
	labelsMap      map[string]string
	latencyMetrics *ObserverVecMetric
}

func (e *dbCallMetrics) start() {
	e.startTime = time.Now().UTC()
}

// WithLabel attaches labels to the metrics.
func (e *dbCallMetrics) WithLabel(labels map[string]string) *dbCallMetrics {
	e.labelsMap = labels
	return e
}

// Error is the function that you need to call after a DB call has failed/returned error.
// It marks the metrics 'result' label as error.
func (e *dbCallMetrics) Error() {
	e.isError = true
}

// CallEnded is the most important function that you need to call after a DB call.
// The metrics won't proceed without calling it.
func (e *dbCallMetrics) CallEnded() {
	e.endTime = time.Now().UTC()
	e.labelsMap[dbCallLabelAction] = e.action
	e.labelsMap[dbCallLabelResult] = getResultLabel(e.isError)
	latencyMetrics := *e.latencyMetrics
	latencyMetrics.With(e.labelsMap).Observe(e.elapsed().Seconds())
}

func getResultLabel(isError bool) string {
	if isError {
		return dbCallResultError
	}
	return dbCallResultSuccess
}

func (e *dbCallMetrics) elapsed() time.Duration {
	if e.startTime.IsZero() || e.endTime.IsZero() {
		return 0
	}
	return e.endTime.Sub(e.startTime)
}
