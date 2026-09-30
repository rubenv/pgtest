package pgtest

import "sync"

const maxOutput = 64 * 1024

type limitedBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := len(p)
	if len(p) >= maxOutput {
		b.data = append(b.data[:0], p[len(p)-maxOutput:]...)
	} else {
		if excess := len(b.data) + len(p) - maxOutput; excess > 0 {
			copy(b.data, b.data[excess:])
			b.data = b.data[:len(b.data)-excess]
		}
		b.data = append(b.data, p...)
	}
	return n, nil
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}
