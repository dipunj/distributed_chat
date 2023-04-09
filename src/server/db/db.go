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
	//	HOST = "localhost"
	//	HOST = "chat_db"
	PORT = 5432
)

// urlExample := "postgres://username:password@localhost:5432/database_name"
//var db_url string = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", USERNAME, PASSWORD, HOST, PORT, DB_NAME)

func ConnectToDB(host string) {
	db_url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", USERNAME, PASSWORD, host, PORT, DB_NAME)

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
		network.PublicServer.DBPool = dbPool
		network.InternalServer.DBPool = dbPool
	}
}

func TerminateDBConn() {
	network.PublicServer.DBPool.Close()
	// both servers point to the same db pool
	// so no need to close it twice

	// network.InternalServer.DBPool.Close()
}
