package quiz_api

import (
	"github.com/luulethe/quiz/quiz_api/quiz"
	pb "github.com/luulethe/quiz/quiz_lib/pb/gen"
)

var routers = map[pb.Command]HandlerFunc{
	pb.Command_CMD_JOIN_QUIZ: middlewareGroup.Wrap(quiz.JoinQuiz),
}
