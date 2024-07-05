package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/luulethe/quiz/config"
	"github.com/luulethe/quiz/go_common/log"
	"github.com/luulethe/quiz/go_common/metrics"
	"github.com/luulethe/quiz/go_common/sentry"
	"github.com/luulethe/quiz/go_common/util"
	"github.com/luulethe/quiz/quiz_api"
	"github.com/luulethe/quiz/quiz_lib/manager"
	rpc "github.com/luulethe/quiz/quiz_lib/pb/gen"
	"github.com/zyxar/grace/fork"
	"github.com/zyxar/grace/sigutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var (
	confPath   = flag.String("config", fmt.Sprintf("config/%s.yml", os.Getenv("DEPLOY")), "set config file")
	consoleLog = flag.Bool("console", true, "enable console log")
	logPath    = os.Getenv("APP_LOG_PATH")
)

func main() {
	flag.Parse()

	conf := &config.Configuration{}
	util.ExitOnErr(log.DefaultContext, conf.LoadFromFile(*confPath))

	ctx := util.InitLog(context.Background(), logPath, conf.Debug, *consoleLog, log.FileConfig{
		log.ErrorLevel: {"error.log", "info.log"},
		log.InfoLevel:  {"info.log"},
	})
	defer log.Flush(ctx)
	log.Debug(ctx, "Starting GRPC Http Server\n")

	if conf.SentryDNS != "" {
		err := sentry.Init(conf.SentryDNS)
		util.ExitOnErr(ctx, err)
		for _, db := range conf.MySQL {
			db.DriverName = sentry.HookDriverName
		}
		defer sentry.Flush()
	}

	statCollector := metrics.NewStatsCollector("now", "grpc_quiz", []string{"action", "result"}, metrics.QueueSize)
	statCollector.SetGauge(1, "ServerHealth", "")
	extraMetrics := &manager.MetricsCollection{}

	dependency := &manager.Dependency{}
	err := dependency.Init(ctx, conf, statCollector, extraMetrics)
	util.ExitOnErr(ctx, err)
	defer dependency.Close()

	if conf.ProfileAddr != "" { // setup pprof & metrics
		l, err := fork.Listen("tcp4", conf.ProfileAddr)
		if err != nil {
			log.Errorff(ctx, "pprof|listen|err:%v", err)
		} else {
			mux := http.NewServeMux()
			statCollector.Bind(mux, extraMetrics.Collectors...)
			srv := &http.Server{Handler: mux}
			go func() {
				if err := srv.Serve(l); err != http.ErrServerClosed {
					log.Errorff(ctx, "pprof|http.Serve|err:%v", err.Error())
				}
			}()
			defer util.WithErrorCaptured(ctx, func() error {
				return srv.Shutdown(ctx)
			}, "pprofServer.Shutdown")
		}
	}

	// ======================= GRPC Service ======================= //

	customFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}
	quizServer := quiz_api.NewQuizServer(ctx, dependency)
	gRPCServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(opts...),
		))
	// todo setup prometheus
	// grpc_handler.UnaryInterceptor(metricsInterceptor(requestCounter, requestLatencySummary)),
	// )
	rpc.RegisterQuizServiceServer(gRPCServer, quizServer)

	healthServer := health.NewServer()
	healthrpc.RegisterHealthServer(gRPCServer, healthServer)
	healthServer.SetServingStatus("notice_server", healthrpc.HealthCheckResponse_SERVING)

	// gRPC listener
	gRPCListener, err := fork.Listen("tcp4", conf.Listen)
	exitOnErr(ctx, err)
	log.Infof(ctx, "gRPC server listener: %s", gRPCListener.Addr().String())

	{
		go func() {
			l, _ := fork.TCPKeepAlive(gRPCListener, time.Minute*5)
			if err := gRPCServer.Serve(l); err != nil {
				log.Fatalff(ctx, "gRPC server not sering|err:%v", err)
			}
		}()
		defer gRPCServer.GracefulStop()
	}

	err = fork.SignalParent()
	if err != nil {
		log.Errorff(ctx, "[SIGNAL] parent|err:%v", err)
	}

	sigutil.Trap(func(s sigutil.Signal) {
		switch s {
		case sigutil.SIGHUP:
			pid, err := fork.ReloadAll()
			if err != nil {
				log.Errorf(ctx, "[RELOADING] err=%v", err)
			} else {
				log.Infof(ctx, "[RELOADING] -> %d", pid)
				err = writePidFile(pid)
				if err != nil {
					log.Errorff(ctx, "[RELOADING]|writePidFile|err:%v", err)
				}
			}
		default:
			log.Infof(ctx, "[CLOSING] by signal <%v>", s)
		}
		log.Infof(ctx, "[KILLED] by signal <%v>", s)
	}, sigutil.SIGINT, sigutil.SIGTERM, sigutil.SIGHUP, sigutil.SIGQUIT)

}

// for fatal error on server initialization
func exitOnErr(ctx context.Context, err error) {
	if err != nil {
		log.Error(ctx, err.Error())
		os.Exit(1)
	}
}

// writePidFile writes the (child) process ID to the file at pidfile.
func writePidFile(pid int) error {
	return ioutil.WriteFile("/app/grpc_sample.pid", []byte(strconv.Itoa(pid)+"\n"), 0600)
}
