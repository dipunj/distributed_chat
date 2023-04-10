package network

import (
	"context"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type VectorClock struct {
	clocks []int64
}

var Clock VectorClock

func LoadSavedTimestamp() VectorClock {
	// TODO: It's probably possible that the last item in the database doesn't
	// have the most recent clock values for all replicas. Should we store the
	// timestamp in some separate table, too?
	var most_recent_query string = `
		SELECT vector_timestamp FROM messages
			WHERE id = (SELECT MAX(id) FROM messages)
	`

	var timestamp_str = "0,0,0,0,0"

	PublicServer.DBPool.QueryRow(
		context.Background(), most_recent_query,
	).Scan(&timestamp_str)

	log.Info("Loaded timestamp", timestamp_str, "from the database.")

	return FromDbFormat(timestamp_str)
}

func InitializeClock(num_replicas int) {
	Clock = VectorClock{clocks: make([]int64, num_replicas)}
}

// Increment the vector clock and return its previous value
func (vc *VectorClock) Increment() VectorClock {
	var ret = VectorClock{clocks: make([]int64, len(vc.clocks))}
	copy(ret.clocks, vc.clocks)
	vc.clocks[SelfID] += 1
	return ret
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (vc *VectorClock) UpdateFrom(other VectorClock) {
	for i := range vc.clocks {
		if i != SelfID {
			vc.clocks[i] = maxInt64(vc.clocks[i], other.clocks[i])
		}
	}
}

// Convert timestamp to a string that can be used in a SQL INSERT statement
func (vc VectorClock) ToDbFormat() string {
	var clock_strings = make([]string, len(vc.clocks))

	for i := range clock_strings {
		clock_strings[i] = strconv.Itoa(int(vc.clocks[i]))
	}

	return "{" + strings.Join(clock_strings[:], ",") + "}"
}

func FromDbFormat(db_str string) VectorClock {
	var clock_strings = strings.Split(db_str, ",")
	var clocks = make([]int64, len(clock_strings))

	for i := range clocks {
		c, err := strconv.Atoi(clock_strings[i])
		if err != nil {
			clocks[i] = -1
		} else {
			clocks[i] = int64(c)
		}
	}

	return VectorClock{clocks: clocks}
}
