package main

import (
	"sync"
)

type Keys struct {
	keys  []string
	lock  sync.Mutex
	index int
}

func NewKeys(keys []string) Keys {
	return Keys{
		keys:  keys,
		lock:  sync.Mutex{},
		index: 0,
	}
}

func (k *Keys) nextKey() string {
	k.lock.Lock()
	key := k.keys[k.index]
	k.index = (k.index + 1) % len(k.keys)
	k.lock.Unlock()

	return key
}
