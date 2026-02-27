package envconf

import (
	"reflect"
	"sync"
)

type singletonState struct {
	mu                 sync.Mutex
	initialized        bool
	cfg                any
	cfgType            reflect.Type
	optionsFingerprint string
}

var state singletonState

func resetSingletonForTest() {
	state.mu.Lock()
	defer state.mu.Unlock()

	state.initialized = false
	state.cfg = nil
	state.cfgType = nil
	state.optionsFingerprint = ""
}
