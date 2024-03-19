package requestid

import (
	"net/http"

	"github.com/google/uuid"
)

const defaultHeaderName = "X-Request-ID"

type Option func(*config)

type config struct {
	externalIDHeaders []string
	headerName        string
	queryIDGenerator  func(r *http.Request) string
	externalIDGetter  func(r *http.Request) string
}

func defaultIDGetter(headers ...string) func(r *http.Request) string {
	return func(r *http.Request) string {
		for _, header := range headers {
			if id := r.Header.Get(header); id != "" {
				return id
			}
		}
		return ""
	}
}

func DefaultRequestIDGenerator(r *http.Request) string {
	return uuid.NewString()
}

// WithExternalIDHeaders sets headers to get external id from
func WithExternalIDHeaders(headers ...string) Option {
	return func(c *config) {
		c.externalIDHeaders = headers
	}
}

// WithHeaderName sets header name to put queryid to
func WithHeaderName(name string) Option {
	return func(c *config) {
		c.headerName = name
	}
}

// WithQueryIDGenerator sets queryid generator
func WithQueryIDGenerator(generator func(r *http.Request) string) Option {
	return func(c *config) {
		c.queryIDGenerator = generator
	}
}

// WithExternalIDGetter sets external id getter
func WithExternalIDGetter(getter func(r *http.Request) string) Option {
	return func(c *config) {
		c.externalIDGetter = getter
	}
}

// WithCloudflareRequestID sets external id getter to get request id from cloudflare headers
func WithCloudflareRequestID() Option {
	return WithExternalIDGetter(cfRequestID)
}

func (cfg *config) getRequestIDGenerator() func(r *http.Request) string {
	if cfg.queryIDGenerator == nil {
		return DefaultRequestIDGenerator
	}
	return cfg.queryIDGenerator
}

func (cfg *config) getExternalIDGetter() func(r *http.Request) string {
	if cfg.externalIDGetter == nil {
		return defaultIDGetter(cfg.externalIDHeaders...)
	}
	return cfg.externalIDGetter
}

func (cfg *config) getHeaderName() string {
	if cfg.headerName == "" {
		return defaultHeaderName
	}
	return cfg.headerName
}
