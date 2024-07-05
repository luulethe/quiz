package manager

import (
	"context"
	"github.com/luulethe/quiz/quiz_lib/db/model"
	pb "github.com/luulethe/quiz/quiz_lib/pb/gen"
)

type QuizManager interface {
	JoinQuiz(ctx context.Context, quizID int64, userID int64) (error, pb.Error)
	HandleNewScoreChange(ctx context.Context, quizID int64) error
}

func NewQuizManager(dep *Dependency) QuizManager {
	return &QuizManagerImpl{dep: dep}
}

type QuizManagerImpl struct {
	dep *Dependency
}

func (q *QuizManagerImpl) JoinQuiz(ctx context.Context, quizID int64, userID int64) (error, pb.Error) {
	if !q.checkUserExited(userID) {
		return nil, pb.Error_ERROR_USER_NOT_EXISTED
	}

	err, quiz := q.dep.QuizDAO.FindQuizByID(ctx, quizID)
	if err != nil {
		return err, 0
	}

	if quiz == nil {
		return nil, pb.Error_ERROR_QUIZ_NOT_EXITED
	}

	if quiz.Status == model.QuizStatusFinished {
		return nil, pb.Error_ERROR_QUIZ_FINISHED
	}

	err, quizParticipant := q.dep.QuizDAO.FindQuizParticipant(ctx, quizID, userID)
	if err != nil {
		return err, 0
	}

	if quizParticipant != nil {
		return nil, pb.Error_ERROR_USER_JOINED
	}

	err, _ = q.dep.QuizDAO.CreateQuizParticipant(ctx, quizID, userID)
	if err != nil {
		return err, 0
	}

	q.sendLeaderBoardChangedMessage(ctx, quizID)

	return nil, pb.Error_ERROR_OK
}

func (n *QuizManagerImpl) HandleNewScoreChange(ctx context.Context, quizID int64) error {
	//implement here
	return nil
}

func (q *QuizManagerImpl) sendLeaderBoardChangedMessage(ctx context.Context, quizID int64) {
	//todo send message to kafka then notify new members join
}

func (q *QuizManagerImpl) checkUserExited(userID int64) bool {
	//todo check user existed by calling account service
	return true
}
