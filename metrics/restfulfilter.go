// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"strconv"
	"time"

	"github.com/emicklei/go-restful/v3"
)

func Filter() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		dateStart := time.Now()
		Namespace := req.PathParameter(namespacePathParameter)
		chain.ProcessFilter(req, resp)
		reqSelectedRoot := req.SelectedRoute()
		if reqSelectedRoot != nil {
			httpMetrics.With(map[string]string{
				labelNamespace:    Namespace,
				labelPath:         reqSelectedRoot.Path(),
				labelMethod:       reqSelectedRoot.Method(),
				labelResponseCode: strconv.Itoa(resp.StatusCode()),
			}).Observe(time.Since(dateStart).Seconds())
		}
	}
}
