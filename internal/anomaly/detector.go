package anomaly

import "context"

type Detector interface {
	Analyze(ctx context.Context, result any)
}
