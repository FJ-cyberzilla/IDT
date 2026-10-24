package inspect

import (
	"context"
	"testing"
	"time"
)

func TestRunDiagnostics(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	diag := RunDiagnostics(ctx)
	if diag.ConnectivityStatus == "" {
		t.Error("expected connectivity status message")
	}
}
