package main

import (
	"chat/client/network"
	"chat/client/state"
	"chat/client/ui"
	"errors"
	"flag"
	"net"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	// name of the container in the docker network
	DEFAULT_SERVER = "chat_server"
	DEFAULT_PORT   = 3000
)

func GetServerFromArgs() string {
	server := flag.String("host", DEFAULT_SERVER, "ip address/hostname of the chat server")
	port := flag.Int("port", DEFAULT_PORT, "port number of the chat server")
	flag.Parse()

	server_ip := net.ParseIP(*server)

	log.Info(*server, server_ip)

	if server_ip == nil {
		server_ips, err := net.LookupIP(*server)

		if err != nil {
			// dns lookup failed
			log.Fatal("DNS lookup failed", err)
		} else if len(server_ips) > 0 {
			server_ip = server_ips[len(server_ips)-1]
			log.Info("hostname", *server, "resolved to", server_ip)
		} else {
			log.Fatal(errors.New("invalid hostname: " + *server))
		}
	}

	return server_ip.String() + ":" + strconv.Itoa(*port)
}

func main() {
	serverAddr := GetServerFromArgs()

	// todo
	// 1. wait for the connection to be established before starting the user input loop
	// 2. add a mechanism to know if the connection is broken
	// 3. configure timeout
	network.EstablishConnection(serverAddr)
	network.EstablishStream()

	os.Stdout.WriteString("Welcome...enter h for help")

	state.Wait.Add(1)
	defer state.Wait.Done()
	go ui.UserViewLoop()

	// on the main thread
	ui.UserInputLoop()
}
