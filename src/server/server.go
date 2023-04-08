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
)

func GetServerID() int {
	server_id := flag.Int("id", -1, "The ID of the server")
	flag.Parse()

	return *server_id
}

func main() {

	server_id := GetServerID()

	client_serve_address := DEFAULT_INTERFACE + ":" + DEFAULT_PUBLIC_PORT
	replication_serve_address := DEFAULT_INTERFACE + ":" + DEFAULT_INTERNAL_PORT

	db.ConnectToDB()
	defer db.TerminateDBConn()

	network.ServerRequestsToReplicas(replication_serve_address, server_id)
	network.ServeRequestsToClients(client_serve_address)
}
