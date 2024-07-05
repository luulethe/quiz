package manager

import (
	"context"
	"github.com/luulethe/quiz/quiz_lib/db/model"
	"gorm.io/gorm"
	"time"
)

type QuizDAO interface {
	FindQuizByID(ctx context.Context, quizID int64) (error, *model.QuizTab)
	CreateQuizParticipant(ctx context.Context, quizID int64, userID int64) (error, *model.QuizParticipantTab)
	FindQuizParticipant(ctx context.Context, quizID int64, userID int64) (error, *model.QuizParticipantTab)
}

func NewQuizDAO(dep *Dependency) QuizDAO {
	return &QuizDAOImpl{dep: dep}
}

type QuizDAOImpl struct {
	dep *Dependency
}

func (d *QuizDAOImpl) FindQuizByID(ctx context.Context, quizID int64) (error, *model.QuizTab) {
	quiz := model.QuizTab{}
	slave := d.dep.DB.Slave()
	sqlResult := slave.First(&quiz, quizID)

	if sqlResult.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if sqlResult.Error != nil {
		return sqlResult.Error, nil
	}

	return nil, &quiz
}

func (d *QuizDAOImpl) CreateQuizParticipant(ctx context.Context, quizID int64, userID int64) (error, *model.QuizParticipantTab) {
	quiz := &model.QuizParticipantTab{
		QuizID:      quizID,
		UserID:      userID,
		Score:       0,
		CreatedTime: time.Now().UnixMilli(),
		UpdatedTime: time.Now().UnixMilli(),
	}
	master := d.dep.DB.Master()
	sqlResult := master.Create(&quiz)
	if sqlResult.Error != nil {
		return sqlResult.Error, nil
	}

	return nil, quiz
}

func (d *QuizDAOImpl) FindQuizParticipant(ctx context.Context, quizID int64, userID int64) (error, *model.QuizParticipantTab) {
	quiz := model.QuizParticipantTab{}
	slave := d.dep.DB.Slave()
	sqlResult := slave.Where("quiz_id = ? and  user_id = ? ", quizID, userID).First(&quiz)

	if sqlResult.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if sqlResult.Error != nil {
		return sqlResult.Error, nil
	}

	return nil, &quiz
}
