package consumer

import (
	"context"
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/luulethe/quiz/go_common/log"
	"github.com/luulethe/quiz/go_common/metrics"
	"github.com/luulethe/quiz/quiz_lib/manager"
)

const (
	NoteEventTopic = "quiz_score_changed_event"
)

type LeaderBoardChangedMessage struct {
	quizID int64 `json:"quiz_id"`
}

// OASyncConsumer represents a Sarama consumer group consumer
type NoteEventConsumer struct {
	ctx    context.Context
	dep    *manager.Dependency
	kqueue sarama.SyncProducer
	stats  *metrics.StatsCollector
}

func NewNoteEventConsumer(
	ctx context.Context, dep *manager.Dependency, kqueue sarama.SyncProducer, stats *metrics.StatsCollector,
) *NoteEventConsumer {
	return &NoteEventConsumer{
		ctx:    ctx,
		dep:    dep,
		kqueue: kqueue,
		stats:  stats,
	}
}

func (c *NoteEventConsumer) Setup(sarama.ConsumerGroupSession) error {
	log.Info(c.ctx, "oa company event consumer is running!...")
	return nil
}

func (c *NoteEventConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *NoteEventConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		msg := &LeaderBoardChangedMessage{}
		err := json.Unmarshal(message.Value, msg)
		if err != nil {
			log.Errorff(c.ctx, "unmarshal error|err:%v|value:%v", err, string(message.Value))
			continue
		}
		err = c.dep.QuizManager.HandleNewScoreChange(c.ctx, msg.quizID)
		if err != nil {
			log.Errorf(c.ctx, "event_consumer_error|err:%v", err)
		}
		log.Infof(c.ctx, "event_consumer|quiz_id:%v", msg.quizID)

		session.MarkMessage(message, "")
	}
	return nil
}
