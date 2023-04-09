package main

import (
	"chat/server/db"
	"chat/server/network"
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

	id := UpdateServerID()
	log.Info("Server ID: ", id)

	db.ConnectToDB()

	network.ServePublicRequests()

	network.ServeInternalRequests()

	network.ConnectToReplicas()

	db.TerminateDBConn()
}
