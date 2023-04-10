package network

import (
	"chat/pb"
	"chat/server/db"
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func insertNewMessage(db *pgxpool.Pool, ctx context.Context, msg *pb.TextMessage, ts VectorClock) error {

	var new_message_query string = `
		INSERT INTO messages (
			message_type,
			client_id,
			sender_name,
			group_name,
			content,
			client_sent_at,
			server_received_at,
			vector_ts
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING 
			id, sender_name, group_name, content, client_sent_at, server_received_at, vector_ts
	`
	server_received_at := time.Now()
	var row pb.TextMessage
	var vector_ts []int

	var vector_ts_str = Clock.Increment().ToDbFormat()

	log.Info("Vector timestamp for Replica ", SelfID, " is ", vector_ts_str)

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	params := []interface{}{
		"text",
		client_id,
		msg.SenderName,
		msg.GroupName,
		msg.Content,
		msg.ClientSentAt.AsTime(),
		server_received_at,
		vector_ts_str,
	}

	var clientSentAt time.Time
	var serverReceivedAt time.Time
	err := db.QueryRow(context.Background(),
		new_message_query,
		params...,
	).Scan(&row.Id, &row.SenderName, &row.GroupName, &row.Content, &clientSentAt, &serverReceivedAt, &vector_ts)

	return err
}

func notifyReplicas(ctx context.Context, msg *pb.TextMessage) {
	wg := sync.WaitGroup{}
	msg_with_clock := &pb.TextMessageWithClock{
		TextMessage: msg,
		Clock:       Clock.clocks,
	}
	for _, replica := range ReplicaState {
		wg.Add(1)
		go replica.Client.CreateNewMessage(ctx, msg_with_clock)
	}

	wg.Wait()
}

// this function sends the latest view of the group
// to all the group members that are directly connected
// members connected to the replicas are handled by the replicas
// when a new message is received, that message is synced to the replica
// which triggers the broadcastGroupUpdatesToMembers on the replica automatically
func broadcastGroupUpdatesToImmediateMembers(group_name string, subscribers map[string]*ResponseStream) {

	wait := sync.WaitGroup{}
	done := make(chan int)

	online_users := getOnlineUsers(group_name, subscribers)
	recent_messages := getRecentMessages(group_name)

	// Broadcast to clients
	group_update := &pb.GroupDetails{
		Status:          true,
		RecentMessages:  recent_messages,
		OnlineUserNames: online_users,
	}

	for _, conn := range subscribers {
		// online users in the same group
		if conn.group_name == group_name && conn.server_id == SelfID {
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

	<-done
}

func getOnlineUsers(group_name string, subscribers map[string]*ResponseStream) []string {
	// go lang doesn't have inbuilt set implementation
	// following is a work around to get unique user names in a group

	online_users_set := make(map[string]bool)

	for _, conn := range subscribers {
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

func getRecentMessages(group_name string) []*pb.TextMessage {
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

	rows, err := db.DBPool.Query(context.Background(), group_recent_messages, group_name)

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
