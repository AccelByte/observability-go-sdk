// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/gremlinflat/observability-go-sdk/metrics"
	"github.com/gremlinflat/observability-go-sdk/trace"
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

func (b *BansDAO) AddBan(ctx context.Context, ban Ban) error {
	ctx, span := trace.NewChildSpan(ctx, "BansDAO.AddBan")
	defer span.End()

	if ban.ID == "" {
		return errors.New("ID can't be empty")
	}

	addBanMetrics := b.dbMetrics.NewCall("add_ban")
	defer addBanMetrics.CallEnded()
	trace.LogTraceInfo(ctx, "adding ban to DB", nil)

	{ // simulate response time and timeout error
		sleepTime := int64(rand.Float64() * 3000)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		if sleepTime > 2000 {
			err := errors.New("request to DB failed")
			addBanMetrics.Error()
			trace.LogTraceError(ctx, err, err.Error())
			return err
		}
	}
	trace.LogTraceInfo(ctx, "ban added to DB", nil)

	b.inMem = map[string]Ban{ban.ID: ban}

	return nil
}

func (b *BansDAO) GetBan(ctx context.Context, banID string) (Ban, error) {
	ctx, span := trace.NewAutoNamedChildSpan(ctx)
	defer span.End()

	getBanMetrics := b.dbMetrics.NewCall("get_ban")
	defer getBanMetrics.CallEnded()

	trace.LogTraceInfo(ctx, "getting ban from DB", nil)

	{ // simulate response time and timeout error
		sleepTime := int64(rand.Float64() * 3000)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		if sleepTime > 2000 {
			err := errors.New("request to DB failed")
			getBanMetrics.Error()
			trace.LogTraceError(ctx, err, err.Error())
			return Ban{}, err
		}
	}

	ban, exist := b.inMem[banID]
	if !exist {
		return Ban{}, notFoundError
	}

	return ban, nil
}

func (b *BansDAO) DeleteBan(ctx context.Context, banID string) error {
	if banID == "" {
		return errors.New("ID can't be empty")
	}

	addBanMetrics := b.dbMetrics.NewCall("delete_ban")
	defer addBanMetrics.CallEnded()

	{ // simulate response time and timeout error
		sleepTime := int64(rand.Float64() * 3000)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		if sleepTime > 2000 {
			err := errors.New("request to DB failed")
			addBanMetrics.Error()
			trace.LogTraceError(ctx, err, err.Error())
			return err
		}
	}

	delete(b.inMem, banID)

	return nil
}
