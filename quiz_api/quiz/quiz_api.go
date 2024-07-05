package quiz

import (
	"context"

	"github.com/luulethe/quiz/quiz_lib/manager"
	pb "github.com/luulethe/quiz/quiz_lib/pb/gen"
	"google.golang.org/protobuf/proto"
)

func JoinQuiz(ctx context.Context, dep *manager.Dependency, request *pb.RequestData, response *pb.ResponseData) (err error) {
	requestData := pb.JoinQuizRequest{}
	err = proto.Unmarshal(request.Request, &requestData)
	if err != nil {
		return
	}

	err, joinStatus := dep.QuizManager.JoinQuiz(ctx, requestData.QuizId, requestData.UserId)
	if err != nil {
		return err
	}

	reply := pb.JoinQuizRequestReply{}

	response.Response, err = proto.Marshal(&reply)
	response.Result = joinStatus

	return err
}
