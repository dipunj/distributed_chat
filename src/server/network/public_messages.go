package network

import (
	pb "chat/pb"
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PublicServerType) CreateNewMessage(ctx context.Context, msg *pb.TextMessage) (*pb.Status, error) {

	var new_message_query string = `
		INSERT INTO messages (
			message_type,
			client_id,
			sender_name,
			group_name,
			content,
			vector_timestamp,
			client_sent_at,
			server_received_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING 
			id, sender_name, group_name, content, client_sent_at, server_received_at
	`
	server_received_at := time.Now()
	var row pb.TextMessage

	var vector_ts_str = CurrentTimestamp.Increment(0).ToDbFormat()

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	params := []interface{}{
		"text",
		client_id,
		msg.SenderName,
		msg.GroupName,
		msg.Content,
		vector_ts_str,
		msg.ClientSentAt.AsTime(),
		server_received_at,
	}

	var clientSentAt time.Time
	var serverReceivedAt time.Time
	err := s.DBPool.QueryRow(context.Background(),
		new_message_query,
		params...,
	).Scan(&row.Id, &row.SenderName, &row.GroupName, &row.Content, &vector_ts_str, &clientSentAt, &serverReceivedAt)

	if err == nil {
		row.ServerReceivedAt = timestamppb.New(serverReceivedAt)
		row.ClientSentAt = timestamppb.New(clientSentAt)

		log.Info("[CreateNewMessage] for ", client_id, " with user name", msg.SenderName)

		defer s.broadcastUpdates(msg.GroupName)

		return &pb.Status{Status: true}, err
	} else {
		log.Error("[CreateNewMessage] for ", client_id, " with user name", msg.SenderName, err)
		return &pb.Status{Status: false}, err
	}
}

func (s *PublicServerType) broadcastUpdates(group_name string) {

	wait := sync.WaitGroup{}
	done := make(chan int)

	online_users := s.getOnlineUsers(group_name)
	recent_messages := s.getRecentMessages(group_name)

	group_update := &pb.GroupDetails{
		Status:          true,
		RecentMessages:  recent_messages,
		OnlineUserNames: online_users,
	}

	for _, conn := range s.Subscribers {
		if conn.group_name == group_name {
			// online users in the same group
			wait.Add(1)
			go func(msg *pb.GroupDetails, groupMember *ResponseStream) {

				defer wait.Done()

				if groupMember.is_online {
					err := groupMember.stream.Send(msg)
					if err != nil {

						groupMember.is_online = false
						groupMember.error <- err
					}
				}
			}(group_update, conn)
		}
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}

func (s *PublicServerType) addLike(client_id string, msg *pb.Reaction) (interface{}, error) {
	var add_reaction_query string = `
		INSERT INTO messages (
			message_type,
			client_id,
			sender_name,
			group_name,
			content,
			parent_msg_id,
			client_sent_at,
			server_received_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) 
	`

	log.Info("[!!!!!!!]")

	server_received_at := time.Now()

	params := []interface{}{
		"reaction",
		client_id,
		msg.SenderName,
		msg.GroupName,
		"like",
		msg.OnMessageId,
		msg.ClientSentAt.AsTime(),
		server_received_at,
	}

	var row interface{}
	err := s.DBPool.QueryRow(context.Background(),
		add_reaction_query,
		params...,
	).Scan(&row)

	return row, err
}

func (s *PublicServerType) removeLike(msg *pb.Reaction) (interface{}, error) {

	var remove_reaction_query string = `
			DELETE FROM messages
			WHERE 
				message_type = $1
				and parent_msg_id = $2
				and sender_name = $3 
				and group_name = $4
				and content = $5
			`

	params := []interface{}{
		"reaction",
		msg.OnMessageId,
		msg.SenderName,
		msg.GroupName,
		"like",
	}

	var row interface{}
	err := s.DBPool.QueryRow(context.Background(),
		remove_reaction_query,
		params...,
	).Scan(&row)

	return row, err
}

func (s *PublicServerType) UpdateReaction(ctx context.Context, msg *pb.Reaction) (*pb.Status, error) {

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	log.Info("[UpdateReaction] from",
		client_id,
		" with user name",
		msg.SenderName,
		" reaction",
		msg.Content,
		" on message",
		msg.OnMessageId,
	)

	var err error
	if msg.Content == "like" {
		_, err = s.addLike(client_id, msg)
	} else {
		_, err = s.removeLike(msg)
	}

	if err != nil && err != pgx.ErrNoRows {
		log.Error(err)
	} else {
		defer s.broadcastUpdates(msg.GroupName)
	}

	return &pb.Status{Status: true}, nil
}

func (s *PublicServerType) PrintGroupHistory(ctx context.Context, msg *pb.GroupName) (*pb.GroupHistory, error) {
	group_name := msg.GroupName
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
		ORDER BY text_message.client_sent_at
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

	defer rows.Close()
	return &pb.GroupHistory{Messages: recent_messages}, err
}

func (s *PublicServerType) Subscribe(_ *emptypb.Empty, stream pb.GroupChat_SubscribeServer) error {
	ctx := stream.Context()
	client, _ := peer.FromContext(ctx)
	clientID := client.Addr.String()

	rs := &ResponseStream{
		stream:     stream,
		client_id:  clientID,
		user_name:  "",
		group_name: "",
		is_online:  true,
		error:      make(chan error),
	}

	s.Subscribers[clientID] = rs

	// Check if the client is offline/disconnected
	go func() {
		<-ctx.Done()

		// Update the isOnline field of the ResponseStream to false
		rs.is_online = false
		s.broadcastUpdates(rs.group_name)

		// Remove the ResponseStream from the Subscribers map
		delete(s.Subscribers, clientID)
	}()

	return <-rs.error
}
