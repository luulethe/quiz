package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/luulethe/quiz/config"
	"github.com/luulethe/quiz/go_common/kafka"
	"github.com/luulethe/quiz/go_common/log"
	"github.com/luulethe/quiz/go_common/metrics"
	"github.com/luulethe/quiz/go_common/sentry"
	"github.com/luulethe/quiz/go_common/util"
	"github.com/luulethe/quiz/kafka_service/consumer"
	"github.com/luulethe/quiz/quiz_lib/manager"
)

var (
	confPath   = flag.String("config", fmt.Sprintf("config/%s.yml", os.Getenv("DEPLOY")), "set config file")
	consoleLog = flag.Bool("console", true, "enable console log")
	logPath    = os.Getenv("APP_LOG_PATH")
)

func main() {
	config := &config.Configuration{}
	ctx, cancel := context.WithCancel(context.Background())
	err := config.LoadFromFile(*confPath)
	util.ExitOnErr(ctx, err)

	if config.SentryDNS != "" {
		err := sentry.Init(config.SentryDNS)
		util.ExitOnErr(ctx, err)
		defer sentry.Recover()
	}

	ctx = util.InitLog(ctx, logPath, config.Debug, *consoleLog, log.FileConfig{
		log.ErrorLevel: {"error.log", "info.log"},
		log.InfoLevel:  {"info.log"},
	})
	defer log.Flush(ctx)
	log.Debugf(ctx, "config: %v\nStarting quiz event consumer\n", config)

	stats := metrics.StartMonitor(ctx, "note_kafka", config.ProfileAddr)

	dep := &manager.Dependency{}
	err = dep.Init(ctx, config, stats, nil)
	util.ExitOnErr(ctx, err)
	defer dep.Close()

	consumerGroup, err := kafka.NewKafkaConsumerClient(config.NoteKafka.Brokers, config.NoteKafka.ConsumerGroup, config.NoteKafka.Version)
	util.ExitOnErr(ctx, err)

	brokers := strings.Split(config.NoteKafka.Brokers, ",")
	kqueue, err := sarama.NewSyncProducer(brokers, nil)
	util.ExitOnErr(ctx, err)

	// setup consumer
	wg := &sync.WaitGroup{}
	topics := []string{consumer.NoteEventTopic}
	handler := consumer.NewNoteEventConsumer(ctx, dep, kqueue, stats)
	kafka.ConsumerServe(ctx, wg, consumerGroup, handler, topics)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Info(ctx, "terminating: context cancelled")
	case <-sigterm:
		log.Info(ctx, "terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err := consumerGroup.Close(); err != nil {
		log.Errorf(ctx, "Error closing client: %v", err)
	}
}
