package main

import (
	"chat/server/db"
	"chat/server/network"
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
)

// changing this value will change the number of replicas
// the docker-compose file also needs to be changed
const REPLICA_COUNT = 5

// I think this is dead code
const MAX_CHAR_PER_LINE = 80

// this function is called before main() to parse the command line arguments
// and set the server id and the possible replica_ids

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

func UpdateServerID() int {
	server_id := flag.Int("id", -1, "The ID of the server")
	flag.Parse()

	if *server_id == -1 {
		log.Fatalln("Server ID not provided")
	} else if *server_id < 1 || *server_id > REPLICA_COUNT {
		log.Fatalln("Server ID out of range")
	}

	network.InternalServer.SelfID = *server_id
	network.InitializeReplicas(REPLICA_COUNT)

	return *server_id
}

func main() {
	//	network.ReplicaId = GetServerID()
	network.ReplicaId = 0

	/*
		//	db_host := fmt.Sprintf("chat_db%d", network.ReplicaId)
		db_host := "chat_db"

		client_serve_address := DEFAULT_INTERFACE + ":" + DEFAULT_PUBLIC_PORT
		replication_serve_address := DEFAULT_INTERFACE + ":" + DEFAULT_INTERNAL_PORT

		db.ConnectToDB(db_host)
		defer db.TerminateDBConn()

		network.CurrentTimestamp = loadSavedTimestamp()

		go network.ServerRequestsToReplicas(replication_serve_address, network.ReplicaId)
		network.ServeRequestsToClients(client_serve_address)
	*/
	id := UpdateServerID()
	log.Info("Server ID: ", id)

	db.ConnectToDB()

	network.ServePublicRequests()

	network.ServeInternalRequests()

	network.ConnectToReplicas()

	db.TerminateDBConn()
}
