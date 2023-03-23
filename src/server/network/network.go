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

func ServeRequests(serverAddress string) {
	l := getTCPListener(serverAddress)

	log.Info("Starting server at", serverAddress)

	Server.GrpcServer = grpc.NewServer()
	pb.RegisterGroupChatServer(Server.GrpcServer, &Server)

	// Serve() accepts each connection
	// and spawns a new goroutine for each new request
	Server.GrpcServer.Serve(l)
}

// func BroadCastToGroup() {

// }
