package network

import (
	pb "chat/pb"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const NUM_REPLICAS = 5

const (
	// listen on all interfaces
	DEFAULT_INTERFACE     = "0.0.0.0"
	DEFAULT_PUBLIC_PORT   = "12000"
	DEFAULT_INTERNAL_PORT = "11000"

	PUBLIC_ADDRESS   = DEFAULT_INTERFACE + ":" + DEFAULT_PUBLIC_PORT
	INTERNAL_ADDRESS = DEFAULT_INTERFACE + ":" + DEFAULT_INTERNAL_PORT
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

func ServePublicRequests() {

	log.Info("Starting public(client) service at", PUBLIC_ADDRESS)

	PublicServer.GrpcServer = grpc.NewServer()
	pb.RegisterPublicServer(PublicServer.GrpcServer, &PublicServer)

	// Serve() spawns a new goroutine for each new request
	l := getTCPListener(PUBLIC_ADDRESS)
	err := PublicServer.GrpcServer.Serve(l)

	if err != nil {
		log.Fatalf("Error while starting the Public server on the %s listen address %v", l, err.Error())
	} else {
		log.Info("Public Server started")
	}
}

// we server requests to replicas on a different port
// This keeps the client-server and replica-replica communication separate
func ServeInternalRequests() {

	log.Info("Starting internal (replication) service at", INTERNAL_ADDRESS)

	InternalServer.GrpcServer = grpc.NewServer()

	pb.RegisterInternalServer(InternalServer.GrpcServer, &InternalServerType{})

	// Serve() spawns a new goroutine under the hood for each new request
	l := getTCPListener(INTERNAL_ADDRESS)
	err := InternalServer.GrpcServer.Serve(l)

	if err != nil {
		log.Fatalf("Error while starting the gRPC server on the %s listen address %v", l, err.Error())
	} else {
		log.Info("Internal Server started")
	}
}

var CurrentTimestamp = makeVectorClock(NUM_REPLICAS)
var ReplicaId int
