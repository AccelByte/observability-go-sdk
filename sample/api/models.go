// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import "time"

type Ban struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ExpiredAt time.Time `json:"expiredAt"`
}

type AddBanRequest struct {
	Name      string    `json:"name"`
	ExpiredAt time.Time `json:"expiredAt"`
}
