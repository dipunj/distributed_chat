package network

import (
	pb "chat/pb"
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PublicServerType) SwitchUser(ctx context.Context, msg *pb.UserState) (*pb.Status, error) {

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	old_group_name := s.Subscribers[client_id].group_name

	s.Subscribers[client_id].group_name = ""
	s.Subscribers[client_id].user_name = *msg.UserName
	s.Subscribers[client_id].server_id = SelfID

	response := pb.Status{Status: true}

	if old_group_name != "" {
		// notify the old group that the user has left
		defer s.broadcastGroupUpdatesToImmediateMembers(old_group_name)
	}

	log.Info("Client [", client_id, "] has logged in with Username: ", *msg.UserName)

	return &response, nil

}

func (s *PublicServerType) SwitchGroup(ctx context.Context, msg *pb.UserState) (*pb.GroupDetails, error) {

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	log.Info("Current Subscribers")

	old_group_name := s.Subscribers[client_id].group_name
	s.Subscribers[client_id].group_name = *msg.GroupName
	s.Subscribers[client_id].user_name = *msg.UserName
	response := pb.GroupDetails{Status: true}

	log.Info("Client [", client_id, "], username", *msg.UserName, "has switch from group: ", old_group_name, " to ", *msg.GroupName)

	if old_group_name != "" {
		// notify the old group that the user has left their group
		defer s.broadcastGroupUpdatesToImmediateMembers(old_group_name)
	}

	// notify the new group that a user has joined the chat
	defer s.broadcastGroupUpdatesToImmediateMembers(*msg.GroupName)

	return &response, nil
}

func (s *PublicServerType) getRecentMessages(group_name string) []*pb.TextMessage {
	var group_recent_messages string = `
		SELECT 
			text_message.id,
			text_message.sender_name,
			text_message.group_name,    
			text_message.content,
			text_message.client_sent_at,
			text_message.server_received_at,
			count(like_message.sender_name) as liked_by
		FROM
			messages text_message
		LEFT JOIN
			messages like_message on like_message.parent_msg_id = text_message.id
		WHERE
			text_message.group_name = $1
			and text_message.message_type = 'text'
			and (like_message.message_type = 'reaction' OR like_message.message_type IS NULL)
			and (like_message.content = 'like' OR like_message.content IS NULL)
		GROUP BY text_message.id
		ORDER BY text_message.client_sent_at DESC
		LIMIT 10
	`
	recent_messages := []*pb.TextMessage{}

	rows, err := s.DBPool.Query(context.Background(), group_recent_messages, group_name)

	if err != nil {
		// handle error
		log.Error("Recent Messages 1", err)
	}

	for rows.Next() {
		var message struct {
			id                 int64
			sender_name        string
			group_name         string
			content            string
			liked_by           int64
			client_sent_at     time.Time
			server_received_at time.Time
		}

		err := rows.Scan(&message.id,
			&message.sender_name,
			&message.group_name,
			&message.content,
			&message.client_sent_at,
			&message.server_received_at,
			&message.liked_by)

		if err != nil {
			// handle error
			log.Error("Recent Messages 2", err)
		}

		recent_messages = append(recent_messages, &pb.TextMessage{
			Id:               &message.id,
			SenderName:       message.sender_name,
			GroupName:        message.group_name,
			Content:          message.content,
			LikedBy:          message.liked_by,
			ClientSentAt:     timestamppb.New(message.client_sent_at),
			ServerReceivedAt: timestamppb.New(message.server_received_at),
		})
	}

	// reverse the messages
	for i, j := 0, len(recent_messages)-1; i < j; i, j = i+1, j-1 {
		recent_messages[i], recent_messages[j] = recent_messages[j], recent_messages[i] //reverse the slice
	}

	defer rows.Close()
	return recent_messages
}

func (s *PublicServerType) getOnlineUsers(group_name string) []string {
	// go lang doesn't have inbuilt set implementation
	// following is a work around to get unique user names in a group

	online_users_set := make(map[string]bool)

	for _, conn := range s.Subscribers {
		if conn.group_name == group_name && conn.is_online {
			online_users_set[conn.user_name] = true
		}
	}

	online_users := make([]string, len(online_users_set))

	i := 0
	for k := range online_users_set {
		online_users[i] = k
		i++
	}

	return online_users
}
