package cmlclient

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"
)

type RetryPolicy struct {
	MaxAttempts int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

func (r RetryPolicy) backoff(i int) time.Duration {
	d := r.BaseBackoff << i
	if r.MaxBackoff > 0 && d > r.MaxBackoff {
		return r.MaxBackoff
	}
	return d
}

func defaultRetry(p RetryPolicy) RetryPolicy {
	if p.MaxAttempts == 0 {
		p.MaxAttempts = 3
	}
	if p.BaseBackoff == 0 {
		p.BaseBackoff = 200 * time.Millisecond
	}
	if p.MaxBackoff == 0 {
		p.MaxBackoff = 2 * time.Second
	}
	return p
}

func RetryMiddleware(policy RetryPolicy) Middleware {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 3
	}
	if policy.BaseBackoff <= 0 {
		policy.BaseBackoff = 200 * time.Millisecond
	}
	if policy.MaxBackoff <= 0 {
		policy.MaxBackoff = 2 * time.Second
	}

	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			ctx := req.Context()
			var bodyBytes []byte
			if req.Body != nil {
				b, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				bodyBytes = b
			}

			for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
				r2 := req.Clone(ctx)
				if bodyBytes != nil {
					r2.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}

				res, err := next(r2)
				if err != nil {
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}
					if attempt+1 < policy.MaxAttempts {
						time.Sleep(policy.backoff(attempt))
						continue
					}
					return nil, err
				}

				switch res.StatusCode {
				case http.StatusTooManyRequests,
					http.StatusBadGateway,
					http.StatusServiceUnavailable,
					http.StatusGatewayTimeout:
					if attempt+1 < policy.MaxAttempts {
						_ = drainAndClose(res.Body)
						time.Sleep(policy.backoff(attempt))
						continue
					}
				}
				return res, nil
			}
			return nil, errors.New("retry attempts exhausted")
		}
	}
}
