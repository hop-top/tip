package observe

import "context"

// ctxBackground returns a background context. Extracted to keep observe.go
// free of direct context import (avoids collision with tip's context pkg).
func ctxBackground() context.Context {
	return context.Background()
}
