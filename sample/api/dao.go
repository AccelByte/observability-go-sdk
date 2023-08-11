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

var (
	notFoundError = errors.New("not found")
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
	defer addBanMetrics.CallEnded()

	{ // simulate response time and timeout error
		sleepTime := int64(rand.Float64() * 3000)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		if sleepTime > 2000 {
			addBanMetrics.Error()
			return errors.New("request to DB failed")
		}
	}

	b.inMem = map[string]Ban{ban.ID: ban}

	return nil
}

func (b *BansDAO) GetBan(banID string) (Ban, error) {
	getBanMetrics := b.dbMetrics.NewCall("get_ban")
	defer getBanMetrics.CallEnded()

	{ // simulate response time and timeout error
		sleepTime := int64(rand.Float64() * 3000)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		if sleepTime > 2000 {
			getBanMetrics.Error()
			return Ban{}, errors.New("request to DB failed")
		}
	}

	ban, exist := b.inMem[banID]
	if !exist {
		return Ban{}, notFoundError
	}

	return ban, nil
}
