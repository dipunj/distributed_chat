// contains application logic related code

package network

import (
	"chat/client/state"
	"context"
	"fmt"

	pb "chat/pb"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SendMessage(content string) bool {
	if state.Current_user == "" {
		log.Error("Please first log in as a user")
		return false
	}
	if state.Current_group == "" {
		log.Error("Please first choose a group to send the message to")
		return false
	}

	client_sent_at := timestamppb.Now()

	message := &pb.TextMessage{
		SenderName:   state.Current_user,
		GroupName:    state.Current_group,
		Content:      content,
		ClientSentAt: client_sent_at,
	}

	_, err := state.ChatClient.CreateNewMessage(context.Background(), message)

	return err == nil
}

func SendReaction(reaction string, message_id int64) bool {
	if state.Current_user == "" {
		log.Error("Please first log in as a user")
		return false
	}
	if state.Current_group == "" {
		log.Error("Please first choose a group to send the message to")
		return false
	}

	client_sent_at := timestamppb.Now()

	message := &pb.Reaction{
		SenderName:   state.Current_user,
		GroupName:    state.Current_group,
		Content:      reaction,
		ClientSentAt: client_sent_at,
		OnMessageId:  message_id,
	}

	_, err := state.ChatClient.UpdateReaction(context.Background(), message)

	return err == nil
}

func PrintGroupHistory() {

	if state.Current_group == "" {
		log.Error("Please first choose a group to")
		return
	}

	params := pb.GroupName{GroupName: state.Current_group}
	reply, err := state.ChatClient.PrintGroupHistory(context.Background(), &params)

	if err == nil {
		state.RenderMu.Lock()
		// print reply
		for line_no, m := range reply.Messages {
			left := fmt.Sprintf("\t%d. [%s] %s", line_no+1, m.SenderName, m.Content)
			right := fmt.Sprintf("likes: %d", m.LikedBy)
			fmt.Printf("%s %s\n", left, right)
		}
		state.RenderMu.Unlock()
	} else {
		log.Error("Couldn't switch user")
	}

}

func RequestUserSwitchTo(new_user_name string) bool {

	params := pb.UserState{UserName: &new_user_name}
	_, err := state.ChatClient.SwitchUser(context.Background(), &params)

	if err == nil {
		// message has been successfully delivered to the server
		// update the recentMessage list
		// log.Info(reply)
		return true
	} else {
		log.Error("Couldn't switch user")
		return false
	}

}

func RequestGroupSwitchTo(new_group_name string) bool {

	params := pb.UserState{GroupName: &new_group_name, UserName: &state.Current_user}
	_, err := state.ChatClient.SwitchGroup(context.Background(), &params)

	if err == nil {
		// message has been successfully delivered to the server
		// update the recentMessage list
		// log.Info(reply)
		return true
	} else {
		log.Error("Couldn't switch group")
		return false
	}
}

func PrintServerView() {
	response, err := state.ChatClient.VisibleReplicas(context.Background(), &emptypb.Empty{})

	if err == nil {
		state.RenderMu.Lock()
		// print reply
		fmt.Printf("\tServer ID. [IP4 addr] [Status]\n")
		for _, replica := range response.Replicas {
			id := replica.Id
			ip_addr := replica.IpAddress
			status := replica.IsOnline

			status_text := ""
			if status {
				status_text = "online"
			} else {
				status_text = "offline"
			}

			fmt.Printf("\t(%d) %s [%s]\n", id, ip_addr, status_text)
		}
		state.RenderMu.Unlock()

	} else {
		log.Error("Couldn't get server view")
	}

}

func GracefulDisconnect(reason string) {

	state.Current_group = ""
	state.Current_user = ""
	state.Current_group_participants = []string{}
	state.RecentMessages.ClearAll()

	if !IsConnected() {
		return
	}

	if reason != "" {
		log.Info("Disconnecting from server. Reason: ", reason)
	}

	// close the stream
	_, cancel := context.WithCancel(context.Background())
	cancel()

	// reset the chat client
	state.ChatClient = nil

	// close the rpc connection
	state.RpcConn.Close()

	state.Rerender <- true
}
