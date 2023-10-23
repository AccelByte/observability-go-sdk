// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package trace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getFuncName(t *testing.T) {
	assert.Equal(t, "trace.Test_getFuncName", getFuncNameInStack(0))

	func() {
		assert.Equal(t, "trace.Test_getFuncName.func1", getFuncNameInStack(0))
		assert.Equal(t, "trace.Test_getFuncName", getCallingFuncName())
	}()
}
