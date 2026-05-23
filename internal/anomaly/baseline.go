package anomaly

import (
	"math"
	"sync"
)

type circularBuffer struct {
	data  []int64
	head  int
	count int
}

type Baseline struct {
	mu         sync.RWMutex
	windowSize int
	buffers    map[string]*circularBuffer
}

func NewBaseline(windowSize int) *Baseline {
	return &Baseline{
		windowSize: windowSize,
		buffers:    make(map[string]*circularBuffer),
	}
}

func (b *Baseline) Add(serviceID string, latencyMs int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	cb, ok := b.buffers[serviceID]
	if !ok {
		cb = &circularBuffer{
			data: make([]int64, b.windowSize),
		}
		b.buffers[serviceID] = cb
	}

	cb.data[cb.head] = latencyMs
	cb.head = (cb.head + 1) % b.windowSize
	if cb.count < b.windowSize {
		cb.count++
	}
}

func (b *Baseline) Mean(serviceID string) float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	cb, ok := b.buffers[serviceID]
	if !ok || cb.count == 0 {
		return 0
	}

	var sum int64
	for i := 0; i < cb.count; i++ {
		sum += cb.data[i]
	}
	return float64(sum) / float64(cb.count)
}

func (b *Baseline) StdDev(serviceID string) float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	cb, ok := b.buffers[serviceID]
	if !ok || cb.count == 0 {
		return 0
	}

	var sum int64
	for i := 0; i < cb.count; i++ {
		sum += cb.data[i]
	}
	mean := float64(sum) / float64(cb.count)

	var varianceSum float64
	for i := 0; i < cb.count; i++ {
		diff := float64(cb.data[i]) - mean
		varianceSum += diff * diff
	}

	return math.Sqrt(varianceSum / float64(cb.count))
}

func (b *Baseline) SampleCount(serviceID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	cb, ok := b.buffers[serviceID]
	if !ok {
		return 0
	}
	return cb.count
}
