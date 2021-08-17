package tmi

import (
	"sync"
	"sync/atomic"
)

type atomicBool struct{ val int32 }

func (atom *atomicBool) set(val bool) {
	if val {
		atomic.StoreInt32(&atom.val, 1)
	} else {
		atomic.StoreInt32(&atom.val, 0)
	}
}
func (atom *atomicBool) get() bool {
	return atomic.LoadInt32(&atom.val) == 1
}

type connCloseErr struct {
	mutex sync.Mutex
	err   error
}

func (t *connCloseErr) update(err error) {
	var override = err == ErrDisconnectCalled || err == ErrLoginFailure
	t.mutex.Lock()
	if t.err == nil {
		t.err = err
	} else if override {
		if t.err != ErrDisconnectCalled && t.err != ErrLoginFailure {
			t.err = err
		}
	}
	t.mutex.Unlock()
}

// notifier's reset() and notify() methods are used in combination to notify multiple goroutines to close.
// call reset() before spawning goroutines
// call notify() in any goroutines to signal one another by listening to the notifier's channel ch
type notifier struct {
	mutex sync.Mutex
	once  *sync.Once
	ch    chan struct{}
}

// notify uses the notifier and makes it unusable until reset.
func (n *notifier) notify() {
	n.mutex.Lock()
	n.once.Do(func() {
		if n.ch != nil {
			close(n.ch)
		}
	})
	n.mutex.Unlock()
}

// reset sets the notifier to be ready to be used.
func (n *notifier) reset() {
	n.mutex.Lock()
	n.once = &sync.Once{}
	n.ch = make(chan struct{})
	n.mutex.Unlock()
}
