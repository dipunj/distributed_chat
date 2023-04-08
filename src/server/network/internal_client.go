// this file is used to create a client on each server to send messages to the other replicas

package network

import (
	"chat/pb"
	"context"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DEFAULT_CONN_TIMEOUT = 5 * time.Second
)

var InternalClient pb.InternalClient = nil

func ConnectToReplica(replicaAddress string) {

	log.Info("Connecting to replica at ", replicaAddress)

	ctx, cancelContext := context.WithTimeout(context.Background(), DEFAULT_CONN_TIMEOUT)

	conn, err := grpc.DialContext(
		ctx,
		replicaAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	defer cancelContext()

	if err != nil {
		// raise error
		log.Info("Unable to dial to server ", err)
		return
	}

	InternalClient = pb.NewInternalClient(conn)

	log.Info("Successfully connected to replica at ", replicaAddress)
}

func GetReplicaAddressFromID(replicaID int) string {
	base_ip := "0.0.0.0"
	return base_ip + strconv.Itoa(replicaID)
}

func StartInternalClient() {
	// this function is called after the server has started
	// it creates a client that connects to the other replicas
	// and sends and receives messages from them
	for _, replicaID := range REPLICA_IDS {
		replicaAddress := GetReplicaAddressFromID(replicaID)
		ConnectToReplica(replicaAddress)
	}
}
