package network

import (
	pb "chat/pb"
	"chat/server/db"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PublicServerType) CreateNewMessage(ctx context.Context, msg *pb.TextMessage) (*pb.Status, error) {

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	// TODO: increment clock here?
	err := insertNewMessage(client_id, msg, Clock)

	if err != nil {
		log.Error("[CreateNewMessage] for ", client_id, " with user name ", msg.SenderName, err.Error())

		return &pb.Status{Status: false}, err
	} else {
		log.Info("[CreateNewMessage]:Success for ", client_id, " with user name ", msg.SenderName)

		log.Debug("[CreateNewMessage] notifying replicas about new message from ", client_id, " with user name ", msg.SenderName)
		go notifyNewMessageToReplica(client_id, ctx, msg)
		defer broadcastGroupUpdatesToImmediateMembers(msg.GroupName, s.Subscribers)

		return &pb.Status{Status: true}, err
	}
}

func (s *PublicServerType) UpdateReaction(ctx context.Context, msg *pb.Reaction) (*pb.Status, error) {

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	// TODO: increment clock here?
	err := insertNewReaction(client_id, msg, Clock)

	if err != nil && err != pgx.ErrNoRows {
		log.Error("[UpdateReaction] for ", client_id, " with user name ", msg.SenderName, err.Error())

		return &pb.Status{Status: false}, nil
	} else {
		log.Info("[UpdateReaction]:Success for ", client_id, " with user name ", msg.SenderName)

		log.Debug("[UpdateReaction] notifying replicas about reaction update for ", client_id, " with user name ", msg.SenderName)
		go notifyReactionUpdateToReplica(client_id, ctx, msg)
		defer broadcastGroupUpdatesToImmediateMembers(msg.GroupName, s.Subscribers)

		return &pb.Status{Status: true}, nil
	}

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

	defer rows.Close()
	return &pb.GroupHistory{Messages: recent_messages}, err
}

func (s *PublicServerType) Subscribe(_ *emptypb.Empty, stream pb.Public_SubscribeServer) error {
	ctx := stream.Context()
	client, _ := peer.FromContext(ctx)
	clientID := client.Addr.String()

	rs := &ResponseStream{
		server_id:  SelfServerID,
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
		broadcastGroupUpdatesToImmediateMembers(rs.group_name, s.Subscribers)

		// Remove the ResponseStream from the Subscribers map
		delete(s.Subscribers, clientID)
	}()

	return <-rs.error
}

// handles the v command
func (s *PublicServerType) VisibleReplicas(ctx context.Context, msg *emptypb.Empty) (*pb.VisibilityResponse, error) {
	response := &pb.VisibilityResponse{}
	for k, replica := range ReplicaState {
		response.Replicas = append(response.Replicas, &pb.ReplicaDetail{
			Id:        int32(k),
			IsOnline:  replica.IsOnline,
			IpAddress: replica.PublicIpAddress,
		})
	}

	return response, nil
}

func (s *PublicServerType) SwitchUser(ctx context.Context, msg *pb.UserState) (*pb.Status, error) {

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	// TODO: update clock?
	handleSwitchUser(client_id, SelfServerID, msg, &s.Subscribers, Clock)
	notifyReplicaAboutUserSwitch(client_id, ctx, msg)

	response := pb.Status{Status: true}
	return &response, nil

}

func (s *PublicServerType) SwitchGroup(ctx context.Context, msg *pb.UserState) (*pb.GroupDetails, error) {

	client, _ := peer.FromContext(ctx)
	client_id := client.Addr.String()

	// TODO: update clock?
	handleSwitchGroup(client_id, SelfServerID, msg, &s.Subscribers, Clock)
	notifyReplicaAboutGroupSwitch(client_id, ctx, msg)

	response := pb.GroupDetails{Status: true}

	return &response, nil
}
