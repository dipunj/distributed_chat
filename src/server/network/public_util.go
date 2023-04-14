package network

import (
	"chat/pb"
	"chat/server/db"
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// inserts new message into database
// with vector timestamp ts
func insertNewMessage(client_id string, msg *pb.TextMessage, ts VectorClock) (string, error) {

	log.Debug("[insertNewMessage] Inserting new message into database")

	server_received_at := time.Now()

	new_message_query := `
		INSERT INTO messages (
			id,
			message_type,
			client_id,
			sender_name,
			group_name,
			content,
			client_sent_at,
			server_received_at,
			vector_ts
		) VALUES (
			coalesce($1, uuid_generate_v4()),
			$2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING 
			id, sender_name, group_name, content, client_sent_at, server_received_at, vector_ts
	`

	params := []interface{}{
		msg.Id,
		"text",
		client_id,
		msg.SenderName,
		msg.GroupName,
		msg.Content,
		msg.ClientSentAt.AsTime(),
		server_received_at,
		ts,
	}

	var row pb.TextMessage

	var clientSentAt time.Time
	var serverReceivedAt time.Time
	var vector_ts []int

	err := db.DBPool.QueryRow(context.Background(),
		new_message_query,
		params...,
	).Scan(&row.Id, &row.SenderName, &row.GroupName, &row.Content, &clientSentAt, &serverReceivedAt, &vector_ts)

	return *row.Id, err
}

func insertNewReaction(client_id string, msg *pb.Reaction, ts VectorClock) (string, error) {

	log.Debug("[insertNewReaction] Inserting new reaction into database")

	server_received_at := time.Now()

	var update_reaction_query string = `
		INSERT INTO messages (
			id,
			message_type,
			client_id,
			sender_name,
			group_name,
			content,
			parent_msg_id,
			client_sent_at,
			server_received_at,
			vector_ts
		) VALUES (
			coalesce($1, uuid_generate_v4()),
			$2, $3, $4, $5, $6, $7, $8, $9, $10
		) 
		ON CONFLICT (message_type, parent_msg_id, sender_name)
			DO UPDATE SET content = $5
		RETURNING
			id
	`

	params := []interface{}{
		msg.Id,
		"reaction",
		client_id,
		msg.SenderName,
		msg.GroupName,
		msg.Content, // either "like" or "unlike" or ""
		msg.OnMessageId,
		msg.ClientSentAt.AsTime(),
		server_received_at,
		ts,
	}

	var rowId string
	err := db.DBPool.QueryRow(context.Background(),
		update_reaction_query,
		params...,
	).Scan(&rowId)

	return rowId, err
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
		if conn.group_name == group_name && conn.server_id == SelfServerID {
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

func markReplicaUsersOffline(replica_id int) {
	// go lang doesn't have inbuilt set implementation
	// following is a work around to get unique user names in a group

	subscribers := &PublicServer.Subscribers

	log.Debug("Updating user list in all groups about replica id", replica_id, "going offline")

	groups := make(map[string]bool)

	for _, conn := range *subscribers {
		if conn.server_id == replica_id {
			conn.is_online = false
			groups[conn.group_name] = true
		}
	}

	for group_name := range groups {
		// notify the old group that the user has left
		defer broadcastGroupUpdatesToImmediateMembers(group_name, *subscribers)
	}
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
			id                 string
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

func handleSwitchUser(user_ip string, on_replica int, msg *pb.UserState, current_subscribers *map[string]*ResponseStream, clock VectorClock) {

	// check if user_ip is present in the current_subscribers map
	// if not, then add it
	if _, ok := (*current_subscribers)[user_ip]; !ok {
		(*current_subscribers)[user_ip] = &ResponseStream{
			server_id:  on_replica,
			stream:     nil,
			client_id:  user_ip,
			user_name:  *msg.UserName,
			group_name: "",
			is_online:  true,
			error:      make(chan error),
		}
	} else {

		old_group_name := (*current_subscribers)[user_ip].group_name

		(*current_subscribers)[user_ip].group_name = ""
		(*current_subscribers)[user_ip].user_name = *msg.UserName
		(*current_subscribers)[user_ip].server_id = on_replica

		if old_group_name != "" {
			// notify the old group that the user has left
			defer broadcastGroupUpdatesToImmediateMembers(old_group_name, *current_subscribers)
		}
	}

	log.Info("Client [", user_ip, "] has logged in with Username: ", *msg.UserName, "on replica ID: ", on_replica)
}

func handleSwitchGroup(user_ip string, on_replica int, msg *pb.UserState, current_subscribers *map[string]*ResponseStream, clock VectorClock) {

	// check if user_ip is present in the current_subscribers map
	// if not, then add it
	if _, ok := (*current_subscribers)[user_ip]; !ok {
		(*current_subscribers)[user_ip] = &ResponseStream{
			server_id:  on_replica,
			stream:     nil,
			client_id:  user_ip,
			user_name:  *msg.UserName,
			group_name: *msg.GroupName,
			is_online:  true,
			error:      make(chan error),
		}
	} else {

		old_group_name := (*current_subscribers)[user_ip].group_name

		(*current_subscribers)[user_ip].group_name = *msg.GroupName
		(*current_subscribers)[user_ip].server_id = on_replica

		if old_group_name != "" {
			// notify the old group that the user has left
			defer broadcastGroupUpdatesToImmediateMembers(old_group_name, *current_subscribers)
		}
	}

	// notify the new group that a user has joined the chat
	defer broadcastGroupUpdatesToImmediateMembers(*msg.GroupName, *current_subscribers)

	log.Info("Client [", user_ip, "] has switched to Group: ", *msg.GroupName, "on replica ID: ", on_replica)
}

func handleUserIsOffline(user_ip string, on_replica int, current_subscribers *map[string]*ResponseStream, clock VectorClock) {

	if _, ok := (*current_subscribers)[user_ip]; !ok {
		// if the user was not present in the map, then add it
		(*current_subscribers)[user_ip] = &ResponseStream{
			server_id:  on_replica,
			stream:     nil,
			client_id:  user_ip,
			user_name:  "",
			group_name: "",
			is_online:  false,
			error:      make(chan error),
		}
	} else {

		user_state := (*current_subscribers)[user_ip]
		user_state.server_id = on_replica
		user_state.is_online = false
		log.Debug("Updated user state: ", user_state)
	}

	old_group_name := (*current_subscribers)[user_ip].group_name

	if old_group_name != "" {
		// notify the old group that the user has left
		defer broadcastGroupUpdatesToImmediateMembers(old_group_name, *current_subscribers)
	}

	log.Info("Client [", user_ip, "] has gone offline on replica ID: ", on_replica)
}
