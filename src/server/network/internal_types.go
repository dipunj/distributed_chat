package network

import (
	"chat/pb"
	"net"

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
