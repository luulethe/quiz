package util

import (
	"context"
	"runtime"

	"github.com/luulethe/quiz/go_common/log"
	"github.com/luulethe/quiz/go_common/sentry"
	"github.com/pkg/errors"
)

func ExitOnErr(ctx context.Context, err error) {
	if err != nil {
		log.Fatal(ctx, err.Error())
	}
}

var errAbort = errors.New("Abort Request Handling")

func RecoverOnAbortError(ctx context.Context) {
	if r := recover(); r != nil {
		if r != errAbort {
			if err, ok := r.(error); ok {
				sentry.CaptureError(ctx, err, 1)
			}
		}
	}
}

func AbortOnErrorFunc(ctx context.Context) func(error, string) {
	return func(err error, msg string) {
		if err != nil {
			errMsg := msg
			if errMsg == "" {
				errMsg = err.Error()
			}
			_, fn, line, _ := runtime.Caller(1)
			switch errors.Cause(err) {
			case context.Canceled:
			default:
				log.Errorff(ctx, "abortOnError|err:%s|fn:%s|line:%d", errMsg, fn, line)
				sentry.CaptureError(ctx, err, 2)
			}
			panic(errAbort)
		}
	}
}

func WithErrorCaptured(ctx context.Context, handler func() error, handlerName string) {
	err := handler()
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.Errorff(ctx, "WithErrorCaptured|handlerName:%s|err:%v|fn:%s|line:%d", handlerName, err, fn, line)
		sentry.CaptureError(ctx, err, 2)
	}
}
