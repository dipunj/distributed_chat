package network

import (
	"chat/pb"
	"net"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

// type for the internal server(handles replica connections)
type InternalServerType struct {
	pb.UnimplementedInternalServer

	SelfID     int
	GrpcServer *grpc.Server
	Listener   net.Listener
	DBPool     *pgxpool.Pool
}

type ReplicaStateType struct {
	Changed           chan bool
	Client            pb.InternalClient
	IsOnline          bool
	PublicIpAddress   string
	InternalIpAddress string
}

// the actual internal server object
var InternalServer = InternalServerType{}

// this object stores the client objects each replica

// ReplicaIds will be populated at run time
// based on what id is given to the current instance of the server

var SelfID int
var ReplicaIds = []int{}

var ReplicaState map[int]*ReplicaStateType = make(map[int]*ReplicaStateType)

func GetReplicaAddressFromID(replicaID int, port string) string {
	ip_prefix := "172.30.100.10"
	return ip_prefix + strconv.Itoa(replicaID) + ":" + port
}

func InitializeReplicas(replica_count int) {

	// populate replica_ids array with ids from 1 to replica_count, except selfID
	for i := 1; i <= replica_count; i++ {
		if i != SelfID {
			ReplicaIds = append(ReplicaIds, i)
			// initially we assume all replicas are offline
			ReplicaState[i] = &ReplicaStateType{
				Changed:           make(chan bool),
				IsOnline:          false,
				PublicIpAddress:   GetReplicaAddressFromID(i, DEFAULT_PUBLIC_PORT),
				InternalIpAddress: GetReplicaAddressFromID(i, DEFAULT_INTERNAL_PORT),
			}
		}
	}

}
