package network

import (
	"chat/pb"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

const REPLICA_COUNT int = 5

// ReplicaIds will be populated at run time
// based on what id is given to the current instance of the server
var ReplicaIds = []int{}

var OnlineReplicas map[string]bool

type ReplicaUser struct {
	client_id  string
	user_name  string
	group_name string
	is_online  bool
}

// type for the internal server(handles replica connections)
type InternalServerType struct {
	pb.UnimplementedInternalServer

	SelfID     int
	GrpcServer *grpc.Server
	Listener   net.Listener
	DBPool     *pgxpool.Pool
}

// the actual internal server object
var InternalServer = InternalServerType{}

// this object stores the client objects for the replicas
// internal client

var InternalClients map[int]pb.InternalClient = map[int]pb.InternalClient{}
