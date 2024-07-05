package sentry

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Init(dsn string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		AttachStacktrace: true,
	})
}

func Flush() {
	sentry.Flush(2 * time.Second)
}

func Recover() {
	if err := recover(); err != nil {
		sentry.CurrentHub().Recover(err)
		Flush()
	}
}

type withStackError interface {
	StackTrace() errors.StackTrace
}

func getFrames(err withStackError) []sentry.Frame {
	var frames []sentry.Frame
	st := err.StackTrace()
	var pcs []uintptr
	for _, f := range st {
		pcs = append(pcs, uintptr(f))
	}
	callersFrames := runtime.CallersFrames(pcs)
	firstFrame := true
	for {
		callerFrame, more := callersFrames.Next()
		sentryFrame := sentry.NewFrame(callerFrame)
		if firstFrame {
			firstFrame = false
			if sentryFrame.Function == "init" {
				// error is generated from module init, not useful for stacktrace
				return nil
			}
		}

		frames = append([]sentry.Frame{sentryFrame}, frames...)

		if !more {
			break
		}
	}
	return frames
}

func filterStacktrace(st *sentry.Stacktrace, skip int) {
	if st != nil {
		frames := []sentry.Frame{}
		length := len(st.Frames)
		for index, frame := range st.Frames {
			if index == length-skip-1 {
				break
			}
			if frame.InApp {
				if strings.Contains(frame.Module, "vendor") || strings.Contains(frame.AbsPath, "vendor") {
					frame.InApp = false
				}
			}
			frames = append(frames, frame)
		}
		st.Frames = frames
	}
}

func getEvent(err error, callerSkip int) *sentry.Event {
	var frames []sentry.Frame
	var stacktrace *sentry.Stacktrace
	if errWithStack, ok := err.(withStackError); ok {
		frames = getFrames(errWithStack)
		if frames != nil {
			stacktrace = &sentry.Stacktrace{
				Frames: frames,
			}
			callerSkip = 0
		}
	}
	if stacktrace == nil {
		stacktrace = sentry.NewStacktrace()
	}
	filterStacktrace(stacktrace, callerSkip)

	event := sentry.NewEvent()
	event.Level = sentry.LevelError
	event.Message = err.Error()
	event.Exception = []sentry.Exception{{
		Type:       err.Error(),
		Value:      reflect.TypeOf(err).String(),
		Stacktrace: stacktrace,
	}}
	return event
}

func CaptureErrorWithContext(c *gin.Context, err error, callerSkip int) {
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			event := getEvent(err, callerSkip+1)
			hub.CaptureEvent(event)
		})
	} else {
		event := getEvent(err, callerSkip)
		sentry.CaptureEvent(event)
	}
}

func CaptureError(ctx context.Context, err error, callerSkip int) {
	hub := ctx.Value(ContextHubKey)
	if sHub, ok := hub.(*sentry.Hub); ok {
		sHub.WithScope(func(scope *sentry.Scope) {
			event := getEvent(err, callerSkip+1)
			sHub.CaptureEvent(event)
		})
	} else {
		event := getEvent(err, callerSkip)
		sentry.CaptureEvent(event)
	}
}

func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			ctx := c.Request.Context()
			ctx = context.WithValue(ctx, ContextHubKey, hub)
			c.Request = c.Request.WithContext(ctx)
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("rid", c.GetString("rid"))
			})
		}
	}
}
