// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	dbCallLabelAction     = "action"
	dbCallLabelResult     = "result"
	dbMetricLabelInstance = "instance"

	dbCallResultSuccess = "success"
	dbCallResultError   = "error"
)

type DBMetrics struct {
	dbName          string
	metricsProvider Provider
	latencyMetrics  ObserverVecMetric
}

type PostgreDBMetrics struct {
	MaxOpenConnections ObserverVecMetric
	OpenConnections    ObserverVecMetric
	InUse              ObserverVecMetric
	Idle               ObserverVecMetric
	WaitCount          ObserverVecMetric
	WaitDuration       ObserverVecMetric
	MaxIdleClosed      ObserverVecMetric
	MaxIdleTimeClosed  ObserverVecMetric
	MaxLifetimeClosed  ObserverVecMetric
}

// NewDBMetrics returns new DB metrics.
func NewDBMetrics(metricsProvider Provider, dbName string, labels ...string) *DBMetrics {
	l := []string{dbCallLabelAction, dbCallLabelResult}
	if len(labels) > 0 {
		l = append(l, labels...)
	}
	latencyMetrics := metricsProvider.NewHistogram(generateDBMetricsName(dbName),
		fmt.Sprintf("Latency of %s in seconds", dbName), prometheus.DefBuckets, l...)
	return &DBMetrics{metricsProvider: metricsProvider, latencyMetrics: latencyMetrics}
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
	latencyMetrics ObserverVecMetric
}

func (e *dbCallMetrics) start() {
	e.startTime = time.Now().UTC()
}

// WithLabel attaches labels to the metrics.
func (e *dbCallMetrics) WithLabel(labels map[string]string) *dbCallMetrics {
	for k, v := range labels {
		e.labelsMap[k] = v
	}
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
	e.labelsMap[dbCallLabelResult] = getResultLabelValue(e.isError)
	e.latencyMetrics.With(e.labelsMap).Observe(e.elapsed().Seconds())
}

func getResultLabelValue(isError bool) string {
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

// NewPostgreDBMetrics returns new PostgreDBMetrics.
func NewPostgreDBMetrics(metricsProvider Provider, dbName string) *PostgreDBMetrics {
	l := []string{dbMetricLabelInstance}
	return &PostgreDBMetrics{
		MaxOpenConnections: metricsProvider.NewHistogram("postgres_db_stat_max_open_connections",
			fmt.Sprintf("Maximum open connections on %s", dbName), prometheus.DefBuckets, l...),
		OpenConnections: metricsProvider.NewHistogram("postgres_db_stat_open_connections",
			fmt.Sprintf("Established connections both in use and idle on %s", dbName), prometheus.DefBuckets, l...),
		InUse: metricsProvider.NewHistogram("postgres_db_stat_in_use",
			fmt.Sprintf("Connections currently in use on %s", dbName), prometheus.DefBuckets, l...),
		Idle: metricsProvider.NewHistogram("postgres_db_stat_idle",
			fmt.Sprintf("Idle connections on %s", dbName), prometheus.DefBuckets, l...),
		WaitCount: metricsProvider.NewHistogram("postgres_db_stat_wait_count",
			fmt.Sprintf("Total connections waited for on %s", dbName), prometheus.DefBuckets, l...),
		WaitDuration: metricsProvider.NewHistogram("postgres_db_stat_wait_duration",
			fmt.Sprintf("Total time blocked waiting for a new connection on %s", dbName), prometheus.DefBuckets, l...),
		MaxIdleClosed: metricsProvider.NewHistogram("postgres_db_stat_max_idle_closed",
			fmt.Sprintf("Total connections closed due to SetMaxIdleConns on %s", dbName), prometheus.DefBuckets, l...),
		MaxIdleTimeClosed: metricsProvider.NewHistogram("postgres_db_stat_max_idle_time_closed",
			fmt.Sprintf("Total connections closed due to SetConnMaxIdleTime on %s", dbName), prometheus.DefBuckets, l...),
		MaxLifetimeClosed: metricsProvider.NewHistogram("postgres_db_stat_max_lifetime_closed",
			fmt.Sprintf("Total connections closed due to SetConnMaxLifetime on %s", dbName), prometheus.DefBuckets, l...),
	}
}

func (dbMetric *PostgreDBMetrics) ObservePostgreDBMetric(dbType string, db *sql.DB) {
	dbStats := db.Stats()
	dbTypeLabel := map[string]string{dbMetricLabelInstance: dbType}
	dbMetric.MaxOpenConnections.With(dbTypeLabel).Observe(float64(dbStats.MaxOpenConnections))
	dbMetric.OpenConnections.With(dbTypeLabel).Observe(float64(dbStats.OpenConnections))
	dbMetric.InUse.With(dbTypeLabel).Observe(float64(dbStats.InUse))
	dbMetric.Idle.With(dbTypeLabel).Observe(float64(dbStats.Idle))
	dbMetric.WaitCount.With(dbTypeLabel).Observe(float64(dbStats.WaitCount))
	dbMetric.WaitDuration.With(dbTypeLabel).Observe(float64(dbStats.WaitDuration))
	dbMetric.MaxIdleClosed.With(dbTypeLabel).Observe(float64(dbStats.MaxIdleClosed))
	dbMetric.MaxIdleTimeClosed.With(dbTypeLabel).Observe(float64(dbStats.MaxIdleTimeClosed))
	dbMetric.MaxLifetimeClosed.With(dbTypeLabel).Observe(float64(dbStats.MaxLifetimeClosed))
}
