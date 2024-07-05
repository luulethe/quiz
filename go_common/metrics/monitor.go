package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/luulethe/quiz/go_common/log"
	"github.com/prometheus/client_golang/prometheus"
)

type ResultType string

const ResultSuccess ResultType = "Success"
const ResultNotInterested ResultType = "NotInterested"
const ResultWarning ResultType = "Warning"
const ResultError ResultType = "Error"

func StartMonitor(ctx context.Context, subsystem, addr string, extraCollectors ...prometheus.Collector) *StatsCollector {
	stats := NewStatsCollector("seatalk", subsystem, []string{"action", "result"}, QueueSize)
	mux := http.NewServeMux()
	stats.Bind(mux, extraCollectors...)
	stats.SetGauge(1, "ServerHealth", "")
	// When redeploy, Counter will be reset, set this to 0 to prevent NO DATA error.
	// "Success" is hard coded here, you can adjust grafana panel to use this.
	stats.ReportCount(0, "", string(ResultSuccess))

	if addr != "" {
		exporterServer := http.Server{
			Addr:    addr,
			Handler: mux,
		}
		go func() {
			if err := exporterServer.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalff(ctx, "startMonitor|fail to start exporter|err:%v", err)
				return
			}
		}()
	}
	return stats
}

func CollectStats(handleMessage func() ResultType, stats *StatsCollector, createTime time.Time, action string) {
	delay := float64(time.Since(createTime)) / float64(time.Millisecond)
	startTime := time.Now()
	result := handleMessage()

	if result == "" {
		// Panic Result
		result = ResultError
	} else if result == ResultNotInterested {
		return
	}

	duration := float64(time.Since(startTime)) / float64(time.Millisecond)
	stats.ReportCount(1, action, string(result))
	stats.ReportLatency(delay, action+"_delay", string(result))
	stats.ReportLatency(duration, action+"_duration", string(result))
}

func StatsWrapper(stats *StatsCollector, action string, handler func() ResultType) func() {
	return func() { CollectStats(handler, stats, time.Now(), action) }
}
