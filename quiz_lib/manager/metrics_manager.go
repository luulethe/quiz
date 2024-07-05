package manager

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/luulethe/quiz/go_common/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultNamespace = "now"
)

type MetricsCollection struct {
	Collectors []prometheus.Collector
}

func (mc *MetricsCollection) AddCollector(c prometheus.Collector) {
	mc.Collectors = append(mc.Collectors, c)
}

type ServiceMetrics struct {
	serviceName           string
	requestCounter        *prometheus.CounterVec
	requestLatencySummary *prometheus.SummaryVec
}

func NewServiceMetrics(serviceName string, metricsCollection *MetricsCollection) *ServiceMetrics {
	if metricsCollection == nil {
		return nil
	}
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: defaultNamespace,
			Subsystem: serviceName,
			Name:      "request_counter",
			Help:      fmt.Sprintf("%s API request counters", serviceName),
		},
		[]string{"api", "result", "status_code"},
	)
	requestLatencySummary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  defaultNamespace,
			Subsystem:  serviceName,
			Name:       "request_latency_summary",
			Help:       fmt.Sprintf("%s API request latency percentile distributions", serviceName),
			Objectives: map[float64]float64{0.2: 0.01, 0.5: 0.01, 0.8: 0.01, 0.9: 0.01, 0.95: 0.001, 0.975: 0.001, 0.99: 0.001, 0.995: 0.001},
		},
		[]string{"api", "result", "status_code"},
	)
	metricsCollection.AddCollector(requestCounter)
	metricsCollection.AddCollector(requestLatencySummary)

	return &ServiceMetrics{
		serviceName:           serviceName,
		requestCounter:        requestCounter,
		requestLatencySummary: requestLatencySummary,
	}
}

func (sm *ServiceMetrics) reportCount(labels ...string) {
	sm.requestCounter.WithLabelValues(labels...).Inc()
}

func (sm *ServiceMetrics) reportLatency(latency float64, labels ...string) {
	sm.requestLatencySummary.WithLabelValues(labels...).Observe(latency)
}

type HTTPHandler func(req *http.Request) (*http.Response, error)

func (sm *ServiceMetrics) WithMetricReport(do HTTPHandler) HTTPHandler {
	return func(req *http.Request) (resp *http.Response, err error) {
		// Monitoring
		var latency float64
		var caller, statusCode string
		defer func() {
			result := string(metrics.ResultSuccess)
			if err != nil {
				result = string(metrics.ResultError)
			}
			sm.reportCount(caller, result, statusCode)
			sm.reportLatency(latency, caller, result, statusCode)
		}()

		// Get API name from caller name
		pc, _, _, ok := runtime.Caller(4)
		details := runtime.FuncForPC(pc)
		if ok && details != nil {
			parts := strings.Split(details.Name(), ".")
			caller = parts[len(parts)-1]
		}

		start := time.Now()
		resp, err = do(req)
		latency = float64(time.Since(start)) / float64(time.Millisecond)
		if err != nil {
			return
		}
		statusCode = strconv.Itoa(resp.StatusCode)
		return
	}
}
