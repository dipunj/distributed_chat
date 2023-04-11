// these represent the state of the client
package state

import (
	pb "chat/pb"
	"sync"

	"google.golang.org/grpc"
)

// represents the tcp connection
var RpcConn *grpc.ClientConn
var ChatClient pb.PublicClient

var Current_user string
var Current_group string
var Current_group_participants []string

var RecentMessages = MQueue{max_size: 10}
var Rerender chan bool = make(chan bool)

var Wait *sync.WaitGroup = &sync.WaitGroup{}
var RenderMu sync.Mutex
var DataMu sync.Mutex
