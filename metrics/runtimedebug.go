// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	netpprof "net/http/pprof"

	"github.com/emicklei/go-restful/v3"
)

const pprofPathParam = "pprof"

func pprofHandlerFunc(request *restful.Request, response *restful.Response) {
	pprof := request.PathParameter(pprofPathParam)
	switch pprof {
	case "profile":
		netpprof.Profile(response.ResponseWriter, request.Request)
	case "cmdline":
		netpprof.Cmdline(response.ResponseWriter, request.Request)
	case "symbol":
		netpprof.Symbol(response.ResponseWriter, request.Request)
	case "trace":
		netpprof.Trace(response.ResponseWriter, request.Request)
	default:
		netpprof.Handler(pprof).ServeHTTP(response.ResponseWriter, request.Request)
	}
}
