package metrics

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type opType int

const (
	// QueueSize is the internal queue's size
	QueueSize = 10000

	// logFreq controls the frequency of printing logs
	logFreq     = 10000
	logIdle     = int32(0)
	logPrinting = int32(1)
)

const (
	_latency  opType = 1 + iota // add
	_counter                    // add
	_gauge                      // add
	_gaugeSet                   // set
)

var (
	// DefaultCollector is the default instance that has three labels:
	// "cmd", "queue_id", "result".
	DefaultCollector Collector = NewStatsCollector("", "", []string{"cmd", "queue_id", "result"}, QueueSize)
)

type task struct {
	op     opType
	data   float64
	labels []string
}

// Collector defines operations to report stats.
type Collector interface {
	// ReportLatency reports the latency.
	ReportLatency(latency float64, labels ...string)
	// ReportCount reports count.
	ReportCount(count float64, labels ...string)
	// ReportGauge reports gauge.
	ReportGauge(count float64, labels ...string)
	// SetGauge sets gauge as target value.
	SetGauge(val float64, labels ...string)
	// Bind binds collector to a http server mux
	Bind(mux *http.ServeMux, extraCollectors ...prometheus.Collector)
}

// StatsCollector implements the Collector interface.
// The percentiles are 0.2, 0.5, 0.8, 0.9, 0.95, 0.975, 0.99, 0.995.
type StatsCollector struct {
	mLatency *prometheus.SummaryVec
	mCounter *prometheus.CounterVec
	mGuage   *prometheus.GaugeVec

	queue      chan *task
	fullCnt    int32
	isPrinting int32
}

// NewStatsCollector creates  StatsCollector instance.
func NewStatsCollector(namespace, subsystem string, labels []string, qSize int) *StatsCollector {
	s := &StatsCollector{}
	s.mLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Subsystem:  subsystem,
			Name:       "Latency_percentile",
			Help:       "RPC latency percentile distributions.",
			Objectives: map[float64]float64{0.2: 0.01, 0.5: 0.01, 0.8: 0.01, 0.9: 0.01, 0.95: 0.001, 0.975: 0.001, 0.99: 0.001, 0.995: 0.001},
		},
		labels,
	)
	s.mCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "Counter",
		Help:      "Stats counter",
	},
		labels,
	)
	s.mGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "Gauge",
		Help:      "Stats Gauge",
	},
		labels,
	)
	s.queue = make(chan *task, qSize)
	return s
}

// ReportLatency reports latency info.
func (s *StatsCollector) ReportLatency(latency float64, labels ...string) {
	s.report(latency, _latency, labels...)
}

// ReportCount reports count info.
func (s *StatsCollector) ReportCount(count float64, labels ...string) {
	s.report(count, _counter, labels...)
}

// ReportGauge reports gauge info.
func (s *StatsCollector) ReportGauge(gauge float64, labels ...string) {
	s.report(gauge, _gauge, labels...)
}

// SetGauge sets gauge as target value.
func (s *StatsCollector) SetGauge(gauge float64, labels ...string) {
	s.report(gauge, _gaugeSet, labels...)
}

func (s *StatsCollector) report(data float64, op opType, labels ...string) {
	t := &task{data: data, op: op, labels: labels}
	select {
	case s.queue <- t:
	default:
		s.incFullCounter()
	}
}

func (s *StatsCollector) incFullCounter() {
	atomic.AddInt32(&s.fullCnt, 1)
	if s.fullCnt >= logFreq {
		atomic.StoreInt32(&s.fullCnt, 0)
		s.printLog()
	}
}

func (s *StatsCollector) printLog() {
	canPrint := atomic.CompareAndSwapInt32(&s.isPrinting, logIdle, logPrinting)
	if !canPrint {
		return
	}
	log.Printf("metrics: internal queue is full")
	atomic.StoreInt32(&s.isPrinting, logIdle)
}

// Start starts the HTTP serving goroutine to expose stats info.
func (s *StatsCollector) Bind(mux *http.ServeMux, extraCollectors ...prometheus.Collector) {
	r := prometheus.NewRegistry()
	goCollector := prometheus.NewGoCollector()
	processCollector := prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{})
	err := register(r, goCollector, processCollector, s.mLatency, s.mCounter, s.mGuage)
	if err != nil {
		log.Printf("metrics: can not register collectors, %s", err)
		return
	}
	err = register(r, extraCollectors...)
	if err != nil {
		log.Printf("metrics: can not register collectors, %s", err)
		return
	}
	// Expose the registered metrics via HTTP.
	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}))
	go s.run()
}

func register(r *prometheus.Registry, cs ...prometheus.Collector) error {
	for _, c := range cs {
		err := r.Register(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsCollector) run() {
	for t := range s.queue {
		switch t.op {
		case _latency:
			s.mLatency.WithLabelValues(t.labels...).Observe(t.data)
		case _counter:
			s.mCounter.WithLabelValues(t.labels...).Add(t.data)
		case _gauge:
			s.mGuage.WithLabelValues(t.labels...).Add(t.data)
		case _gaugeSet:
			s.mGuage.WithLabelValues(t.labels...).Set(t.data)
		}
	}
}
