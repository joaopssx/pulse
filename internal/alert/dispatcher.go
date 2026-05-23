package alert

import (
	"context"

	"github.com/pulse/pulse/internal/models"
)

type Dispatcher interface {
	Dispatch(ctx context.Context, alert models.Alert) error
}
