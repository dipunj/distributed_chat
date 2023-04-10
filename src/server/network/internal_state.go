package network

// the actual internal server object
var InternalServer = InternalServerType{}

// this object stores the client objects each replica

// ReplicaIds will be populated at run time
// based on what id is given to the current instance of the server

var SelfID int
var ReplicaIds = []int{}

var ReplicaState map[int]*ReplicaStateType = make(map[int]*ReplicaStateType)
