package network

import (
	"strconv"
	"strings"
)

const NUM_REPLICAS = 5

type VectorClock struct {
	my_id  int
	clocks [NUM_REPLICAS]int
}

func (vc *VectorClock) Increment() {
	vc.clocks[vc.my_id] += 1
}

func (vc *VectorClock) UpdateFrom(other VectorClock) {
	for i := range vc.clocks {
		if i == vc.my_id {
			continue
		}

		if other.clocks[i] > vc.clocks[i] {
			vc.clocks[i] = other.clocks[i]
		}
	}
}

func (vc *VectorClock) ToDbFormat() string {
	var clock_strings [NUM_REPLICAS]string

	for i := range clock_strings {
		clock_strings[i] = strconv.Itoa(vc.clocks[i])
	}

	return strings.Join(clock_strings[:], ",")
}

func FromDbFormat(db_str string) VectorClock {
	var clock_strings = strings.Split(db_str, ",")
	var clocks [NUM_REPLICAS]int
	for i := range clocks {
		c, err := strconv.Atoi(clock_strings[i])
		if err != nil {
			clocks[i] = -1
		} else {
			clocks[i] = c
		}
	}

	return VectorClock{my_id: -1, clocks: clocks}
}
