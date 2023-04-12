package network

import (
	"chat/pb"
	"chat/server/db"
	"context"
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const REPLICA_CONNECTION_TIMEOUT = 5 * time.Second

func ListenHeartBeat(state *ReplicaStateType, replicaId int) {

	stream, err := state.Client.SubscribeToHeartBeat(context.Background(), &emptypb.Empty{})

	if err != nil {
		log.Debug("An error occurred in streaming for replica id", replicaId, err.Error())
	} else {
		// loop infinitely
		for {
			log.Debug("Waiting for connection state to change from replica", replicaId)

			// the following call is blocking
			_, err := stream.Recv()

			if err != nil {
				code := status.Code(err)

				if code == codes.Unavailable {
					// connection was closed because the server is unavailable
					log.Debug("Connection was closed because the server is unavailable code = ", code)
				} else if code == codes.Canceled {
					// connection was closed by the client
					log.Debug("Connection was closed by the client code = ", code)
				} else {
					log.Debug("An error occurred in streaming", err.Error())
				}
				// break because there is an error with the connection
				break
			}
		}
	}

	state.IsOnline = false
}

// this function dials to the replica and
// hooks onto the replica using a stream
// when the stream is closed, it will retry to connect to the replica

func DialAndMonitor(replicaId int) {

	state := ReplicaState[replicaId]
	ip_address := state.InternalIpAddress

	// this loop keeps on retrying to connect to the replica
	// once the connection is established

	for {
		ctx, _ := context.WithTimeout(context.Background(), REPLICA_CONNECTION_TIMEOUT)

		conn, err := grpc.DialContext(
			ctx,
			ip_address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)

		if err == nil {
			log.Info("[DialAndPing] TCP Dial successful: Replica ID ", replicaId, " at ", ip_address)

			// create a new client on this connection
			state.Client = pb.NewInternalClient(conn)

			SyncReplica(replicaId)

			state.IsOnline = true

			// the following call is blocking
			ListenHeartBeat(state, replicaId)

			if !state.IsOnline {
				// if the replica is offline, close the connection
				conn.Close()
			}
		} else {
			log.Error("[DialAndPing] TCP Dial failed: Replica ID ", replicaId, " at ", ip_address, " with error ", err)
		}

		// retry after 1 second
		time.Sleep(1 * time.Second)
	}
}

func ConnectToReplicas() {
	// this method will start REPLICA_COUNT - 1 goroutines to connect to each of the replicas

	// each go routine will run indefinitely
	// the go routine will ping the replica to check if it is online
	// if the replica is online, it will set the replica state to online
	// if the replica is offline, it will set the replica state to offline and call grpc dial every 1 second

	for _, replicaId := range ReplicaIds {
		go DialAndMonitor(replicaId)
	}

}

func SyncReplica(replicaId int) {
	// this function will sync the replica with the current state of the server

	// 1. request replica's latest message's clock
	// 2. find all the messages happened after the replica's latest message and be
	// 3. send the messages to the replica
	log.Debug("Syncing replica called for", replicaId)

	state := ReplicaState[replicaId]
	client := state.Client

	log.Debug("Syncing replica called for", replicaId)
	replica_clock, err := client.GetLatestClock(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Error("Error while couldn't sync the latest clock from replica", replicaId, err)
		return
	}
	log.Debug("Latest clock from replica", replicaId, " is ", replica_clock.Clock)
	messages := GetMessagesAfter(replica_clock.Clock)

	log.Debug("Found", len(messages), "messages after the clock", replica_clock, "while syncing replica", replicaId)

	client.PushDBMessages(context.Background(), &pb.DBMessages{Messages: messages})

}

func GetMessagesAfter(timestamp []int64) []*pb.DBSchema {
	// this function will return all the messages after the clock

	messages := make([]*pb.DBSchema, 0)

	// postgres array indexes start from 1 not from 0
	// so the first clock value at index 2
	get_after_query := `SELECT
		id,
		message_type,
		client_id,
		sender_name,
		group_name,
		content,
		parent_msg_id,
		vector_ts,
		client_sent_at,
		server_received_at
	FROM
		messages
	WHERE 
		vector_ts[2] >= $1 AND
		vector_ts[3] >= $2 AND
		vector_ts[4] >= $3 AND
		vector_ts[5] >= $4 AND
		vector_ts[6] >= $5 
	`

	rows, err := db.DBPool.Query(context.Background(), get_after_query, timestamp[1], timestamp[2], timestamp[3], timestamp[4], timestamp[5])

	if err != nil {
		log.Error("[GetMessagesAfter]: Error while querying the database", err)
	}

	for rows.Next() {
		var row = &pb.DBSchema{}
		var parentMsgId sql.NullString
		var clientSentAt time.Time
		var serverReceivedAt time.Time

		err := rows.Scan(
			&row.Id,
			&row.MessageType,
			&row.ClientId,
			&row.SenderName,
			&row.GroupName,
			&row.Content,
			&parentMsgId,
			&row.VectorTs,
			&clientSentAt,
			&serverReceivedAt,
		)

		if err != nil {
			log.Error("Error while scanning row", err)
		}

		if parentMsgId.Valid {
			row.ParentMsgId = parentMsgId.String
		}

		row.ClientSentAt = timestamppb.New(clientSentAt)
		row.ServerReceivedAt = timestamppb.New(serverReceivedAt)

		messages = append(messages, row)
	}

	return messages
}
