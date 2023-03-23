package main

import (
	"chat/server/db"
	"chat/server/network"
	"flag"
	"strconv"
)

const MAX_CHAR_PER_LINE = 80

const (
	// listen on all interfaces
	DEFAULT_INTERFACE = "0.0.0.0"
	DEFAULT_PORT      = 3000
)

func GetServerFromArgs() string {
	port := flag.Int("port", DEFAULT_PORT, "port number of the chat server")
	flag.Parse()

	return DEFAULT_INTERFACE + ":" + strconv.Itoa(*port)
}

func main() {
	serverAddress := GetServerFromArgs()
	db.ConnectToDB()
	defer db.TerminateDBConn()

	// sync from other replicas
	db.SyncFromReplicas()
	network.ServeRequests(serverAddress)
}
