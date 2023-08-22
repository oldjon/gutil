package ip

import (
	"net"
	"net/http"
	"strings"
)

type FilterStrategy uint8

const (
	FirstIP FilterStrategy = iota
	LastPublicIP
)

type GetHTTPClientIPOptions struct {
	FilterStrategy       FilterStrategy
	SupportHeaderXRealIP bool // X-Real-IP is used by nginx ngx_http_realip_module
}

const (
	headerXForwardedFor = "X-Forwarded-For"
	headerXRealIP       = "X-Real-Ip"
)

// GetHTTPClientIP will extract ip from headers or just remoteaddr
//  code mainly from dts gateway code
// ref, https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
// todo, add support for Forwarded , https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded
func GetHTTPClientIP(r *http.Request, options GetHTTPClientIPOptions) string {
	ipList := r.Header.Get(headerXForwardedFor)

	ip := getForwardIP(ipList, options.FilterStrategy)

	if ip != "" {
		return ip
	}

	if options.SupportHeaderXRealIP {
		ipList = r.Header.Get(headerXRealIP)

		ip = getForwardIP(ipList, options.FilterStrategy)

		if ip != "" {
			return ip
		}
	}

	// we will always return remoteAddr if headers not matched
	return extractRemoteAddr(r.RemoteAddr)
}

func extractRemoteAddr(remoteAddr string) string {
	//  find last occurrence of : to split addr and port
	//  ipv6 address will be like [0000:0000:0000:0000:0000:0000:0000:0000]:12345
	lastInd := strings.LastIndex(remoteAddr, ":")
	if lastInd >= 0 {
		return remoteAddr[:lastInd]
	}

	// it would have some error in r.RemoteAddr?
	return ""
}

// getForwardIP is core function to get an ip for ipList(stored in http headers)
//   if can not find a valid ip , return an empty string
func getForwardIP(ipList string, filterStrategy FilterStrategy) string {
	ips := strings.Split(ipList, ",")

	if len(ips) == 0 {
		return ""
	}

	if filterStrategy == FirstIP {
		ip := strings.TrimSpace(ips[0])

		//verify this is a valid ip
		realIP := net.ParseIP(ip)
		if realIP == nil {
			return ""
		}

		return ip
	}

	if filterStrategy == LastPublicIP {
		// march from right to left until we get a public address
		// that will be the address right before our proxy.
		for i := len(ips) - 1; i >= 0; i-- {
			// header can contain spaces too, strip those out.
			ip := strings.TrimSpace(ips[i])

			// parse ip
			realIP := net.ParseIP(ip)
			if realIP == nil {
				continue
			}

			if realIP.IsGlobalUnicast() && IsPublicIP(realIP) {
				return ip
			}
		}

		return ""
	}

	// or default , should not go here
	return ""
}
