package network

import (
	"chat/pb"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *InternalServerType) SubscribeToHeartBeat(_ *emptypb.Empty, stream pb.Internal_SubscribeToHeartBeatServer) error {
	err := make(chan error)
	return <-err
}

func (s *InternalServerType) SendOnlineUsers(ctx context.Context, msg *pb.UserStateWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// the state of a user logged into the server A changes

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SendMessages(ctx context.Context, msg *pb.TextMessageWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// a new message is sent to server A DIRECTLY by the client
	// i.e any replicated messages will not be forwarded by the server A to other servers

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SendReactions(ctx context.Context, msg *pb.ReactionWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// a reaction update is sent to server A DIRECTLY by the client
	// i.e any replicated reaction update will not be forwarded by the server A to other servers

	return &emptypb.Empty{}, nil
}
