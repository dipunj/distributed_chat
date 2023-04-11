// contains the queue data structure which is used to render the 10 recent messages for the client

package state

import (
	pb "chat/pb"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

type MQueue struct {
	max_size int
	q        *[]*pb.TextMessage
	mutex    sync.Mutex
}

func (qu *MQueue) ClearAll() {
	qu.mutex.Lock()

	qu.q = &[]*pb.TextMessage{}

	qu.mutex.Unlock()
}

func (qu *MQueue) ListAll() []*pb.TextMessage {
	qu.mutex.Lock()
	defer qu.mutex.Unlock()

	return *qu.q
}

func (qu *MQueue) Replace(newList *[]*pb.TextMessage) {
	qu.mutex.Lock()

	qu.q = newList

	qu.mutex.Unlock()
}

// position 1...max_size
func (qu *MQueue) GetFromLineNo(position int) *pb.TextMessage {
	qu.mutex.Lock()
	defer qu.mutex.Unlock()

	if position >= 1 && position <= len(*qu.q) {
		return (*qu.q)[position-1]
	} else {
		log.Error("invalid line number. Please choose a line no b/n [1..."+strconv.Itoa(len(*qu.q))+"]", position)
		return nil
	}

}

func (qu *MQueue) Update_if_exists(data *pb.TextMessage) bool {
	qu.mutex.Lock()
	defer qu.mutex.Unlock()

	for i := 0; i < len(*qu.q); i++ {
		if (*qu.q)[i].Id == data.Id {
			(*qu.q)[i] = data
			return true // element was present in the queue
		}
	}

	// element wasn't present in the queue
	return false
}
