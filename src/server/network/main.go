package network

import (
	"chat/pb"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type ResponseStream struct {
	stream     pb.GroupChat_SubscribeServer
	client_id  string
	user_name  string
	group_name string
	is_online  bool
	error      chan error
}

// this server is used to server client requests
type PublicServerType struct {
	pb.UnimplementedGroupChatServer

	// streams to send messages to clients
	// hash from client_id to the stream object
	Subscribers map[string]*ResponseStream
	GrpcServer  *grpc.Server
	Listener    net.Listener
	DBPool      *pgxpool.Pool
}

// this server is used to handle replication
type ReplicationServerType struct {
	pb.UnimplementedReplicationServer

	selfID         int
	vector_clock   VectorClock
	onlineReplicas map[string]bool
	GrpcServer     *grpc.Server
	Listener       net.Listener
	DBPool         *pgxpool.Pool
}

var PublicServer = PublicServerType{
	Subscribers: map[string]*ResponseStream{},
}

var InternalServer = ReplicationServerType{
	onlineReplicas: map[string]bool{},
}
