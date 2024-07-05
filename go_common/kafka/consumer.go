package kafka

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/luulethe/quiz/go_common/log"
)

func NewKafkaConsumerClient(brokers, consumerGroup, ver string) (sarama.ConsumerGroup, error) {
	if len(brokers) == 0 {
		return nil, errors.New("no Kafka bootstrap brokers defined")
	}
	if len(consumerGroup) == 0 {
		return nil, errors.New("no Kafka consumer group defined")
	}

	version, err := sarama.ParseKafkaVersion(ver)
	if err != nil {
		return nil, err
	}
	conf := sarama.NewConfig()
	conf.Version = version
	conf.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	conf.Consumer.Offsets.Initial = sarama.OffsetNewest
	return sarama.NewConsumerGroup(strings.Split(brokers, ","), consumerGroup, conf)
}

func ConsumerServe(
	ctx context.Context, wg *sync.WaitGroup, group sarama.ConsumerGroup, consumer sarama.ConsumerGroupHandler, topics []string,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := group.Consume(ctx, topics, consumer); err != nil {
				// dont fatal when kafka_service service is Down. retry after some time.
				log.Errorff(ctx, "Error from consumer|err:%v", err)
				<-time.After(time.Second * time.Duration(5))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()
}

type ConsumerConfig struct {
	Brokers       string `yaml:"brokers"`
	Version       string `yaml:"version"`
	ConsumerGroup string `yaml:"consumer_group"`
}
