package network

import (
	"strconv"
	"strings"
)

type VectorClock struct {
	clocks []int
}

// Increment the vector clock and return its previous value
func (vc *VectorClock) Increment(my_id int) VectorClock {
	var ret = VectorClock{clocks: make([]int, len(vc.clocks))}
	copy(ret.clocks, vc.clocks)
	vc.clocks[my_id] += 1
	return ret
}

func (vc *VectorClock) UpdateFrom(other VectorClock, my_id int) {
	for i := range vc.clocks {
		if i == my_id {
			continue
		}

		if other.clocks[i] > vc.clocks[i] {
			vc.clocks[i] = other.clocks[i]
		}
	}
}

// Convert timestamp
func (vc VectorClock) ToDbFormat() string {
	var clock_strings = make([]string, len(vc.clocks))

	for i := range clock_strings {
		clock_strings[i] = strconv.Itoa(vc.clocks[i])
	}

	return "{" + strings.Join(clock_strings[:], ",") + "}"
}

func FromDbFormat(db_str string) VectorClock {
	var clock_strings = strings.Split(db_str, ",")
	var clocks = make([]int, len(clock_strings))

	for i := range clocks {
		c, err := strconv.Atoi(clock_strings[i])
		if err != nil {
			clocks[i] = -1
		} else {
			clocks[i] = c
		}
	}

	return VectorClock{clocks: clocks}
}
