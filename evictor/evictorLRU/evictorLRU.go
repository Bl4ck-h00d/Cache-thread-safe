package evictorLRU

import (
	"container/heap"
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	ErrorNotFound = errors.New("error: element with this key not found in cache")
)

type EvictorLRU struct {
	pq PriorityQueue
}

func (e *EvictorLRU) Evict(evictElement func(key string) bool) {
	//loop to keep evicting
	for {
		if len(e.pq) == 0 {
			log.Panicf("Length of pq has become 0")
			return
		}
		// Typecasting the empty interface type to struct
		popped := heap.Pop(&e.pq).(*item)

		fmt.Printf("Popping ----%v", popped)
		ele := popped.key
		if !evictElement(ele) {
			return
		}
	}
}

func (e *EvictorLRU) OnAdd(key string) error {
	ele := item{
		key:              key,
		timeOfLastAccess: time.Now(),
		priority:         int(time.Now().UnixNano()),
	}
	heap.Push(&e.pq, &ele)
	return nil
}

func (e *EvictorLRU) OnAccess(key string) error {
	for _, ele := range e.pq {
		if ele.key == key {
			e.pq.update(ele, time.Now(), int(time.Now().UnixNano()))
			return nil
		}
	}
	return ErrorNotFound
}


func (e *EvictorLRU) OnDelete(key string) error {
	//Pop until we find the right one, storing them temporarily in temp, and then push those popped ones back in.

	temp:=make([]*item,0)

	for e.pq.Len()>0 {
		popped:=heap.Pop(&e.pq).(*item)
		ele:=popped.key
		if ele==key {
			break
		} else {
			temp=append(temp,popped)
		}
	}

	//push back all the remaining items
	for _,ele:=range temp {
		heap.Push(&e.pq,ele)
	}
	return nil
}

func (e *EvictorLRU) OnUpdate(key string, value string) error {
	//Pop until we find the right one, storing them temporarily in temp, and then push those popped ones back in.

	temp:=make([]*item,0)

	for e.pq.Len()>0 {
		popped:=heap.Pop(&e.pq).(*item)
		ele:=popped.key
		if ele==key {
			//Update
			e.pq.update(popped,time.Now(),int(time.Now().UnixNano()))
		} else {
			temp=append(temp,popped)
		}
	}

	//push back all the remaining items
	for _,ele:=range temp {
		heap.Push(&e.pq,ele)
	}
	return nil
}