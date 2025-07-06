package zitadelproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Middleware returns an http.Handler that forwards matching requests
// to the Zitadel instance while adding required headers.
func Middleware(target *url.URL, loginClient, publicHost, instanceHost string, paths []string) func(http.Handler) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range paths {
				if strings.HasPrefix(r.URL.Path, p) {
					r.Header.Set("x-zitadel-login-client", loginClient)
					r.Header.Set("x-zitadel-public-host", publicHost)
					r.Header.Set("x-zitadel-instance-host", instanceHost)
					proxy.ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
