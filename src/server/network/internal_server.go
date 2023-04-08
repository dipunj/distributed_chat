package network

import (
	"chat/pb"
	"context"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *InternalServerType) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	// when a replica calls check,
	// we inform the replica that we are healthy by returning SERVING
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

func (s *InternalServerType) SyncOnlineUsers(ctx context.Context, msg *pb.UserStateWithClock) (*emptypb.Empty, error) {
	// this method will be called by the replica when user state changes
	// we need to update the user state in our server as well

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SyncMessages(ctx context.Context, msg *pb.TextMessageWithClock) (*emptypb.Empty, error) {
	// this method will be called by the replica when a new message is received on the replica
	// we need to update the message in our server as well

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SyncReactions(ctx context.Context, msg *pb.ReactionWithClock) (*emptypb.Empty, error) {
	// this method will be called by the replica when a reaction is created/removed on the replica
	// we need to update the reaction in our server as well

	return &emptypb.Empty{}, nil
}
