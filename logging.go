package cmlclient

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func getURL(req *http.Request) string {
	url := "unknown"
	if req.URL != nil {
		url = req.URL.String()
	}
	return url
}

func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			res, err := next(req)
			dur := time.Since(start)
			if err != nil {
				logger.Error("HTTP", "method", req.Method, "url", req.URL, "err", err, "duration", dur)
				return nil, err
			}
			if res == nil {
				return nil, err
			}
			logger.Info("HTTP", "method", req.Method, "url", getURL(req), "code", res.StatusCode, "duration", dur)
			return res, nil
		}
	}
}

func LogRequestBodyMiddleware(logger *slog.Logger) Middleware {
	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			// no body, no logging
			if req.Body == nil {
				return next(req)
			}
			// read the body and log it
			body, err := io.ReadAll(req.Body)
			if err != nil {
				logger.Info("Error reading request body", "err", err)
				return nil, err
			}

			// Log the body
			logger.Info("Request body", "url", getURL(req), "eody", body)

			// create a new one / reset the body
			req.Body = io.NopCloser(bytes.NewReader(body))
			return next(req)
		}
	}
}
