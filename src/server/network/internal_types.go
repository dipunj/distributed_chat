package network

import (
	"chat/pb"
	"net"

	"google.golang.org/grpc"
)

// type for the internal server(handles replica connections)
type InternalServerType struct {
	pb.UnimplementedInternalServer

	SelfID     int
	GrpcServer *grpc.Server
	Listener   net.Listener
}

type ReplicaStateType struct {
	// we listen to this channel to know when to sync
	ShouldSync chan bool

	Client            pb.InternalClient
	IsOnline          bool
	PublicIpAddress   string
	InternalIpAddress string
}
