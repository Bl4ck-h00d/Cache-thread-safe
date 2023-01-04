### Highlights of the project
- The evictor has been made an interface, and every/any strategy (LRU, FIFO, LIFO, Custom/Random, etc) we want to add-on can be done by implementing this interface.
- The cache implementation and the eviction strategies are decoupled, making it easily pluggable for anyone to add more strategies.
- Its is thread safe since we are making use of RWMutex and acquiring lock whenever we are accessing the shared memory in the implemented cache (in our case it is both c.kv and c.ev).
- The evictor interface need not take care of thread safety as the cache itself will take care of that.
- We have set an eviction percentage, hence Evict is not called very often, and whenever called, it is called for few entries.

### Simple Testing
The two major test files are: 
1. stress_test.go 
  - This is doing an end to end stress test with multiple go routines where we are simultaneuously adding/getting/deleting from the cache. Hence also checks if its thread safe and ensures that we don't enter a deadlock.
  - To run:  `go test -v stress_test.go`
2. main_test.go.
  - This has muliple smaller individual unittests to test smaller functionalities in the code.
  - To run:  `go test -v main_test.go`