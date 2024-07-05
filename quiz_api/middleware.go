package quiz_api

import (
	"context"
	"fmt"

	basesentry "github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/luulethe/quiz/go_common/log"
	"github.com/luulethe/quiz/go_common/sentry"
	"github.com/luulethe/quiz/quiz_lib/manager"
	pb "github.com/luulethe/quiz/quiz_lib/pb/gen"

	"time"
)

type HandlerFunc func(context.Context, *manager.Dependency, *pb.RequestData, *pb.ResponseData) error
type GRPCMiddleware func(handlerFunc HandlerFunc) HandlerFunc

type MiddlewareGroup struct {
	Middlewares []GRPCMiddleware
}

func (w *MiddlewareGroup) Wrap(handler HandlerFunc) HandlerFunc {
	for _, middleware := range w.Middlewares {
		handler = middleware(handler)
	}
	return handler
}

func LogMiddleware(handlerFunc HandlerFunc) HandlerFunc {
	return func(context context.Context, dep *manager.Dependency, request *pb.RequestData, response *pb.ResponseData) (err error) {
		startTime := time.Now()
		context = log.WithFields(context, log.Fields{"rid": uuid.New().String()})
		err = handlerFunc(context, dep, request, response)
		latency := time.Since(startTime)
		log.Infof(context,
			"Access Log: %s|latency=%v",
			request.Command.String(),
			latency,
		)
		return
	}
}

func sentryRecover(ctx context.Context) {
	if err := recover(); err != nil {
		e, ok := err.(error)
		if !ok {
			e = fmt.Errorf("error: %v", e)
		}
		sentry.CaptureError(ctx, e, 0)
		panic(err)
	}
}

func SentryMiddleware(handlerFunc HandlerFunc) HandlerFunc {
	return func(ctx context.Context, dep *manager.Dependency, request *pb.RequestData, response *pb.ResponseData) (err error) {
		hub := basesentry.CurrentHub().Clone()
		ctx = context.WithValue(ctx, sentry.ContextHubKey, hub) //nolint
		defer sentryRecover(ctx)
		err = handlerFunc(ctx, dep, request, response)
		return
	}
}

func MetricsMiddleware(handlerFunc HandlerFunc) HandlerFunc {
	return func(ctx context.Context, dep *manager.Dependency, request *pb.RequestData, response *pb.ResponseData) (err error) {
		if dep.Stats == nil {
			return handlerFunc(ctx, dep, request, response)
		}
		start := time.Now()
		err = handlerFunc(ctx, dep, request, response)
		latency := float64(time.Since(start)) / float64(time.Millisecond)
		result := "Success"
		if err != nil {
			result = "Error"
		}

		dep.Stats.ReportCount(1, request.Command.String(), result)
		dep.Stats.ReportLatency(latency, request.Command.String(), result)
		return
	}
}

var middlewareGroup = MiddlewareGroup{
	Middlewares: []GRPCMiddleware{
		LogMiddleware,
		SentryMiddleware,
		MetricsMiddleware,
	},
}
