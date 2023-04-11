package network

import (
	"chat/pb"
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

// we server requests to replicas on a different port
// This keeps the client-server and replica-replica communication separate

func (s *InternalServerType) SubscribeToHeartBeat(_ *emptypb.Empty, stream pb.Internal_SubscribeToHeartBeatServer) error {
	err := make(chan error)
	return <-err
}

func (s *InternalServerType) CreateNewMessage(ctx context.Context, msg_w_clock *pb.TextMessageWithClock) (*pb.Status, error) {
	log.Debug("[CreateNewMessage] (internal server) called")
	// this method will be called by the server A after
	// a new message is sent to server A DIRECTLY by the client

	// i.e any replicated messages will not be forwarded by the server A to other servers")
	msg := msg_w_clock.TextMessage
	msgTimestamp := msg_w_clock.Clock
	client_id := msg_w_clock.ClientId

	Clock.UpdateFrom(msgTimestamp)
	err := insertNewMessage(client_id, msg, msgTimestamp)

	if err == nil {
		log.Info("[CreateNewMessage] from replica",
			client_id,
			" sender name ",
			msg.SenderName)

		defer broadcastGroupUpdatesToImmediateMembers(msg.GroupName, PublicServer.Subscribers)

		return &pb.Status{Status: true}, err

	} else {
		log.Error("[CreateNewMessage] from replica",
			client_id,
			" sender name ",
			msg.SenderName,
			" error",
			err.Error())

		return &pb.Status{Status: false}, err
	}

}

func (s *InternalServerType) UpdateReaction(ctx context.Context, msg_w_clock *pb.ReactionWithClock) (*pb.Status, error) {
	log.Debug("[UpdateReaction] (internal server) called")
	// this method will be called by the server A after
	// a reaction update is sent to server A DIRECTLY by the client
	// i.e any replicated reaction update will not be forwarded by the server A to other servers

	msg := msg_w_clock.Reaction
	client_id := msg_w_clock.ClientId
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	err := insertNewReaction(client_id, msg, msgTimestamp)

	if err == nil {
		log.Info("[UpdateReaction] from replica for user with IP at replica",
			"client_id",
			client_id,
		)
		defer broadcastGroupUpdatesToImmediateMembers(msg.GroupName, PublicServer.Subscribers)
		return &pb.Status{Status: true}, err
	}

	return &pb.Status{Status: false}, err
}

func (s *InternalServerType) SwitchUser(ctx context.Context, msg_w_clock *pb.UserStateWithClock) (*pb.Status, error) {
	log.Debug("[SwitchUser] (internal server) called")
	// this method will be called by the server A after
	// a user switches user name

	user_ip := msg_w_clock.ClientId
	replica_id := int(msg_w_clock.ReplicaId)
	msg := msg_w_clock.UserState
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	handleSwitchUser(user_ip, replica_id, msg, &PublicServer.Subscribers, msgTimestamp)

	return &pb.Status{Status: true}, nil
}

func (s *InternalServerType) SwitchGroup(ctx context.Context, msg_w_clock *pb.UserStateWithClock) (*pb.Status, error) {
	log.Debug("[SwitchGroup] (internal server) called")
	// this method will be called by the server A after
	// a user switches from one group to another

	user_ip := msg_w_clock.ClientId
	replica_id := int(msg_w_clock.ReplicaId)
	msg := msg_w_clock.UserState
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	handleSwitchGroup(user_ip, replica_id, msg, &PublicServer.Subscribers, msgTimestamp)

	return &pb.Status{Status: true}, nil
}

func (s *InternalServerType) UserIsOffline(ctx context.Context, msg_w_clock *pb.ClientIdWithClock) (*pb.Status, error) {
	log.Debug("[UserIsOffline] (internal server) called")
	// this method will be called by the server A after
	// a user goes offline on server A

	user_ip := msg_w_clock.ClientId
	replica_id := int(msg_w_clock.ReplicaId)
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	handleUserIsOffline(user_ip, replica_id, &PublicServer.Subscribers, msgTimestamp)

	return &pb.Status{Status: true}, nil
}
