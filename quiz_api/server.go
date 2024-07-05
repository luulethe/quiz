package quiz_api

import (
	"context"

	"github.com/luulethe/quiz/go_common/sentry"
	"github.com/luulethe/quiz/quiz_lib/manager"
	pb "github.com/luulethe/quiz/quiz_lib/pb/gen"
)

type Server struct {
	pb.UnimplementedQuizServiceServer
	dep *manager.Dependency
}

func NewQuizServer(ctx context.Context, dep *manager.Dependency) *Server {
	return &Server{dep: dep}
}

func (s *Server) Handle(ctx context.Context, in *pb.RequestData) (*pb.ResponseData, error) {
	return handleCommand(ctx, s.dep, &in.Command, in)
}

func handleCommand(ctx context.Context, dep *manager.Dependency, command *pb.Command, in *pb.RequestData) (*pb.ResponseData, error) {
	res := &pb.ResponseData{}
	err := routers[*command](ctx, dep, in, res)
	if err != nil {
		sentry.CaptureError(ctx, err, 0)
	}
	return res, err
}
