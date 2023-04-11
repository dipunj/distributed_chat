// this file is used to create a client on each server to send messages to the other replicas

package network

import (
	"chat/pb"
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const REPLICA_CONNECTION_TIMEOUT = 5 * time.Second

func ListenHeartBeat(state *ReplicaStateType, replicaId int) {

	stream, err := state.Client.SubscribeToHeartBeat(context.Background(), &emptypb.Empty{})

	if err != nil {
		log.Debug("An error occurred in streaming for replica id", replicaId, err.Error())
	} else {
		// loop infinitely
		for {
			log.Debug("Waiting for connection state to change from replica", replicaId)

			// the following call is blocking
			_, err := stream.Recv()

			if err != nil {
				code := status.Code(err)

				if code == codes.Unavailable {
					// connection was closed because the server is unavailable
					log.Debug("Connection was closed because the server is unavailable code = ", code)
				} else if code == codes.Canceled {
					// connection was closed by the client
					log.Debug("Connection was closed by the client code = ", code)
				} else {
					log.Debug("An error occurred in streaming", err.Error())
				}
				// break because there is an error with the connection
				break
			}
		}
	}

	state.IsOnline = false
}

// this function dials to the replica and
// hooks onto the replica using a stream
// when the stream is closed, it will retry to connect to the replica

func DialAndMonitor(replicaId int) {

	state := ReplicaState[replicaId]
	ip_address := state.InternalIpAddress

	// this loop keeps on retrying to connect to the replica
	// once the connection is established

	for {
		ctx, _ := context.WithTimeout(context.Background(), REPLICA_CONNECTION_TIMEOUT)

		conn, err := grpc.DialContext(
			ctx,
			ip_address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)

		if err == nil {
			log.Info("[DialAndPing] TCP Dial successful: Replica ID ", replicaId, " at ", ip_address)

			// create a new client on this connection
			state.Client = pb.NewInternalClient(conn)

			SyncReplica(replicaId)

			state.IsOnline = true

			ListenHeartBeat(state, replicaId)

			if !state.IsOnline {
				// if the replica is offline, close the connection
				conn.Close()
			}
		} else {
			log.Error("[DialAndPing] TCP Dial failed: Replica ID ", replicaId, " at ", ip_address, " with error ", err)
		}

		// retry after 1 second
		time.Sleep(1 * time.Second)
	}
}

func ConnectToReplicas() {
	// this method will start REPLICA_COUNT - 1 goroutines to connect to each of the replicas

	// each go routine will run indefinitely
	// the go routine will ping the replica to check if it is online
	// if the replica is online, it will set the replica state to online
	// if the replica is offline, it will set the replica state to offline and call grpc dial every 1 second

	for _, replicaId := range ReplicaIds {
		go DialAndMonitor(replicaId)
	}

}

func SyncReplica(replicaId int) {
	// this function will sync the replica with the current state of the server

	// 1. request replica's latest message's clock
	// 2. find all the messages happened after the replica's latest message and be
	// 3. send the messages to the replica
	state := ReplicaState[replicaId]
	client := state.Client
	replica_clock, err := client.GetLatestClock(context.Background(), &emptypb.Empty{})

	messages := GetMessagesAfter(replica_clock.Clock)
	client.PushDBMessages(context.Background(), &pb.DBMessages{Messages: messages})

}

func GetMessagesAfter(clock int64) []*pb.DBMessages {
	// this function will return all the messages after the clock

	messages := make([]*pb.DBMessages, 0)

	// query the database to get all the messages after the clock
	// append the messages to the messages array

	return messages
}
