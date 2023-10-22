// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package trace

import (
	"runtime"
	"strings"
)

func getFuncNameInStack(stackOffset int) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2+stackOffset, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	parts := strings.Split(frame.Function, "/")
	return parts[len(parts)-1]
}

func getCallingFuncName() string {
	return getFuncNameInStack(2)
}
