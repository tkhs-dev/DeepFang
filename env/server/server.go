package server

import (
	"fmt"
	"golib/env"
	"golib/neuralnet"
	"golib/rule"
	grpc_gen "golib/server/grpc"
	"golib/server/service"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type streamSenderImpl func(chunk *service.ExperienceChunk) error

func (s streamSenderImpl) Send(chunk *service.ExperienceChunk) error {
	return s(chunk)
}

type Server struct {
	grpc_gen.UnimplementedEnvServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s Server) CollectExperience(request *grpc_gen.CollectExperienceRequest, g grpc.ServerStreamingServer[grpc_gen.CollectExperienceResponse]) error {
	collector := service.NewStreamingExperienceCollector(
		int(request.ChunkSize),
		streamSenderImpl(func(chunk *service.ExperienceChunk) error {
			err := g.Send(&grpc_gen.CollectExperienceResponse{
				Experiences: ToGrpcExperiences(chunk.Experiences),
			})
			if err != nil {
				return fmt.Errorf("failed to send chunk: %v", err)
			}
			return nil
		}), nil)

	args := request.GetArgs()

	epArg := &env.EpisodeArg{
		Setting: &rule.Setting{
			Width:           int(args.BoardWidth),
			Height:          int(args.BoardHeight),
			Seed:            time.Now().UnixNano(),
			FoodSpawnChance: int(args.FoodSpawnChance),
			MinimumFood:     int(args.MinimumFood),
			InitialTurn:     0,
		},
	}

	if len(request.Models) == 0 {
		return status.Error(codes.InvalidArgument, "At least one model must be provided")
	}
	models := make(map[string]neuralnet.Model)
	for i, model := range request.Models {
		onnx, err := neuralnet.NewOnnxModel(model, len(args.Snakes), int(args.BoardWidth), int(args.BoardHeight))
		if err != nil {
			println("Failed to create ONNX model:", err.Error())
			return status.Error(codes.Internal, "failed to create ONNX model")
		}
		models["model:"+strconv.Itoa(i)] = onnx
	}

	for _, snakeName := range args.Snakes {
		if _, exists := models[snakeName]; !exists {
			return status.Error(codes.InvalidArgument, "At least one snake must be powered by a model")
		}
	}

	err := collector.CollectExperiencesStreaming(g.Context(), epArg, models, args.Snakes, int(request.StepNum), int(request.WorkerNum), 10*time.Millisecond)
	if err != nil {
		println("Failed to collect experiences:", err.Error())
		if grpcStatus, ok := status.FromError(err); ok {
			return grpcStatus.Err()
		}
		return status.Error(codes.Internal, "failed to collect experiences")
	}
	return nil
}
