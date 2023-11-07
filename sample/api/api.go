// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	auth "github.com/AccelByte/go-restful-plugins/v4/pkg/auth/iam"
	logger "github.com/AccelByte/go-restful-plugins/v4/pkg/logger/log"
	"github.com/AccelByte/iam-go-sdk"
	"github.com/AccelByte/observability-go-sdk/metrics"
	"github.com/AccelByte/observability-go-sdk/trace"
	"github.com/emicklei/go-restful/v3"
)

func InitWebService(basePath string) *WebService {
	iamClient := iam.NewMockClient()

	authFilterOptions := auth.FilterInitializationOptionsFromEnv()
	authFilter := auth.NewFilterWithOptions(iamClient, authFilterOptions)

	bansDAO := NewBansDAO()
	h := newHandlers(bansDAO)

	serviceContainer := newServiceContainer(basePath, authFilter, h)
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", "8080"))
	if err != nil {
		log.Fatalf("unable to listen on port 8080: %s", err.Error())
	}

	return &WebService{
		serviceContainer: serviceContainer,
		listener:         listener,
	}
}

func newServiceContainer(basePath string, authFilter *auth.Filter, h *handlers) *restful.Container {
	container := restful.NewContainer()
	container.Filter(logger.AccessLog)

	// register filter to send http metrics
	container.Filter(metrics.RestfulFilter())

	// register to add userid and flightid in span attributes
	container.Filter(trace.InstrumentCommonAttributes("test-service", map[string]string{
		"/sampleservice/bans/{banId}": http.MethodDelete,
	}))

	// register metrics and runtime debug routes
	container.Add(metrics.
		NewWebService(basePath).
		MetricsRoute(metrics.DefaultMetricsHandler).
		RuntimeDebugRoute().
		WebService())

	container.Add(bansService(basePath, h))

	// register mock API that will return random response code with random response time
	container.Add(addMockAPI(basePath))

	return container
}

func bansService(basePath string, h *handlers) *restful.WebService {
	webService := new(restful.WebService)
	webService.Path(basePath + "/bans")
	webService.Filter(logger.AccessLog)

	// POST {basePath}/bans
	webService.Route(webService.
		POST("").
		To(h.AddBan))

	// GET {basePath}/bans/{banId}
	webService.Route(webService.
		GET("/{banId}").
		To(h.GetBan))

	// DELETE {basePath}/bans/{banId}
	webService.Route(webService.
		DELETE("/{banId}").
		To(h.DeleteBan))

	return webService
}

type WebService struct {
	serviceContainer *restful.Container
	listener         net.Listener
}

func (w *WebService) Serve() {
	if err := http.Serve(w.listener, w.serviceContainer); err != nil {
		log.Fatal("unable to serve: ", err)
	}
}

func addMockAPI(basePath string) *restful.WebService {
	mockApi := new(restful.WebService)

	mockApi.Path(basePath)
	mockApi.Route(mockApi.GET("mock-api").To(func(request *restful.Request, response *restful.Response) {

		// sleep 0-2500ms to simulate slow response time
		time.Sleep(time.Duration(int64(rand.Float64()*2500)) * time.Millisecond)

		// randomize response code
		responseCodeList := []int{http.StatusOK, http.StatusBadRequest, http.StatusUnauthorized, http.StatusBadGateway, http.StatusInternalServerError}

		response.WriteHeaderAndJson(responseCodeList[rand.Intn(len(responseCodeList))], nil, restful.MIME_JSON)
	}))

	return mockApi
}
