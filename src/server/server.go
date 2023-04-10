package main

import (
	"chat/server/db"
	"chat/server/network"
	"flag"

	log "github.com/sirupsen/logrus"
)

// changing this value will change the number of replicas
// the docker-compose file also needs to be changed
const TOTAL_SERVER_COUNT = 5

// I think this is dead code
const MAX_CHAR_PER_LINE = 80

// this function is called before main() to parse the command line arguments
// and set the server id and the possible replica_ids

func InitializeServerID() int {
	server_id := flag.Int("id", -1, "The ID of the server")
	flag.Parse()

	if *server_id == -1 {
		log.Fatalln("Server ID not provided")
	} else if *server_id < 1 || *server_id > TOTAL_SERVER_COUNT {
		log.Fatalln("Server ID out of range")
	}

	log.Info("Server ID is: ", *server_id)

	network.SelfID = *server_id
	network.InitializeClock(TOTAL_SERVER_COUNT)

	network.InitializeReplicas(TOTAL_SERVER_COUNT)

	return *server_id
}

func InitializeLogger() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	InitializeLogger()

	InitializeServerID()

	db.ConnectToDB(network.SelfID)

	go network.ServeInternalRequests()

	network.ConnectToReplicas()

	network.ServePublicRequests()

	db.TerminateDBConn()
}
