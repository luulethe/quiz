package db

import (
	"math/rand"

	"gorm.io/gorm"
)

type QuizDB interface {
	Master() *gorm.DB
	Slave() *gorm.DB
	Close() error
}

type quizDB struct {
	master   *gorm.DB
	replicas []*gorm.DB
}

func NewQuizDBFromConfig(config Config) (QuizDB, error) {
	master, err := dbConnect(config)
	if err != nil {
		return nil, err
	}
	replicas, err := dbConnectToReplica(config)
	if err != nil {
		return nil, err
	}

	return &quizDB{master: master, replicas: replicas}, nil
}

// NewNoteDBForTest creates a QuizDB object for testing
func NewNoteDBForTest(dbConn *gorm.DB) QuizDB {
	return &quizDB{dbConn, nil}
}

func (om *quizDB) Master() *gorm.DB {
	return om.master
}

func (om *quizDB) Slave() *gorm.DB {
	if len(om.replicas) > 0 {
		return om.replicas[rand.Intn(len(om.replicas))] //nolint
	}
	return om.master
}

func (om *quizDB) Close() (err error) {
	sql, err := om.master.DB()
	if err != nil {
		return
	}
	err = sql.Close()
	if err != nil {
		return
	}

	for _, replica := range om.replicas {
		sql, err := replica.DB()
		if err != nil {
			return err
		}
		err = sql.Close()
		if err != nil {
			return err
		}
	}
	return
}
