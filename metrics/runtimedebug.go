// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package metrics

import (
	"bytes"
	"errors"
	"net/http"
	"runtime/pprof"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/sirupsen/logrus"
)

const (
	defaultProfile = "goroutine"
)

func runtimeDebugHandlerFunc(req *restful.Request, res *restful.Response) {
	profile := req.QueryParameter("profile")
	if profile == "" {
		profile = defaultProfile
	}

	if !isProfileValid(profile) {
		errInvalidProfileQueryParam := errors.New("invalid profile query param. " +
			"Available profile: " + strings.Join(getProfileNames(), ","))
		err := res.WriteErrorString(http.StatusBadRequest, errInvalidProfileQueryParam.Error())
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	buf := new(bytes.Buffer)
	err := pprof.Lookup(profile).WriteTo(buf, 1)
	if err != nil {
		logrus.WithField("pprof", profile).Error("Unable to return runtime profiling")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	logrus.WithField("pprof", profile).Println(buf.String())

	res.WriteHeader(http.StatusOK)
	_, err = res.Write(buf.Bytes())
	if err != nil {
		logrus.WithField("pprof", profile).Error("Unable to write runtime profiling response")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func isProfileValid(p string) bool {
	for _, profile := range pprof.Profiles() {
		if profile.Name() == p {
			return true
		}
	}
	return false
}

func getProfileNames() []string {
	names := make([]string, 0)
	for _, p := range pprof.Profiles() {
		names = append(names, p.Name())
	}

	return names
}
