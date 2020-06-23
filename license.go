package main

import (
	"net/http"
	"strings"
)

const (
	licenseKey = "license"
	licenseLen = len(licenseKey)
	licenseUrl = "https://github.com/Shivam010/bypass-cors/blob/master/LICENSE"
)

func License(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(licenseKey, licenseUrl)
		r.Header.Add(licenseKey, licenseUrl)
		path := strings.ToLower(r.URL.Path)
		// redirect to license if URL path is "/license*"
		if len(path) > licenseLen && path[1:licenseLen+1] == licenseKey {
			http.Redirect(w, r, licenseUrl, http.StatusPermanentRedirect)
			return
		}
		h.ServeHTTP(w, r)
		w.Header().Add(licenseKey, licenseUrl)
		r.Header.Add(licenseKey, licenseUrl)
	})
}
