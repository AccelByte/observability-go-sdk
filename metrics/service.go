// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

var (
	DefaultMetricsHandler = PrometheusHandler()
)

type serviceBuilder struct {
	basePath   string
	webService *restful.WebService
}

// NewWebService returns new metrics web service builder
func NewWebService(basePath string) *serviceBuilder {
	webService := new(restful.WebService)
	webService.Path(basePath + "/admin/internal")

	return &serviceBuilder{basePath: basePath, webService: webService}
}

func (s *serviceBuilder) MetricsRoute(metricsHandler http.Handler) *serviceBuilder {
	s.webService.Route(s.webService.
		GET("/metrics").
		To(func(req *restful.Request, res *restful.Response) {
			metricsHandler.ServeHTTP(res.ResponseWriter, req.Request)
		}))

	return s
}

func (s *serviceBuilder) RuntimeDebugRoute() *serviceBuilder {
	s.webService.Route(s.webService.
		GET(fmt.Sprintf("/debug/pprof/{%s}", pprofPathParam)).
		To(pprofHandlerFunc))

	return s
}

// WebService is the final method to return the WebService.
func (s *serviceBuilder) WebService() *restful.WebService {
	return s.webService
}
