package network

import (
	"chat/pb"
	"net"

	"google.golang.org/grpc"
)

type ResponseStream struct {
	// denotes the id of the server that the client is currently connected to
	server_id int

	// will be null if the client is connected to a peer replica
	stream pb.Public_SubscribeServer

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

	GrpcServer *grpc.Server
	Listener   net.Listener
}
