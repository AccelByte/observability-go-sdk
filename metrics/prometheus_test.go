// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	testCases := []struct {
		testName       string
		input          string
		expectedOutput string
	}{
		{
			testName:       "replace special character",
			input:          "justice-test-service/othermetrics",
			expectedOutput: "justice_test_service_othermetrics",
		},
		{
			testName:       "not replace anything",
			input:          "justice_test_service_safe",
			expectedOutput: "justice_test_service_safe",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, sanitizeName(testCase.input))
		})
	}
}
