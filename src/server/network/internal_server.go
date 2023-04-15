package network

import (
	"chat/pb"
	"chat/server/db"
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

// we server requests to replicas on a different port
// This keeps the client-server and replica-replica communication separate

func (s *InternalServerType) SubscribeToHeartBeat(_ *emptypb.Empty, stream pb.Internal_SubscribeToHeartBeatServer) error {
	err := make(chan error)
	return <-err
}

func (s *InternalServerType) CreateNewMessage(ctx context.Context, msg_w_clock *pb.TextMessageWithClock) (*pb.Status, error) {
	log.Debug("[CreateNewMessage] (internal server) called")
	// this method will be called by the server A after
	// a new message is sent to server A DIRECTLY by the client

	// i.e any replicated messages will not be forwarded by the server A to other servers")
	msg := msg_w_clock.TextMessage
	msgTimestamp := msg_w_clock.Clock
	client_id := msg_w_clock.ClientId

	Clock.UpdateFrom(msgTimestamp)
	_, err := insertNewMessage(client_id, msg, msgTimestamp)

	if err == nil {
		log.Info("[CreateNewMessage] from replica",
			client_id,
			" sender name ",
			msg.SenderName)

		defer broadcastGroupUpdatesToImmediateMembers(msg.GroupName, PublicServer.Subscribers)

		return &pb.Status{Status: true}, err

	} else {
		log.Error("[CreateNewMessage] from replica",
			client_id,
			" sender name ",
			msg.SenderName,
			" error",
			err.Error())

		return &pb.Status{Status: false}, err
	}

}

func (s *InternalServerType) UpdateReaction(ctx context.Context, msg_w_clock *pb.ReactionWithClock) (*pb.Status, error) {
	log.Debug("[UpdateReaction] (internal server) called")

	// this method will be called by the server A after
	// a reaction update is sent to server A DIRECTLY by the client
	// i.e any replicated reaction update will not be forwarded by the server A to other servers

	msg := msg_w_clock.Reaction
	client_id := msg_w_clock.ClientId
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	_, err := insertNewReaction(client_id, msg, msgTimestamp)

	if err == nil {
		log.Info("[UpdateReaction] from replica for user with IP at replica",
			"client_id",
			client_id,
		)
		defer broadcastGroupUpdatesToImmediateMembers(msg.GroupName, PublicServer.Subscribers)
		return &pb.Status{Status: true}, err
	}

	return &pb.Status{Status: false}, err
}

func (s *InternalServerType) SwitchUser(ctx context.Context, msg_w_clock *pb.UserStateWithClock) (*pb.Status, error) {
	log.Debug("[SwitchUser] (internal server) called")
	// this method will be called by the server A after
	// a user switches user name

	user_ip := msg_w_clock.ClientId
	replica_id := int(msg_w_clock.ReplicaId)
	msg := msg_w_clock.UserState
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	handleSwitchUser(user_ip, replica_id, msg, &PublicServer.Subscribers, msgTimestamp)

	return &pb.Status{Status: true}, nil
}

func (s *InternalServerType) SwitchGroup(ctx context.Context, msg_w_clock *pb.UserStateWithClock) (*pb.Status, error) {
	log.Debug("[SwitchGroup] (internal server) called")
	// this method will be called by the server A after
	// a user switches from one group to another

	user_ip := msg_w_clock.ClientId
	replica_id := int(msg_w_clock.ReplicaId)
	msg := msg_w_clock.UserState
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	handleSwitchGroup(user_ip, replica_id, msg, &PublicServer.Subscribers, msgTimestamp)

	return &pb.Status{Status: true}, nil
}

func (s *InternalServerType) UserIsOffline(ctx context.Context, msg_w_clock *pb.ClientIdWithClock) (*pb.Status, error) {
	log.Debug("[UserIsOffline] (internal server) called")
	// this method will be called by the server A after
	// a user goes offline on server A

	user_ip := msg_w_clock.ClientId
	replica_id := int(msg_w_clock.ReplicaId)
	msgTimestamp := msg_w_clock.Clock

	Clock.UpdateFrom(msgTimestamp)
	handleUserIsOffline(user_ip, replica_id, &PublicServer.Subscribers, msgTimestamp)

	return &pb.Status{Status: true}, nil
}

func (s *InternalServerType) GetLatestClock(ctx context.Context, _ *emptypb.Empty) (*pb.Clock, error) {
	log.Debug("[GetLatestClock] (internal server) called")

	// TODO: write query to get the latest clock from the database
	latest_clock_query := `
		SELECT vector_ts
		FROM messages 
		WHERE
			vector_ts[$1] = (
				SELECT MAX(vector_ts[$1])
				FROM messages
			)
		ORDER BY server_received_at DESC
		LIMIT 1
		`

	var latestClk []int64

	row := db.DBPool.QueryRow(context.Background(), latest_clock_query, SelfServerID+1)

	err := row.Scan(&latestClk)
	if err != nil {
		log.Error("[GetLatestClock]", err.Error(), "Returning Clock: ", Clock)
		return &pb.Clock{Clock: Clock}, nil
	}

	return &pb.Clock{Clock: latestClk}, nil
}

func (s *InternalServerType) PushDBMessages(ctx context.Context, msg *pb.DBMessages) (*pb.Status, error) {
	log.Debug("[PushDBMessages] (internal server) called")

	tx, err := db.DBPool.Begin(ctx)
	if err != nil {
		log.Error("[PushDBMessages] error in begin transaction", err.Error())
		return &pb.Status{Status: false}, err
	}

	batch := &pgx.Batch{}

	messages := msg.Messages
	insert_query := `INSERT INTO messages (
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
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			) ON CONFLICT (vector_ts) DO NOTHING`

	for _, message := range messages {

		parentMsgID := sql.NullString{}
		if message.ParentMsgId != "" {
			parentMsgID = sql.NullString{String: message.ParentMsgId, Valid: true}
		}

		batch.Queue(insert_query,
			message.Id,
			message.MessageType,
			message.ClientId,
			message.SenderName,
			message.GroupName,
			message.Content,
			parentMsgID,
			message.VectorTs,
			message.ClientSentAt.AsTime(),
			message.ServerReceivedAt.AsTime())
	}

	results := db.DBPool.SendBatch(context.Background(), batch)

	if err := results.Close(); err != nil {
		log.Error("Failed to execute batch: %v\n", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		log.Error("Unable to commit transaction: %v\n", err)
	}

	return &pb.Status{Status: true}, nil
}
