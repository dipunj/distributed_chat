package network

import (
	pb "chat/pb"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func getTCPListener(serverAddress string) net.Listener {
	l, err := net.Listen("tcp", serverAddress)
	if err != nil {
		// raise error
		log.Fatal("TCP Listen Error", err)
	} else {
		log.Info("TCP Listen Successful")
	}

	return l
}

func ServeRequestsToClients(serverAddress string) {
	l := getTCPListener(serverAddress)

	log.Info("Starting client service at", serverAddress)

	PublicServer.GrpcServer = grpc.NewServer()
	pb.RegisterPublicServer(PublicServer.GrpcServer, &PublicServer)

	// Serve() accepts each connection
	// and spawns a new goroutine for each new request
	PublicServer.GrpcServer.Serve(l)
}

// we server requests to replicas on a different port
// to keep the client-server and replica-replica
// communication separate

func ServerRequestsToReplicas(serverAddress string) {

	log.Info("Starting replication service at", serverAddress)

	l := getTCPListener(serverAddress)
	InternalServer.GrpcServer = grpc.NewServer()
	pb.RegisterInternalServer(PublicServer.GrpcServer, &InternalServer)

	// Serve() accepts each connection
	// and spawns a new goroutine for each new request
	InternalServer.GrpcServer.Serve(l)
}
