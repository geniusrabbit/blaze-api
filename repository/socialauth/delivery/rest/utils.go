package rest

import (
	"net/url"
	"strings"
)

func urlSetQueryParams(sUrl string, params map[string]string) string {
	if len(params) == 0 {
		return sUrl
	}

	setted := make(map[string]bool, len(params))

	// Replace pattern params in URL by patter `{paramName}`
	for k, v := range params {
		if strings.Contains(sUrl, `{`+k+`}`) {
			sUrl = strings.Replace(sUrl, `{`+k+`}`, v, -1)
			setted[k] = true
		}
	}

	query := url.Values{}
	baseURL := strings.SplitN(sUrl, "?", 2)
	if len(baseURL) == 2 {
		query, _ = url.ParseQuery(baseURL[1])
	}

	for k, v := range params {
		if _, ok := setted[k]; !ok {
			query.Set(k, v)
		}
	}
	return baseURL[0] + "?" + query.Encode()
}
