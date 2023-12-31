// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

const (
	metricsNameFormat  = "ab.%s_%s" // ab is the namespace and the next two placeholder is for service name and metrics name
	metricsNameHTTP    = "request_http"
	genericServiceName = "service"

	defaultNamespacePathParameter = "namespace"

	labelNamespace    = "namespace"
	labelPath         = "path"
	labelMethod       = "method"
	labelResponseCode = "response_code"
)
