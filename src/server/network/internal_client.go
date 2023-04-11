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
)

const REPLICA_CONNECTION_TIMEOUT = 5 * time.Second

func ListenHeartBeat(state *ReplicaStateType, replicaId int) {

	stream, err := state.Client.SubscribeToHeartBeat(context.Background())

	if err != nil {
		log.Debug("An error occurred in streaming for replica id", replicaId, err.Error())
	} else {
		// loop infinitely
		for {
			//			log.Debug("Waiting for connection state to change from replica", replicaId)

			time.Sleep(500 * time.Millisecond)

			stream.Send(&pb.Clock{Clock: Clock}) // Send our timestamps
			hb_reply, err := stream.Recv()       // Get items newer than our timestamp

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

			// TODO: I don't think the way of updating the vector timestamps is
			// correct. What happens if we have a recent message but are missing
			// an older message from the replica?

			// Apply new messages
			for _, new_msg := range hb_reply.MsgWClock {
				insertNewMessage(new_msg.ClientId, new_msg.TextMessage, new_msg.Clock.Clock)
				Clock.UpdateFrom(new_msg.Clock.Clock)
			}

			// Apply new reactions
			for _, new_react := range hb_reply.ReactWClock {
				insertNewReaction(new_react.ClientId, new_react.Reaction, new_react.Clock.Clock)
				Clock[replicaId] = maxInt64(Clock[replicaId], new_react.Clock.Clock[replicaId])
				Clock.UpdateFrom(new_react.Clock.Clock)
			}
		}
	}

	state.IsOnline = false
}

func DialAndPing(replicaId int) {

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
		go DialAndPing(replicaId)
	}

}
