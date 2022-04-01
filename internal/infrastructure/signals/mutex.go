package signals

import "sync"

type boolMutex struct {
	mutex *sync.Mutex
	value bool
}

func newBoolMutex() *boolMutex {
	return &boolMutex{
		mutex: &sync.Mutex{},
	}
}

func (m *boolMutex) setTrue() {
	m.mutex.Lock()
	m.value = true
	m.mutex.Unlock()
}

func (m *boolMutex) read() bool {
	m.mutex.Lock()
	v := m.value
	m.mutex.Unlock()
	return v
}

type errorMutex struct {
	mutex *sync.Mutex
	value error
}

func newErrorMutex() *errorMutex {
	return &errorMutex{
		mutex: &sync.Mutex{},
	}
}

func (m *errorMutex) set(err error) {
	m.mutex.Lock()
	m.value = err
	m.mutex.Unlock()
}

func (m *errorMutex) read() error {
	m.mutex.Lock()
	v := m.value
	m.mutex.Unlock()
	return v
}
