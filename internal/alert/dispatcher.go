package alert

import (
	"context"
	"strings"
	"sync"

	"github.com/pulse/pulse/internal/models"
)

type Dispatcher interface {
	Send(ctx context.Context, alert models.Alert) error
}

type MultiDispatcher struct {
	dispatchers []Dispatcher
}

func NewMultiDispatcher(dispatchers ...Dispatcher) *MultiDispatcher {
	return &MultiDispatcher{
		dispatchers: dispatchers,
	}
}

type MultiError []error

func (m MultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

func (m *MultiDispatcher) Send(ctx context.Context, alert models.Alert) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for _, d := range m.dispatchers {
		wg.Add(1)
		go func(disp Dispatcher) {
			defer wg.Done()
			if err := disp.Send(ctx, alert); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(d)
	}

	wg.Wait()

	if len(errs) > 0 {
		return MultiError(errs)
	}
	return nil
}
