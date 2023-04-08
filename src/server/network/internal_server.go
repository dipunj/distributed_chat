package network

import (
	"chat/pb"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

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

// health check methods
func (s *InternalServerType) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	// when a replica calls check,
	// we inform the replica that we are healthy by returning SERVING
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}

func (s *InternalServerType) Watch(req *pb.HealthCheckRequest, stream pb.Internal_WatchServer) error {
	return stream.Send(&pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	})
}
