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
		log.Fatal("Couldn't obtain TCP Listener. Error:", err)
	} else {
		log.Info("Obtained TCP Listener Successfully")
	}

	return l
}

func ServePublicRequests() {

	log.Info("[ServePublicRequests]: Starting public (client) service at", PUBLIC_ADDRESS)

	PublicServer.GrpcServer = grpc.NewServer()
	pb.RegisterPublicServer(PublicServer.GrpcServer, &PublicServer)

	// Serve() spawns a new goroutine for each new request
	l := getTCPListener(PUBLIC_ADDRESS)
	err := PublicServer.GrpcServer.Serve(l)

	if err != nil {
		log.Fatalf("[ServePublicRequests]: Error while starting the Public server on the %s listen address %v", l, err.Error())
	} else {
		log.Info("[ServePublicRequests]: Public Server started")
	}
}
