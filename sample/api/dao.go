// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import (
	"errors"
	"math/rand"
	"time"

	"github.com/AccelByte/observability-go-sdk/metrics"
)

type BansDAO struct {
	inMem     map[string]Ban
	dbMetrics *metrics.DBMetrics
}

func NewBansDAO() *BansDAO {
	bansDAOMetrics := metrics.NewDBMetrics(metrics.DefaultProvider, "bans_dao")

	return &BansDAO{dbMetrics: bansDAOMetrics}
}

func (b *BansDAO) AddBan(ban Ban) error {
	if ban.ID == "" {
		return errors.New("ID can't be empty")
	}

	addBanMetrics := b.dbMetrics.NewCall("add_ban")
	time.Sleep(time.Duration(int64(rand.Float64()*2500)) * time.Millisecond) // simulate response time
	b.inMem = map[string]Ban{ban.ID: ban}
	addBanMetrics.CallEnded()

	return nil
}

func (b *BansDAO) GetBan(banID string) (Ban, error) {
	getBanMetrics := b.dbMetrics.NewCall("get_ban")
	defer getBanMetrics.CallEnded()
	time.Sleep(time.Duration(int64(rand.Float64()*2500)) * time.Millisecond) // simulate response time
	ban, exist := b.inMem[banID]
	if !exist {
		return Ban{}, errors.New("not found")
	}

	return ban, nil
}
