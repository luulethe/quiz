package kafka

import (
	"context"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/luulethe/quiz/go_common/log"
)

func NewSyncKafkaProducer(ctx context.Context, addr string) (sarama.SyncProducer, error) {
	brokers := strings.Split(addr, ",")
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V1_1_0_0
	return sarama.NewSyncProducer(brokers, config)
}

func NewAsyncKafkaProducer(
	ctx context.Context, addr string, onSuccess func(*sarama.ProducerMessage), onError func(error),
) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V1_1_0_0
	brokers := strings.Split(addr, ",")
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	go func(ctx context.Context, producer sarama.AsyncProducer) {
		for {
			select {
			case result, ok := <-producer.Successes():
				if ok {
					log.Debugff(ctx, "KafkaProduce|message:%v|partition:%v|offset:%v", result.Value, result.Partition, result.Offset)
					if onSuccess != nil {
						onSuccess(result)
					}
				} else {
					return
				}
			case err, ok := <-producer.Errors():
				if ok {
					log.Errorff(ctx, "KafkaProduce|err:%v", err)
					if onError != nil {
						onError(err)
					}
				} else {
					return
				}
			}
		}
	}(ctx, producer)
	return producer, nil
}
