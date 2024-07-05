package manager

import (
	"context"
	"github.com/luulethe/quiz/config"
	"github.com/luulethe/quiz/go_common/metrics"
	"github.com/luulethe/quiz/quiz_lib/db"
)

// Dependency defines the global dependencies
type Dependency struct {
	DB          db.QuizDB
	QuizManager QuizManager
	QuizDAO     QuizDAO
	Stats       *metrics.StatsCollector
}

// Close release resources
func (d *Dependency) Close() {
	d.DB.Close()
}

// Init initializes the dependency
func (d *Dependency) Init(
	ctx context.Context, conf *config.Configuration, stats *metrics.StatsCollector, metricsCollection *MetricsCollection,
) error {
	d.Stats = stats
	mainDB := conf.MySQL[0]
	quizDBManager, err := db.NewQuizDBFromConfig(mainDB)
	if err != nil {
		return err
	}
	d.DB = quizDBManager
	d.QuizManager = NewQuizManager(d)
	d.QuizDAO = NewQuizDAO(d)

	return nil
}
