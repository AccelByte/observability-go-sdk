// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeDebugHandlerFunc(t *testing.T) {
	successReq := httptest.NewRequest("GET", "/admin/internal/runtimedebug", nil)
	successReqWithQuery := httptest.NewRequest("GET", "/admin/internal/runtimedebug?profile=heap", nil)
	failedReqWithInvalidQuery := httptest.NewRequest("GET", "/admin/internal/runtimedebug?profile=unknown", nil)

	testCases := []struct {
		name             string
		req              *restful.Request
		expectedProfile  string
		expectedRespCode int
	}{
		{
			name:             "success",
			req:              restful.NewRequest(successReq),
			expectedProfile:  defaultProfile,
			expectedRespCode: http.StatusOK,
		},
		{
			name:             "success with specified profile query param",
			req:              restful.NewRequest(successReqWithQuery),
			expectedProfile:  "heap",
			expectedRespCode: http.StatusOK,
		},
		{
			name:             "failed with invalid profile query param",
			req:              restful.NewRequest(failedReqWithInvalidQuery),
			expectedRespCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			runtimeDebugHandlerFunc(testCase.req, restful.NewResponse(w))
			result := w.Result()
			defer result.Body.Close()
			respBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedRespCode, result.StatusCode)
			if testCase.expectedProfile != "" {
				assert.True(t, strings.HasPrefix(string(respBody), fmt.Sprintf("%s profile", testCase.expectedProfile)))
			} else {
				assert.Contains(t, string(respBody), strings.Join(getProfileNames(), ","))
			}
		})
	}
}
