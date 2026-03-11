package client

import (
	"context"
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()
	if !opts.tls {
		t.Error("expected TLS enabled by default")
	}
	if opts.timeout != 0 {
		t.Errorf("expected zero timeout by default, got %v", opts.timeout)
	}
}

func TestWithTLS(t *testing.T) {
	opts := defaultOptions()
	WithTLS(false)(opts)
	if opts.tls {
		t.Error("expected TLS disabled")
	}
}

func TestWithAPIKey(t *testing.T) {
	opts := defaultOptions()
	WithAPIKey("test-key")(opts)
	if opts.apiKey != "test-key" {
		t.Error("expected api key to be set")
	}
}

func TestWithHeaders(t *testing.T) {
	opts := defaultOptions()
	WithHeaders(map[string]string{"x-custom": "val"})(opts)
	if opts.headers["x-custom"] != "val" {
		t.Error("expected custom header")
	}
}

func TestWithTimeout(t *testing.T) {
	opts := defaultOptions()
	WithTimeout(10 * time.Second)(opts)
	if opts.timeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", opts.timeout)
	}
}

func TestApplyTimeoutAddsDeadline(t *testing.T) {
	ctx := context.Background()
	newCtx, cancel := applyTimeout(ctx, 5*time.Second)
	defer cancel()

	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	remaining := time.Until(deadline)
	if remaining < 4*time.Second || remaining > 6*time.Second {
		t.Errorf("expected ~5s remaining, got %v", remaining)
	}
}

func TestApplyTimeoutZeroIsNoop(t *testing.T) {
	ctx := context.Background()
	newCtx, cancel := applyTimeout(ctx, 0)
	defer cancel()

	if _, ok := newCtx.Deadline(); ok {
		t.Error("expected no deadline for zero timeout")
	}
}

func TestApplyTimeoutRespectsExistingTighterDeadline(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer ctxCancel()

	// Apply a 10s timeout — the existing 2s deadline should win.
	newCtx, cancel := applyTimeout(ctx, 10*time.Second)
	defer cancel()

	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}
	remaining := time.Until(deadline)
	if remaining > 3*time.Second {
		t.Errorf("expected caller's tighter deadline (~2s), got %v", remaining)
	}
}

func TestApplyTimeoutOverridesLooserDeadline(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	// Apply a 2s timeout — should override the 30s deadline.
	newCtx, cancel := applyTimeout(ctx, 2*time.Second)
	defer cancel()

	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}
	remaining := time.Until(deadline)
	if remaining > 3*time.Second {
		t.Errorf("expected ~2s deadline, got %v", remaining)
	}
}

func TestNewClientInsecure(t *testing.T) {
	c, err := New("localhost:50051", WithTLS(false))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if c.ledger == nil || c.state == nil || c.txExec == nil || c.pkg == nil || c.name == nil || c.sub == nil || c.sigVer == nil {
		t.Error("expected all service clients to be initialized")
	}
}

func TestNewClientWithAPIKey(t *testing.T) {
	c, err := New("localhost:50051", WithTLS(false), WithAPIKey("key123"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}

func TestNewClientWithTimeout(t *testing.T) {
	c, err := New("localhost:50051", WithTLS(false), WithTimeout(15*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}
