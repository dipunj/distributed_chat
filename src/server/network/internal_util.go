package network

import (
	"context"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"chat/pb"
	"chat/server/db"
)

func GetReplicaAddressFromID(replicaID int, port string) string {
	ip_prefix := "172.30.100.10"
	return ip_prefix + strconv.Itoa(replicaID) + ":" + port
}

func InitializeReplicas(replica_count int) {

	// populate replica_ids array with ids from 1 to replica_count, except selfID
	for i := 1; i <= replica_count; i++ {
		if i != SelfServerID {
			ReplicaIds = append(ReplicaIds, i)
			// initially we assume all replicas are offline
			ReplicaState[i] = &ReplicaStateType{
				Client:            nil,
				Changed:           make(chan bool),
				IsOnline:          false,
				PublicIpAddress:   GetReplicaAddressFromID(i, DEFAULT_PUBLIC_PORT),
				InternalIpAddress: GetReplicaAddressFromID(i, DEFAULT_INTERNAL_PORT),
			}
		}
	}

}

func GetNewerThan(vc VectorClock) ([]pb.TextMessageWithClock, []pb.ReactionWithClock) {
	// Query the database for any messages, reactions, etc. that are newer
	// than the given timestamp.

	new_messages := []pb.TextMessageWithClock{}
	new_reactions := []pb.ReactionWithClock{}

	query_str := `
		SELECT
			id,
			message_type,
			client_id,
			sender_name,
			group_name,
			content,
			client_sent_at,
			server_received_at,
			vector_ts,
			parent_msg_id
		FROM messages
		WHERE message_type = 'text' AND vector_ts[$1] > $2
	`

	var query_result struct {
		id                 int64
		message_type       string
		client_id          string
		sender_name        string
		group_name         string
		content            string
		client_sent_at     time.Time
		server_received_at time.Time
		vector_ts          []int64
		parent_msg_id      int
	}

	// There's probably a way to do this with a single query, but I'm not smart
	// enough to know what it is...
	for i := 1; i < len(vc); i++ {
		params := []interface{}{i, vc[i]}
		// Check for messages
		rows, err := db.DBPool.Query(context.Background(), query_str, params...)
		if err != nil {
			log.Fatal("Failed in heartbeat query: ", err)
		}

		for rows.Next() {
			rows.Scan(
				&query_result.id,
				&query_result.message_type,
				&query_result.client_id,
				&query_result.sender_name,
				&query_result.group_name,
				&query_result.content,
				&query_result.client_sent_at,
				&query_result.server_received_at,
				&query_result.vector_ts,
				&query_result.parent_msg_id,
			)

			if query_result.message_type == "text" {
				msg := pb.TextMessageWithClock{
					ClientId: query_result.client_id,
					TextMessage: &pb.TextMessage{
						Id:         &query_result.id,
						SenderName: query_result.sender_name,
						GroupName:  query_result.group_name,
						Content:    query_result.content,
						LikedBy:    0, // TODO: This isn't right.
						//						ClientSentAt: query_result.client_sent_at,
						//						ServerReceivedAt: query_result.server_received_at,
					},
					Clock: &pb.Clock{Clock: query_result.vector_ts},
				}

				new_messages = append(new_messages, msg)

			} else if query_result.message_type == "reaction" {

			} else {
				log.Error("Got unknown message type in query ({})", query_result.message_type)
			}
		}
	}

	return new_messages, new_reactions
}
