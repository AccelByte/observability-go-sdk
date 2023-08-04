// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

const (
	metricsNameFormat = "ab_%s_%s"
	metricsNameHTTP   = "total_request_http"

	defaultNamespacePathParameter = "namespace"

	labelNamespace    = "game_namespace"
	labelPath         = "path"
	labelMethod       = "method"
	labelResponseCode = "response_code"
)
