package db

import (
	"chat/server/network"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

const (
	USERNAME = "postgres"
	PASSWORD = "postgres"
	DB_NAME  = "postgres"

	// TODO change this to a "chat_db" when running from a container
	HOST = "localhost"
	PORT = 5432
)

// urlExample := "postgres://username:password@localhost:5432/database_name"
var db_url string = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", USERNAME, PASSWORD, HOST, PORT, DB_NAME)

func ConnectToDB() {
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, db_url)

	if err != nil {
		log.Fatal("Couldn't connect to database", err)
	}
	ping_err := dbPool.Ping(ctx)

	if ping_err != nil {
		log.Fatal("Failed to establish connection with DB", ping_err)
	} else {
		log.Info("Database connection established")
		network.Server.DBPool = dbPool
	}
}

func SyncFromReplicas() {

}

func TerminateDBConn() {
	network.Server.DBPool.Close()
}
