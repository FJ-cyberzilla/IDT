package helper

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSafeRun(t *testing.T) {
	// Normal execution
	err := SafeRun(func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Panic recovery execution
	err = SafeRun(func() error {
		panic("boom")
	})
	if err == nil {
		t.Fatal("expected panic to be caught and returned as error, got nil")
	}
}

func TestRetry(t *testing.T) {
	ctx := context.Background()

	// Succeeds on first try
	attempts := 0
	err := Retry(ctx, 3, 1*time.Millisecond, 10*time.Millisecond, func(c context.Context) error {
		attempts++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected retry error: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}

	// Succeeds after failures
	attempts = 0
	err = Retry(ctx, 3, 1*time.Millisecond, 10*time.Millisecond, func(c context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected retry error after recovery: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}

	// Fails permanently
	attempts = 0
	err = Retry(ctx, 2, 1*time.Millisecond, 10*time.Millisecond, func(c context.Context) error {
		attempts++
		return errors.New("permanent failure")
	})
	if err == nil {
		t.Fatal("expected permanent failure error, got nil")
	}
}

func TestGetSharedTransport(t *testing.T) {
	tr1 := GetSharedTransport()
	if tr1 == nil {
		t.Fatal("expected shared transport, got nil")
	}
	tr2 := GetSharedTransport()
	if tr1 != tr2 {
		t.Fatal("expected singleton transport instance")
	}
}

func TestIsPrivateIP(t *testing.T) {
	if !IsPrivateIP("127.0.0.1") {
		t.Error("expected 127.0.0.1 to be recognized as private/loopback")
	}
	if !IsPrivateIP("10.1.2.3") {
		t.Error("expected 10.1.2.3 to be recognized as private")
	}
	if !IsPrivateIP("192.168.1.1") {
		t.Error("expected 192.168.1.1 to be recognized as private")
	}
	if IsPrivateIP("8.8.8.8") {
		t.Error("expected 8.8.8.8 NOT to be recognized as private")
	}
	if IsPrivateIP("invalid-ip") {
		t.Error("expected invalid IP string to return false")
	}
}
