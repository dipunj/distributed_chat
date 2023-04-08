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
