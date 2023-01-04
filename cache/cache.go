package cache

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/Bl4ck-h00d/Cache-thread-safe/evictor"
	"github.com/Bl4ck-h00d/Cache-thread-safe/evictor/evictorLRU"
)

var (
	ErrorNotFound = errors.New("error: element not found in cache")
)

const (
	// Setting the eviction percentage which decides how many of the total elements in
	// the cache it will evict once the cache limit exceeds
	EVICTION_PERCENTAGE = 30
)

// Eviction Strategies our cache supports (pluggable)
const (
	FIFO   = "FIFO"
	LRU    = "LRU"
	LIFO   = "LIFO"
	FirstN = "FirstN"
)

type Cache struct {
	// Key-value store where key is of type string, and the value is of type pointer to "Element".
	kv               map[string]*Element
	evictionStrategy string
	MaxSize          int32
	currentSize      int32
	EvictPercent     int32
	// ev is of type Evictor which is implemented by various eviction strategies and assigned accordingly for the cache
	ev    evictor.Evictor
	mutex *sync.RWMutex
}

func (c *Cache) Init(size int32, evictionStrategy string) {
	c.MaxSize = size
	c.currentSize = 0
	c.kv = make(map[string]*Element)
	c.evictionStrategy = evictionStrategy
	c.EvictPercent = EVICTION_PERCENTAGE
	c.mutex = new(sync.RWMutex)
	//TODO: Eviction strategy

	switch strategy := c.evictionStrategy; strategy {
	case LRU:
		c.ev = &evictorLRU.EvictorLRU{}
	default:
		panic("Invalid eviction strategy chosen")
	}
}

// Put inserts/sets a new key-value pair in the cache, and if the key already exists, then
// the value of that key gets updated to the latest value given.
func (c *Cache) Put(key string, value string) {
	e := new(Element)
	e.Init(key, value)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	//If the key is already present in the cache, then we only update it & dont increment current size
	if _, ok := c.kv[key]; ok {
		c.ev.OnUpdate(key, value)
	} else {
		c.currentSize++
		c.ev.OnAdd(key)
	}

	c.kv[key] = e

	if c.currentSize <= c.MaxSize {
		return
	}

	//Running an eviction whenever we add to the cache and the size of the cache exceeds the limit of the cache

	func() {
		fmt.Println("Evicting now...  [Current size: ", c.currentSize, " , Max size: ", c.MaxSize, "]")
		numEvicted := 0
		toEvict := int(c.EvictPercent * c.currentSize / 100)
		fmt.Println("Evicting", toEvict, "out of", c.currentSize, "elements...")
		fmt.Println("Size of cache: ", c.GetSize())

		// The Evict method takes a callback function which gets called and executed inside the
		// Evict method. Another way of doing this is to pass the entire cache as an argument in
		// the function but the down-side of that is that the cache implementation would become
		// tightly coupled but with this current implementation we have both the cache and evictor
		// decoupled and independent.

		// So here the internal Evict function will only return the key to be evicted each time and to
		// indicate when we should stop evicting, we have made this callback function return a boolean

		c.ev.Evict(func(key string) bool {
			numEvicted++
			//For each element selected by our evictor, we delete it from the map in the cache.
			fmt.Println("Evicting element with key ", key)
			delete(c.kv, key)
			c.currentSize--
			if numEvicted >= toEvict {
				return false
			}
			return true
		})
	}()
}

// Get returns the value of the specified key given if found, and return the ErrorNotFound error
// with null string if key not found.
func (c *Cache) Get(key string) (string, error) {
	// The reason we took a Write lock here (and not a Read lock) is bcoz we are modifying
	// in c.ev.OnAccess(key), and since we want to make the cache thread-safe
	//and leave out the evictor from having to handle all this, hence we handle everything here in the cache itself
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k, v := range c.kv {
		if k == key {
			c.ev.OnAccess(key)
			return v.value, nil
		}
	}
	return "", ErrorNotFound
}

func (c *Cache) ViewCacheElements() {
	c.mutex.RLock()
	for k, v := range c.kv {
		fmt.Println("Element in cache: ", k, v)
	}
	c.mutex.RUnlock()
}

func (c *Cache) GetAllKeysSorted() []string {
	sorted := []string{}
	c.mutex.RLock()
	for k, v := range c.kv {
		sorted = append(sorted, k)
		fmt.Println("Element in cache: ", k, v)
	}
	sort.Strings(sorted)
	c.mutex.RUnlock()
	return sorted
}

func (c *Cache) GetSize() int32 {
	return c.currentSize
}

func (c *Cache) Delete(key string) {
	// Calling Delete method on the evictor as well to update it too.
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.kv[key]; !ok {
		fmt.Println("Element to be deleted", key, "not found in cache.")
		return
	}
	c.currentSize--
	c.ev.OnDelete(key)
	fmt.Println("Element", key, "successfully deleted.")
	delete(c.kv, key)
}
