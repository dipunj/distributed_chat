// contains network connection related code
package network

import (
	"chat/client/state"
	pb "chat/pb"
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// name of the container in the docker network
	DEFAULT_SERVER       = "chat_server"
	DEFAULT_PORT         = 3000
	DEFAULT_CONN_TIMEOUT = time.Second * 5
)

func parseServerFromArg(server_port string) string {
	// parse the server address and port from server_port
	// if server_port is empty, use the default server address and port
	addr := strings.Split(server_port, ":")

	server := addr[0]
	port := addr[1]

	server_ip := net.ParseIP(server)

	if server_ip == nil {
		log.Debug("Server is not an IP address, resolving hostname using DNS")
		server_ips, err := net.LookupIP(server)

		if err != nil {
			// dns lookup failed
			log.Info("DNS lookup failed", err)
		} else if len(server_ips) > 0 {
			server_ip = server_ips[len(server_ips)-1]
			log.Info("hostname ", server, " resolved to ", server_ip)
		} else {
			log.Info(errors.New("invalid hostname: " + server))
			log.Info("using default server address and port")
			server_ip = net.ParseIP(DEFAULT_SERVER)
			port = strconv.Itoa(DEFAULT_PORT)
		}
	}

	return server_ip.String() + ":" + port
}

func IsConnected() bool {
	return state.RpcConn != nil && state.ChatClient != nil
}

func EstablishConnection(serverAddress string) {

	parsedAddress := parseServerFromArg(serverAddress)
	log.Info("Connecting to chat server at ", parsedAddress)

	ctx, cancelContext := context.WithTimeout(context.Background(), DEFAULT_CONN_TIMEOUT)

	conn, err := grpc.DialContext(
		ctx,
		parsedAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	defer cancelContext()

	if err != nil {
		// raise error
		log.Info("Unable to dial to server ", err)
		return
	}

	client := pb.NewGroupChatClient(conn)

	if IsConnected() {
		state.RpcConn.Close()
		state.ChatClient = nil
	}

	state.RpcConn = conn
	state.ChatClient = client

	EstablishStream()
}

func ListenForUpdatesInBackground(conn pb.GroupChat_SubscribeClient) {
	defer state.Wait.Done()

	// loop infinitely
	for {
		msg, err := conn.Recv()

		if err != nil {
			GracefulQuit("Server is down")
			return
		}

		state.RecentMessages.Replace(msg.RecentMessages)
		state.Current_group_participants = msg.OnlineUserNames
		state.Rerender <- true
	}

}

func EstablishStream() {
	var err error
	stream, err := state.ChatClient.Subscribe(context.Background(), &emptypb.Empty{})

	if err != nil {
		log.Info("Failed to establish connection with server ")
	} else {
		state.Wait.Add(1)
		go ListenForUpdatesInBackground(stream)
	}
}
