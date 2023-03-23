// contains network connection related code
package network

import (
	"chat/client/state"
	pb "chat/pb"
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func EstablishConnection(serverAddress string) {
	log.Info("Connecting to chat server at ", serverAddress)

	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		// raise error
		log.Fatal("Unable to dial to server - %s", err)
	}

	state.RpcConn = conn

	client := pb.NewGroupChatClient(conn)
	state.ChatClient = client
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
		log.Fatal("Failed to establish connection with server ")
	} else {
		state.Wait.Add(1)
		go ListenForUpdatesInBackground(stream)
	}
}
