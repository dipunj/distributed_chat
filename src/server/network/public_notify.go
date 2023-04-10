package network

import (
	"chat/pb"
	"context"

	log "github.com/sirupsen/logrus"
)

func notifyReactionUpdateToReplica(client_id string, ctx context.Context, msg *pb.Reaction) {
	msg_with_clock := &pb.ReactionWithClock{
		ClientId: client_id,
		Reaction: msg,
		Clock:    Clock,
	}
	for r_id, replica := range ReplicaState {
		log.Debug("Notifying replica ", r_id, " about reaction update from client_id", client_id)
		replica.Client.UpdateReaction(ctx, msg_with_clock)
	}

}

func notifyNewMessageToReplica(client_id string, ctx context.Context, msg *pb.TextMessage) {
	msg_with_clock := &pb.TextMessageWithClock{
		ClientId:    client_id,
		TextMessage: msg,
		Clock:       Clock,
	}
	for r_id, replica := range ReplicaState {
		log.Debug("Notifying replica ", r_id, " about new message from client_id", client_id)
		replica.Client.CreateNewMessage(ctx, msg_with_clock)
	}

}

func notifyReplicaAboutUserSwitch(client_id string, ctx context.Context, msg *pb.UserState) {
	msg_with_clock := &pb.UserStateWithClock{
		ReplicaId: int32(SelfServerID),
		ClientId:  client_id,
		UserState: msg,
		Clock:     Clock,
	}
	for r_id, replica := range ReplicaState {
		log.Debug("Notifying replica ", r_id, " about username change for client_id", client_id)
		replica.Client.SwitchUser(ctx, msg_with_clock)
	}

}

func notifyReplicaAboutGroupSwitch(client_id string, ctx context.Context, msg *pb.UserState) {
	msg_with_clock := &pb.UserStateWithClock{
		ReplicaId: int32(SelfServerID),
		ClientId:  client_id,
		UserState: msg,
		Clock:     Clock,
	}
	for r_id, replica := range ReplicaState {
		log.Debug("Notifying replica ", r_id, " about group change for client_id", client_id)
		replica.Client.SwitchGroup(ctx, msg_with_clock)
	}

}
