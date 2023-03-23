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

type GroupChatServer struct {
	pb.UnimplementedGroupChatServer

	// streams to send messages to clients
	// hash from client_id to the stream object
	Subscribers map[string]*ResponseStream
	GrpcServer  *grpc.Server
	Listener    net.Listener
	DBPool      *pgxpool.Pool
}

var Server GroupChatServer = GroupChatServer{
	Subscribers: map[string]*ResponseStream{},
}
