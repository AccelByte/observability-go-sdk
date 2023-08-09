// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"net/http"

	auth "github.com/AccelByte/go-restful-plugins/v4/pkg/auth/iam"
	"github.com/AccelByte/go-restful-plugins/v4/pkg/logger/log"
	"github.com/AccelByte/iam-go-sdk"
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
	webService.Filter(log.AccessLog)

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

func (s *serviceBuilder) RuntimeDebugRoute(authFilter *auth.Filter) *serviceBuilder {
	defaultPermission := &iam.Permission{
		Resource: "ADMIN:SYSTEM:DEBUG",
		Action:   2,
	}

	s.webService.Route(s.webService.
		GET("/runtimedebug").
		To(runtimeDebugHandlerFunc).
		Filter(
			authFilter.Auth(
				auth.WithPermission(defaultPermission),
			),
		))

	return s
}

// WebService is the final method to return the WebService.
func (s *serviceBuilder) WebService() *restful.WebService {
	return s.webService
}
