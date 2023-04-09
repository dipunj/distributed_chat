package main

import (
	"chat/server/db"
	"chat/server/network"
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
)

// I think this is dead code
const MAX_CHAR_PER_LINE = 80

const (
	// listen on all interfaces
	DEFAULT_INTERFACE     = "0.0.0.0"
	DEFAULT_PUBLIC_PORT   = "12000"
	DEFAULT_INTERNAL_PORT = "11000"
)

func loadSavedTimestamp() network.VectorClock {
	// TODO: It's probably possible that the last item in the database doesn't
	// have the most recent clock values for all replicas. Should we store the
	// timestamp in some separate table, too?
	var most_recent_query string = `
		SELECT vector_timestamp FROM messages
			WHERE id = (SELECT MAX(id) FROM messages)
	`

	var timestamp_str = "0,0,0,0,0"

	network.PublicServer.DBPool.QueryRow(
		context.Background(), most_recent_query,
	).Scan(&timestamp_str)

	log.Info("Loaded timestamp", timestamp_str, "from the database.")

	return network.FromDbFormat(timestamp_str)
}

func GetServerID() int {
	server_id := flag.Int("id", -1, "The ID of the server")
	flag.Parse()

	return *server_id
}

func main() {
	//	network.ReplicaId = GetServerID()
	network.ReplicaId = 0

	//	db_host := fmt.Sprintf("chat_db%d", network.ReplicaId)
	db_host := "chat_db"

	client_serve_address := DEFAULT_INTERFACE + ":" + DEFAULT_PUBLIC_PORT
	replication_serve_address := DEFAULT_INTERFACE + ":" + DEFAULT_INTERNAL_PORT

	db.ConnectToDB(db_host)
	defer db.TerminateDBConn()

	network.CurrentTimestamp = loadSavedTimestamp()

	go network.ServerRequestsToReplicas(replication_serve_address, network.ReplicaId)
	network.ServeRequestsToClients(client_serve_address)
}
