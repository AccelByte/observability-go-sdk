// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AccelByte/observability-go-sdk/trace"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
)

type handlers struct {
	bansDAO *BansDAO
}

func newHandlers(bansDAO *BansDAO) *handlers {
	return &handlers{bansDAO: bansDAO}
}

func (h *handlers) AddBan(req *restful.Request, res *restful.Response) {
	ctx, span := trace.NewRootSpan(req.Request.Context(), req.Request.RequestURI)
	defer span.End()

	var payload AddBanRequest
	err := req.ReadEntity(&payload)
	if err != nil {
		trace.LogTraceError(ctx, err, err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	ban := Ban{
		ID:        generateUUID(),
		Name:      payload.Name,
		ExpiredAt: payload.ExpiredAt,
	}

	err = h.bansDAO.AddBan(ctx, ban)
	if err != nil {
		trace.LogTraceError(ctx, err, err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeaderAndJson(http.StatusOK, ban, restful.MIME_JSON)
}

func (h *handlers) GetBan(req *restful.Request, res *restful.Response) {
	ctx, span := trace.NewRootSpan(req.Request.Context(), req.Request.RequestURI)
	defer span.End()

	banID := req.PathParameter("banId")
	ban, err := h.bansDAO.GetBan(ctx, banID)
	if err != nil {
		if err == notFoundError {
			err = res.WriteErrorString(http.StatusNotFound, fmt.Sprintf("ban with ID %s not found", banID))
			if err != nil {
				trace.LogTraceError(ctx, err, err.Error())
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
		trace.LogTraceError(ctx, err, err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = res.WriteHeaderAndJson(http.StatusOK, ban, restful.MIME_JSON)
	if err != nil {
		trace.LogTraceError(ctx, err, err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func generateUUID() string {
	id, _ := uuid.NewRandom()

	return strings.ReplaceAll(id.String(), "-", "")
}
