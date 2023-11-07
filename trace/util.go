package trace

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func HTTPServerRequest(req *http.Request) []attribute.KeyValue {

	n := 4 // Method, scheme, proto, and host name.
	var host string
	var p int
	host, p = splitHostPort(req.Host)
	hostPort := requiredHTTPPort(req.TLS != nil, p)
	if hostPort > 0 {
		n++
	}

	peer, peerPort := splitHostPort(req.RemoteAddr)
	if peer != "" {
		n++
		if peerPort > 0 {
			n++
		}
	}

	useragent := req.UserAgent()
	if useragent != "" {
		n++
	}

	clientIP := serverClientIP(req.Header.Get("X-Forwarded-For"))
	if clientIP != "" {
		n++
	}

	userID, _, hasUserID := req.BasicAuth()
	if hasUserID {
		n++
	}

	attrs := make([]attribute.KeyValue, 0, n)

	attrs = append(attrs, HTTPMethodKey.String(req.Method))
	attrs = append(attrs, scheme(req.TLS != nil))
	attrs = append(attrs, flavor(req.Proto))
	attrs = append(attrs, NetHostNameKey.String(host))

	if hostPort > 0 {
		attrs = append(attrs, NetHostPortKey.Int(hostPort))
	}

	if peer != "" {
		// The Go HTTP server sets RemoteAddr to "IP:port", this will not be a
		// file-path that would be interpreted with a sock family.
		attrs = append(attrs, NetSockPeerAddrKey.String(peer))
		if peerPort > 0 {
			attrs = append(attrs, NetSockPeerPortKey.Int(peerPort))
		}
	}

	if useragent != "" {
		attrs = append(attrs, HTTPUserAgentKey.String(useragent))
	}

	if hasUserID {
		attrs = append(attrs, EnduserIDKey.String(userID))
	}

	if clientIP != "" {
		attrs = append(attrs, HTTPClientIPKey.String(clientIP))
	}

	return attrs
}

func HTTPServerStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}

func scheme(https bool) attribute.KeyValue {
	if https {
		return HTTPSchemeHTTPS
	}
	return HTTPSchemeHTTP
}

func flavor(proto string) attribute.KeyValue {
	switch proto {
	case "HTTP/1.0":
		return HTTPFlavorKey.String("1.0")
	case "HTTP/1.1":
		return HTTPFlavorKey.String("1.1")
	case "HTTP/2":
		return HTTPFlavorKey.String("2.0")
	case "HTTP/3":
		return HTTPFlavorKey.String("3.0")
	default:
		return HTTPFlavorKey.String(proto)
	}
}

func serverClientIP(xForwardedFor string) string {
	if idx := strings.Index(xForwardedFor, ","); idx >= 0 {
		xForwardedFor = xForwardedFor[:idx]
	}
	return xForwardedFor
}

func splitHostPort(hostport string) (host string, port int) {
	port = -1

	if strings.HasPrefix(hostport, "[") {
		addrEnd := strings.LastIndex(hostport, "]")
		if addrEnd < 0 {
			// Invalid hostport.
			return
		}
		if i := strings.LastIndex(hostport[addrEnd:], ":"); i < 0 {
			host = hostport[1:addrEnd]
			return
		}
	} else {
		if i := strings.LastIndex(hostport, ":"); i < 0 {
			host = hostport
			return
		}
	}

	host, pStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return
	}

	p, err := strconv.ParseUint(pStr, 10, 16)
	if err != nil {
		return
	}
	return host, int(p)
}

func requiredHTTPPort(https bool, port int) int { // nolint:revive
	if https {
		if port > 0 && port != 443 {
			return port
		}
	} else {
		if port > 0 && port != 80 {
			return port
		}
	}
	return -1
}
