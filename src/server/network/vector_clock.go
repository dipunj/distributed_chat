package network

import (
	"chat/server/db"
	"context"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type VectorClock []int64

var Clock VectorClock
var ClockMu sync.Mutex

func LoadSavedTimestamp(num_replicas int) VectorClock {
	// TODO: It's probably possible that the last item in the database doesn't
	// have the most recent clock values for all replicas. Should we store the
	// timestamp in some separate table, too?
	var most_recent_query string = `
		SELECT vector_timestamp FROM messages
			WHERE id = (SELECT MAX(id) FROM messages)
	`

	var timestamp_str = "0,0,0,0,0"

	db.DBPool.QueryRow(
		context.Background(), most_recent_query,
	).Scan(&timestamp_str)

	log.Info("Loaded timestamp", timestamp_str, "from the database.")

	return FromDbFormat(timestamp_str)
}

func InitializeClock(num_replicas int) {
	// 0th index is unused
	ClockMu.Lock()

	Clock = make(VectorClock, num_replicas+1)
	res, _ := InternalServer.GetLatestClock(context.Background(), &emptypb.Empty{})
	Clock = res.Clock

	ClockMu.Unlock()
}

// Increment the vector clock and return a copy of the new clock
func (clk *VectorClock) Increment() VectorClock {
	log.Debug("Incrementing clock. Value before incrementing: ", *clk)
	ClockMu.Lock()
	defer ClockMu.Unlock()

	// Increment the clock for this replica
	(*clk)[SelfServerID]++

	// Return a copy of the new clock
	new_clock := make(VectorClock, len(*clk))
	copy(new_clock, *clk)

	log.Debug("Incrementing clock. Value after incrementing: ", *clk, new_clock)
	return new_clock
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (own *VectorClock) UpdateFrom(other VectorClock) {
	ClockMu.Lock()

	for i := range *own {
		if i != SelfServerID {
			(*own)[i] = maxInt64((*own)[i], other[i])
		}
	}

	ClockMu.Unlock()
}

// Convert timestamp to a string that can be used in a SQL INSERT statement
func (vc *VectorClock) ToDbFormat() string {
	stringArray := make([]string, len(*vc))

	for i, val := range *vc {
		stringArray[i] = strconv.FormatInt(val, 10)
	}

	result := strings.Join(stringArray, ",")

	return "{" + result + "}"
}

// Convert string in DB format to a VectorClock
func FromDbFormat(db_str string) VectorClock {
	db_str = strings.Trim(db_str, "{}")    // Remove the outer braces from the input string
	strArray := strings.Split(db_str, ",") // Split the string into an array of strings
	node_count := len(strArray)

	vc := make(VectorClock, node_count+1)

	for i, val := range strArray {
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			vc[i] = intVal
		}
	}

	return vc
}
