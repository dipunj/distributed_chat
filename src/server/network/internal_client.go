// this file is used to create a client on each server to send messages to the other replicas

package network

import (
	"chat/pb"
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const REPLICA_CONNECTION_TIMEOUT = 5 * time.Second

var ReplicaConnectionWaitGroup = &sync.WaitGroup{}

func DialAndPing(replicaId int) {
	// this method will be called by a goroutine
	// this method will ping the replica every 1 second
	// if the replica is online, it will set the replica state to online
	// if the replica is offline, it will set the replica state to offline and call grpc dial every 1 second

	defer ReplicaConnectionWaitGroup.Done()
	state := ReplicaState[replicaId]
	ip_address := state.InternalIpAddress

	for {
		ctx, _ := context.WithTimeout(context.Background(), REPLICA_CONNECTION_TIMEOUT)

		conn, err := grpc.DialContext(
			ctx,
			ip_address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                5 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			}),
		)

		if err == nil {
			// create a new client
			state := ReplicaState[replicaId]
			state.Client = pb.NewInternalClient(conn)
			ReplicaState[replicaId] = state

			state.IsOnline <- true

			conn_state := conn.GetState()

			// the following call is blocking
			conn.WaitForStateChange(context.Background(), conn_state)

			// Check if the connection is still alive
			if conn_state == connectivity.TransientFailure || conn_state == connectivity.Shutdown {
				log.Error("Connection to replica", replicaId, " at ", ip_address, " is closed")
			}
		} else {
			log.Error("Dial failed: Replica ID ", replicaId, " at ", ip_address, " with error ", err)
		}

		// retry after 1 second
		state.IsOnline <- false
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
		ReplicaConnectionWaitGroup.Add(1)
		go DialAndPing(replicaId)
	}

}
