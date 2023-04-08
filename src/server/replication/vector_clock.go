package replication

type VectorClock struct {
	my_id  int
	clocks [5]int
}

func (vc *VectorClock) increment() {
	vc.clocks[vc.my_id] += 1
}

func (vc *VectorClock) update_from(other VectorClock) {
	for i := range vc.clocks {
		if i == vc.my_id {
			continue
		}

		if other.clocks[i] > vc.clocks[i] {
			vc.clocks[i] = other.clocks[i]
		}
	}
}
