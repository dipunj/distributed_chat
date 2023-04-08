package network

import (
	"chat/pb"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type ResponseStream struct {
	stream     pb.Public_SubscribeServer
	client_id  string
	user_name  string
	group_name string
	is_online  bool
	error      chan error
}

// type for the public server(handles client connections)
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

// the actual public server object
var PublicServer = PublicServerType{
	Subscribers: map[string]*ResponseStream{},
}
