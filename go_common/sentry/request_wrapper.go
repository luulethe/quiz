package sentry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

type HTTPHandler func(req *http.Request) (*http.Response, error)

func WithSentryBreadcrumb(ctx context.Context, do HTTPHandler, alertLatency uint64) HTTPHandler {
	return func(req *http.Request) (*http.Response, error) {
		start := time.Now()
		res, err := do(req)
		latency := uint64(float64(time.Since(start)) / float64(time.Millisecond))

		requestURI := fmt.Sprintf("%s %s", req.Method, req.URL.String())
		var latencyErr error
		if latency > alertLatency {
			latencyErr = fmt.Errorf("high latency|requestURI:%s|latency:%d", requestURI, latency)
		}
		fingerprint := fmt.Sprintf("%s %s%s", req.Method, req.URL.Host, req.URL.Path)

		breadcrumb := &sentry.Breadcrumb{
			Category: "request",
			Message:  requestURI,
			Level:    sentry.LevelInfo,
			Data: map[string]interface{}{
				"latency": latency,
			},
		}
		hub := ctx.Value(ContextHubKey)
		sHub, sHubOK := hub.(*sentry.Hub)
		if sHubOK {
			sHub.AddBreadcrumb(breadcrumb, nil)
			if latencyErr != nil {
				sHub.WithScope(func(scope *sentry.Scope) {
					event := getEvent(latencyErr, 1)
					event.Fingerprint = []string{"high_latency", fingerprint}
					sHub.CaptureEvent(event)
				})
			}
		} else {
			sentry.AddBreadcrumb(breadcrumb)
			if latencyErr != nil {
				event := getEvent(latencyErr, 0)
				event.Fingerprint = []string{"high_latency", fingerprint}
				sentry.CaptureEvent(event)
			}
		}
		return res, err
	}
}
