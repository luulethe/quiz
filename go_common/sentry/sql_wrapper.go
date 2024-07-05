package sentry

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-sql-driver/mysql"
)

type sDriver struct {
	driver.Driver
}

// Open opens a connection
func (drv *sDriver) Open(name string) (driver.Conn, error) {
	conn, err := drv.Driver.Open(name)
	if err != nil {
		return conn, err
	}
	_, ok := conn.(driver.ExecerContext)
	if !ok {
		return conn, nil
	}
	_, ok = conn.(driver.QueryerContext)
	if !ok {
		return conn, nil
	}
	conn = &sConn{conn}
	return conn, nil
}

type sConn struct {
	driver.Conn
}

func (c *sConn) addBreadcrumb(ctx context.Context, category string, query string, args []driver.NamedValue, start time.Time) {
	latency := time.Since(start).Milliseconds()
	var latencyErr error
	if latency > 200 {
		latencyErr = fmt.Errorf("high latency|db_query:%s|latency:%d", query, latency)
	}
	var argValues []string
	for _, namedValue := range args {
		argValues = append(argValues, fmt.Sprintf("%v", namedValue.Value))
	}
	breadcrumb := &sentry.Breadcrumb{
		Category: category,
		Message:  query,
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"latency": latency,
			"args":    strings.Join(argValues, ","),
		},
	}
	hub := ctx.Value(ContextHubKey)
	if sHub, ok := hub.(*sentry.Hub); ok {
		sHub.AddBreadcrumb(breadcrumb, nil)
		if latencyErr != nil {
			sHub.WithScope(func(scope *sentry.Scope) {
				event := getEvent(latencyErr, 1)
				event.Fingerprint = []string{"high_latency_db_query"}
				sHub.CaptureEvent(event)
			})
		}
	} else {
		sentry.AddBreadcrumb(breadcrumb)
		if latencyErr != nil {
			event := getEvent(latencyErr, 0)
			event.Fingerprint = []string{"high_latency_db_query"}
			sentry.CaptureEvent(event)
		}
	}
}

func (c *sConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	defer c.addBreadcrumb(ctx, "exec", query, args, time.Now())
	if execer, ok := c.Conn.(driver.ExecerContext); ok {
		return execer.ExecContext(ctx, query, args)
	}
	return nil, fmt.Errorf("execContext not implemented")
}

func (c *sConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	defer c.addBreadcrumb(ctx, "query", query, args, time.Now())
	if queryer, ok := c.Conn.(driver.QueryerContext); ok {
		return queryer.QueryContext(ctx, query, args)
	}
	return nil, fmt.Errorf("execContext not implemented")
}

func init() {
	sql.Register(HookDriverName, &sDriver{&mysql.MySQLDriver{}})
}
