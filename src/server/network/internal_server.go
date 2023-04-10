package network

import (
	"chat/pb"
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// we server requests to replicas on a different port
// This keeps the client-server and replica-replica communication separate
func ServeInternalRequests() {

	log.Info("[ServeInternalRequests] Starting internal (replication) service at", INTERNAL_ADDRESS)

	InternalServer.GrpcServer = grpc.NewServer()

	//	pb.RegisterInternalServer(InternalServer.GrpcServer, &InternalServerType{})
	pb.RegisterInternalServer(InternalServer.GrpcServer, &InternalServer)

	// Serve() spawns a new goroutine under the hood for each new request
	l := getTCPListener(INTERNAL_ADDRESS)
	err := InternalServer.GrpcServer.Serve(l)

	if err != nil {
		log.Fatalf("[ServeInternalRequests]: Error while starting the gRPC server on the %s listen address %v", l, err.Error())
	} else {
		log.Info("[ServeInternalRequests]: Internal Server started")
	}
}

func (s *InternalServerType) SubscribeToHeartBeat(_ *emptypb.Empty, stream pb.Internal_SubscribeToHeartBeatServer) error {
	err := make(chan error)
	return <-err
}

func (s *InternalServerType) SendOnlineUsers(ctx context.Context, msg *pb.UserStateWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// the state of a user logged into the server A changes

	fmt.Println("NONON: SendOnlineUsers")

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SendMessages(ctx context.Context, msg *pb.TextMessageWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// a new message is sent to server A DIRECTLY by the client
	// i.e any replicated messages will not be forwarded by the server A to other servers

	fmt.Println("NONON: SendMessages\n", msg)

	// Insert the new message to our database
	fmt.Println("dbpool: ", s.DBPool)
	fmt.Println("ctx: ", ctx)
	insertNewMessage(s.DBPool, ctx, msg.TextMessage, VectorClock{clocks: msg.Clock})

	// TODO: If their timestamp was more recent than ours, request updates from the relevant replicas (cry).

	return &emptypb.Empty{}, nil
}

func (s *InternalServerType) SendReactions(ctx context.Context, msg *pb.ReactionWithClock) (*emptypb.Empty, error) {
	// this method will be called by the server A after
	// a reaction update is sent to server A DIRECTLY by the client
	// i.e any replicated reaction update will not be forwarded by the server A to other servers

	fmt.Println("NONON: SendReactions")

	return &emptypb.Empty{}, nil
}
