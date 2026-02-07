// agent/auth/http.go
package auth

import (
	"net"
	"net/http"
	"time"
)

// sharedHTTPClient is a package-level HTTP client with optimized connection pooling.
// Using a shared client allows connection reuse across all auth operations.
var sharedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}
