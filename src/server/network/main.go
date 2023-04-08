package network

import (
	"chat/pb"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type ReplicaUser struct {
	client_id  string
	user_name  string
	group_name string
	is_online  bool
}

type ResponseStream struct {
	stream     pb.Public_SubscribeServer
	client_id  string
	user_name  string
	group_name string
	is_online  bool
	error      chan error
}

// this server is used to server client requests
type PublicServerType struct {
	pb.UnimplementedPublicServer

	// streams to send messages to clients
	// hash from client_id to the stream object
	Subscribers map[string]*ResponseStream

	// this keeps track of users that are online replicas
	// a map from replica_id to the user object
	ReplicaSubscribers map[string]*ReplicaUser

	GrpcServer *grpc.Server
	Listener   net.Listener
	DBPool     *pgxpool.Pool
}

// this server is used to handle replication
type InternalServerType struct {
	pb.UnimplementedInternalServer

	SelfID         int
	OnlineReplicas map[string]bool
	GrpcServer     *grpc.Server
	Listener       net.Listener
	DBPool         *pgxpool.Pool
}

var PublicServer = PublicServerType{
	Subscribers: map[string]*ResponseStream{},
}

var InternalServer = InternalServerType{
	OnlineReplicas: map[string]bool{},
}

const REPLICA_COUNT int = 5

var REPLICA_IDS = []int{}
