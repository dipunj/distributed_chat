package network

import (
	"chat/pb"
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

func notifyReactionUpdateToReplica(client_id string, msg *pb.Reaction) {
	ctx, cancel := context.WithCancel(context.Background())

	msg_with_clock := &pb.ReactionWithClock{
		ClientId: client_id,
		Reaction: msg,
		Clock:    Clock,
	}
	var wg sync.WaitGroup
	for r_id, replica := range ReplicaState {
		wg.Add(1)
		client := &replica.Client

		go func(r_id int, client *pb.InternalClient) {
			defer wg.Done()
			log.Debug("Notifying replica ", r_id, " about reaction update from client_id", client_id)
			(*client).UpdateReaction(ctx, msg_with_clock)
		}(r_id, client)
	}
	wg.Wait()
	cancel()

}

func notifyNewMessageToReplica(client_id string, msg *pb.TextMessage) {
	ctx, cancel := context.WithCancel(context.Background())

	msg_with_clock := &pb.TextMessageWithClock{
		ClientId:    client_id,
		TextMessage: msg,
		Clock:       Clock,
	}
	var wg sync.WaitGroup
	for r_id, replica := range ReplicaState {
		wg.Add(1)
		client := &replica.Client

		go func(r_id int, client *pb.InternalClient) {
			defer wg.Done()
			log.Debug("Notifying replica ", r_id, " about new message from client_id", client_id)
			(*client).CreateNewMessage(ctx, msg_with_clock)
		}(r_id, client)
	}
	wg.Wait()
	cancel()
}

func notifyReplicaAboutUserSwitch(client_id string, msg *pb.UserState) {
	ctx, cancel := context.WithCancel(context.Background())

	msg_with_clock := &pb.UserStateWithClock{
		ReplicaId: int32(SelfServerID),
		ClientId:  client_id,
		UserState: msg,
		Clock:     Clock,
	}
	var wg sync.WaitGroup
	for r_id, replica := range ReplicaState {
		wg.Add(1)
		client := &replica.Client

		go func(r_id int, client *pb.InternalClient) {
			defer wg.Done()
			log.Debug("Notifying replica ", r_id, " about username change for client_id", client_id)
			(*client).SwitchUser(ctx, msg_with_clock)
		}(r_id, client)

	}
	wg.Wait()
	cancel()

}

func notifyReplicaAboutGroupSwitch(client_id string, msg *pb.UserState) {
	ctx, cancel := context.WithCancel(context.Background())

	msg_with_clock := &pb.UserStateWithClock{
		ReplicaId: int32(SelfServerID),
		ClientId:  client_id,
		UserState: msg,
		Clock:     Clock,
	}
	var wg sync.WaitGroup
	for r_id, replica := range ReplicaState {
		wg.Add(1)
		client := &replica.Client

		go func(r_id int, client *pb.InternalClient) {
			defer wg.Done()
			log.Debug("Notifying replica ", r_id, " about group change for client_id", client_id)
			(*client).SwitchGroup(ctx, msg_with_clock)
		}(r_id, client)

	}
	wg.Wait()
	cancel()
}

func notifyReplicaAboutOfflineImmediateUser(client_id string) {

	ctx, cancel := context.WithCancel(context.Background())

	msg_with_clock := &pb.ClientIdWithClock{
		ReplicaId: int32(SelfServerID),
		ClientId:  client_id,
		Clock:     Clock,
	}
	var wg sync.WaitGroup
	for r_id, replica := range ReplicaState {
		wg.Add(1)
		client := &replica.Client

		go func(r_id int, client *pb.InternalClient) {
			defer wg.Done()
			(*client).UserIsOffline(ctx, msg_with_clock)
		}(r_id, client)

	}

	wg.Wait()
	cancel()
}
