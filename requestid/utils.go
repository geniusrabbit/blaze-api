package requestid

import "net/http"

func cfRequestID(r *http.Request) string {
	cfReqID := r.Header.Get("Cf-Request-Id")
	cfRay := r.Header.Get("Cf-Ray")
	if cfReqID == "" || cfRay == "" {
		return ""
	}
	return cfReqID + "/" + cfRay
}
