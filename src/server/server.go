package main

import (
	"chat/server/db"
	"chat/server/network"
	"flag"
)

// I think this is dead code
const MAX_CHAR_PER_LINE = 80

const (
	// listen on all interfaces
	DEFAULT_INTERFACE     = "0.0.0.0"
	DEFAULT_PUBLIC_PORT   = "12000"
	DEFAULT_INTERNAL_PORT = "11000"

	PUBLIC_ADDRESS   = DEFAULT_INTERFACE + ":" + DEFAULT_PUBLIC_PORT
	INTERNAL_ADDRESS = DEFAULT_INTERFACE + ":" + DEFAULT_INTERNAL_PORT
)

// this function is called before main() to parse the command line arguments
// and set the server id and the possible replica_ids

func UpdateServerID() int {
	server_id := flag.Int("id", -1, "The ID of the server")
	flag.Parse()

	if *server_id == -1 {
		panic("Server ID not provided")
	} else if *server_id < 1 || *server_id > network.REPLICA_COUNT {
		panic("Server ID out of range")
	}

	network.InternalServer.SelfID = *server_id

	// populate replica_ids array with ids from 1 to replica_count, except selfID
	for i := 1; i <= network.REPLICA_COUNT; i++ {
		if i != *server_id {
			network.ReplicaIds = append(network.ReplicaIds, i)
		}
	}
	return *server_id
}

func main() {

	UpdateServerID()

	db.ConnectToDB()

	network.ServePublicRequests(PUBLIC_ADDRESS)

	network.ServeInternalRequests(INTERNAL_ADDRESS)

	// one client per replica
	network.StartInternalClients()

	db.TerminateDBConn()
}
