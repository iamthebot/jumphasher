package jumphasher

import (
	"sync/atomic"
)

//Golang implementation of std::atomic_flag from C++11
//
//Used for lock-free coordination between workers for shutdown
type AtomicFlag struct {
	flag uint32
}

//Atomically clears AtomicFlag and sets it to false
func (a *AtomicFlag) Clear() {
	atomic.StoreUint32(&a.flag, 0)
}

//Atomically sets AtomicFlag to true and returns its previous value
func (a *AtomicFlag) TestAndSet() bool {
	return atomic.SwapUint32(&a.flag, 1) == 1
}

//Atomically checks AtomicFlag
func (a *AtomicFlag) Test() bool {
	return atomic.LoadUint32(&a.flag) == 1
}
