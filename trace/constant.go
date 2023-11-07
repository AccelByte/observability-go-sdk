// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package trace

import "go.opentelemetry.io/otel/attribute"

const (
	UserIDAttribute = "LogUserId"

	FlightID = "x-flight-id"

	HTTPMethodKey     = attribute.Key("http.method")
	HTTPStatusCodeKey = attribute.Key("http.status_code")
	HTTPFlavorKey     = attribute.Key("http.flavor")
	HTTPUserAgentKey  = attribute.Key("http.user_agent")
	HTTPSchemeKey     = attribute.Key("http.scheme")
	HTTPClientIPKey   = attribute.Key("http.client_ip")

	EnduserIDKey = attribute.Key("enduser.id")

	NetHostNameKey     = attribute.Key("net.host.name")
	NetHostPortKey     = attribute.Key("net.host.port")
	NetSockPeerAddrKey = attribute.Key("net.sock.peer.addr")
	NetSockPeerPortKey = attribute.Key("net.sock.peer.port")
)

var (
	HTTPSchemeHTTP  = HTTPSchemeKey.String("http")
	HTTPSchemeHTTPS = HTTPSchemeKey.String("https")
)
