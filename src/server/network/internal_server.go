package network

import (
	"chat/pb"
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *InternalServerType) SendOnlineUsers(ctx context.Context, msg *pb.UserStateWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// the state of a user logged into the server A changes

	fmt.Println("NONON: SendOnlineUsers")

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SendMessages(ctx context.Context, msg *pb.TextMessageWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// a new message is sent to server A DIRECTLY by the client
	// i.e any replicated messages will not be forwarded by the server A to other servers

	fmt.Println("NONON: SendMessages\n", msg)

	// Insert the new message to our database
	fmt.Println("dbpool: ", s.DBPool)
	fmt.Println("ctx: ", ctx)
	insertNewMessage(s.DBPool, ctx, msg.TextMessage, VectorClock{clocks: msg.Clock})

	// TODO: If their timestamp was more recent than ours, request updates from the relevant replicas (cry).

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SendReactions(ctx context.Context, msg *pb.ReactionWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// a reaction update is sent to server A DIRECTLY by the client
	// i.e any replicated reaction update will not be forwarded by the server A to other servers

	fmt.Println("NONON: SendReactions")

	return &emptypb.Empty{}, nil
}
