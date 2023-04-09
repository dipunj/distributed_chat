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
		log.Fatal("Cannot start server. Reason: TCP Listen Error. Error:", err)
	} else {
		log.Info("TCP Listen Successful")
	}

	return l
}

func ServePublicRequests(ip_address string) {
	log.Info("Starting public(client) service at", ip_address)

	PublicServer.GrpcServer = grpc.NewServer()
	pb.RegisterPublicServer(PublicServer.GrpcServer, &PublicServer)

	// Serve() spawns a new goroutine for each new request
	l := getTCPListener(ip_address)
	err := PublicServer.GrpcServer.Serve(l)

	if err != nil {
		log.Fatalf("Error while starting the Public server on the %s listen address %v", l, err.Error())
	} else {
		log.Info("Public Server started")
	}
}

// we server requests to replicas on a different port
// This keeps the client-server and replica-replica communication separate
func ServeInternalRequests(ip_address string) {
	log.Info("Starting internal (replication) service at", ip_address)

	InternalServer.GrpcServer = grpc.NewServer()

	pb.RegisterInternalServer(InternalServer.GrpcServer, &InternalServerType{})

	// Serve() spawns a new goroutine under the hood for each new request
	l := getTCPListener(ip_address)
	err := InternalServer.GrpcServer.Serve(l)

	if err != nil {
		log.Fatalf("Error while starting the gRPC server on the %s listen address %v", l, err.Error())
	} else {
		log.Info("Internal Server started")
	}
}
