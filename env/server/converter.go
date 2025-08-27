package server

import (
	"golib/env"
	"golib/rule"
	grpc_gen "golib/server/grpc"
)

func ToGrpcExperience(e env.Experience) *grpc_gen.Experience {
	return &grpc_gen.Experience{
		State:   ToGrpcState(e.State),
		LogProb: e.LogProb,
		Critic:  e.Critic,
		Action:  e.Action,
		Reward:  e.Reward,
		Done:    e.Done,
	}
}

func ToGrpcExperiences(experiences []env.Experience) []*grpc_gen.Experience {
	grpcExperiences := make([]*grpc_gen.Experience, len(experiences))
	for i, e := range experiences {
		grpcExperiences[i] = ToGrpcExperience(e)
	}
	return grpcExperiences
}

func ToGrpcState(s *env.State) *grpc_gen.State {
	return &grpc_gen.State{
		BoardState:   ToGrpcBoardState(s.Board),
		FeatureState: &grpc_gen.FeatureState{Features: s.Feature},
	}
}

func ToGrpcBoardState(b *rule.Board) *grpc_gen.BoardState {
	return &grpc_gen.BoardState{
		Snakes: ToGrpcSnakes(b.Snakes),
		Food:   ToGrpcPoints(b.Food.ToSlice()),
	}
}

func ToGrpcSnakes(snakes []rule.Snake) []*grpc_gen.Snake {
	grpcSnakes := make([]*grpc_gen.Snake, len(snakes))
	for i, s := range snakes {
		grpcSnakes[i] = ToGrpcSnake(&s)
	}
	return grpcSnakes
}

func ToGrpcSnake(s *rule.Snake) *grpc_gen.Snake {
	if !s.Alive {
		return &grpc_gen.Snake{
			Body: []*grpc_gen.Point{},
		}
	}
	return &grpc_gen.Snake{
		Body: ToGrpcPoints(s.Body),
	}
}

func ToGrpcPoints(points []rule.Point) []*grpc_gen.Point {
	grpcPoints := make([]*grpc_gen.Point, len(points))
	for i, p := range points {
		grpcPoints[i] = ToGrpcPoint(p)
	}
	return grpcPoints
}

func ToGrpcPoint(p rule.Point) *grpc_gen.Point {
	return &grpc_gen.Point{
		X: int32(p.X),
		Y: int32(p.Y),
	}
}
