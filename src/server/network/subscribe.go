package network

import (
	"chat/pb"

	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *GroupChatServer) Subscribe(_ *emptypb.Empty, stream pb.GroupChat_SubscribeServer) error {
	ctx := stream.Context()
	client, _ := peer.FromContext(ctx)
	clientID := client.Addr.String()

	rs := &ResponseStream{
		stream:     stream,
		client_id:  clientID,
		user_name:  "",
		group_name: "",
		is_online:  true,
		error:      make(chan error),
	}

	s.Subscribers[clientID] = rs

	// Check if the client is offline/disconnected
	go func() {
		<-ctx.Done()

		// Update the isOnline field of the ResponseStream to false
		rs.is_online = false
		s.broadcastUpdates(rs.group_name)

		// Remove the ResponseStream from the Subscribers map
		delete(s.Subscribers, clientID)
	}()

	return <-rs.error
}
