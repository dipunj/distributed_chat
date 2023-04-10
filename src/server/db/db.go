package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

var (
	USERNAME = "postgres"
	PASSWORD = "postgres"
	DB_NAME  = "postgres"

	// TODO change this to a "chat_db" when running from a container
	HOST_PREFIX = "chat_db"
	PORT        = 5432
)

func ConnectToDB(selfID int) {
	HOST := HOST_PREFIX + "_" + strconv.Itoa(selfID)

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	db_url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", USERNAME, PASSWORD, HOST, PORT, DB_NAME)

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
		DBPool = dbPool
	}
}

func TerminateDBConn() {
	DBPool.Close()
	// both servers point to the same db pool
	// so no need to close it twice

	// network.InternalServer.DBPool.Close()
}
