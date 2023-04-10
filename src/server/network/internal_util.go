package network

import (
	"strconv"
)

func GetReplicaAddressFromID(replicaID int, port string) string {
	ip_prefix := "172.30.100.10"
	return ip_prefix + strconv.Itoa(replicaID) + ":" + port
}

func InitializeReplicas(replica_count int) {

	// populate replica_ids array with ids from 1 to replica_count, except selfID
	for i := 1; i <= replica_count; i++ {
		if i != SelfID {
			ReplicaIds = append(ReplicaIds, i)
			// initially we assume all replicas are offline
			ReplicaState[i] = &ReplicaStateType{
				Changed:           make(chan bool),
				IsOnline:          false,
				PublicIpAddress:   GetReplicaAddressFromID(i, DEFAULT_PUBLIC_PORT),
				InternalIpAddress: GetReplicaAddressFromID(i, DEFAULT_INTERNAL_PORT),
			}
		}
	}

}
