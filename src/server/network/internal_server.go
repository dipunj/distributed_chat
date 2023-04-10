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

	err := insertNewMessage(client_id, msg, msgTimestamp)

	if err == nil {
		log.Info("[CreateNewMessage] from replica",
			client_id,
			"message sender name",
			msg.SenderName,
			"message text",
			msg.Content)

		defer broadcastGroupUpdatesToImmediateMembers(msg.GroupName, PublicServer.Subscribers)

		return &pb.Status{Status: true}, err

	} else {
		log.Error("[CreateNewMessage] from replica",
			client_id,
			"message sender name",
			msg.SenderName,
			"message text",
			msg.Content,
			"error",
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
	reactionTimestamp := msg_w_clock.Clock

	// TODO: update clock?

	err := insertNewReaction(client_id, msg, reactionTimestamp)

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
	timestamp := msg_w_clock.Clock

	// TODO: update clock?

	handleSwitchUser(user_ip, replica_id, msg, &PublicServer.Subscribers, timestamp)

	return &pb.Status{Status: true}, nil
}

func (s *InternalServerType) SwitchGroup(ctx context.Context, msg_w_clock *pb.UserStateWithClock) (*pb.Status, error) {
	log.Info("[SwitchGroup] (internal server) called")
	log.Debug("[SwitchGroup] (internal server) called")
	// this method will be called by the server A after
	// a user switches from one group to another

	user_ip := msg_w_clock.ClientId
	replica_id := int(msg_w_clock.ReplicaId)
	msg := msg_w_clock.UserState
	timestamp := msg_w_clock.Clock

	// TODO: update clock?

	handleSwitchGroup(user_ip, replica_id, msg, &PublicServer.Subscribers, timestamp)

	return &pb.Status{Status: true}, nil
}
