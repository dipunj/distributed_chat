// this file is used to create a client on each server to send messages to the other replicas

package network

import (
	"chat/pb"
	"context"
	"math"
	"strconv"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const REPLICA_CONNECTION_TIMEOUT = 5 * time.Second

func ConnectToReplica(ip_address string, replicaID int) {

	log.Info("Connecting to replica at ", ip_address)

	ctx, cancelContext := context.WithTimeout(context.Background(), REPLICA_CONNECTION_TIMEOUT)

	conn, err := grpc.DialContext(
		ctx,
		ip_address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(
				grpc_retry.WithMax(math.MaxUint64),
				grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
			),
		),
	)

	defer cancelContext()

	if err != nil {
		// raise error
		log.Error("Unable to dial to replica", replicaID, ip_address, "because of:", err)
		return
	}

	InternalClients[replicaID] = pb.NewInternalClient(conn)

	log.Info("Successfully connected to replica at ", ip_address)
}

func GetReplicaAddressFromID(replicaID int) string {
	ip_prefix := "0.0.0.0"
	return ip_prefix + strconv.Itoa(replicaID)
}

func StartInternalClients() {
	for _, replicaID := range ReplicaIds {
		replicaAddress := GetReplicaAddressFromID(replicaID)
		ConnectToReplica(replicaAddress, replicaID)
	}
}
