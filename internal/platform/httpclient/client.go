package httpclient

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// New returns a resty client configured with sensible defaults.
func New(timeoutSeconds, maxRetries int) *resty.Client {
	client := resty.New()
	client.SetTimeout(time.Duration(timeoutSeconds) * time.Second)
	client.SetRetryCount(maxRetries)
	return client
}
