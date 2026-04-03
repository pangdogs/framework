package concurrent

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.golaxy.org/core/utils/async"
)

func TestFutureControllerResolve(t *testing.T) {
	fc := NewFutureController(context.Background(), time.Second)

	handle, err := fc.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	want := async.NewResult("ok", nil)
	if err := fc.Resolve(handle.Id(), want); err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	got := handle.Future().Wait(context.Background())
	if got.Error != nil {
		t.Fatalf("unexpected future error: %v", got.Error)
	}
	if got.Value != want.Value {
		t.Fatalf("unexpected future value: got %v want %v", got.Value, want.Value)
	}

	if err := fc.Resolve(handle.Id(), want); !errors.Is(err, ErrFutureExceeded) {
		t.Fatalf("expected ErrFutureExceeded, got %v", err)
	}
}

func TestFutureHandleResolveAndCancel(t *testing.T) {
	fc := NewFutureController(context.Background(), time.Second)

	handle, err := fc.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := handle.Resolve(async.NewResult(123, nil)); err != nil {
		t.Fatalf("handle.Resolve failed: %v", err)
	}

	got := handle.Future().Wait(context.Background())
	if got.Error != nil || got.Value != 123 {
		t.Fatalf("unexpected result: %+v", got)
	}

	handle2, err := fc.New()
	if err != nil {
		t.Fatalf("second New failed: %v", err)
	}

	cancelErr := errors.New("cancelled")
	handle2.Cancel(cancelErr)

	got = handle2.Future().Wait(context.Background())
	if !errors.Is(got.Error, cancelErr) {
		t.Fatalf("unexpected cancel error: %v", got.Error)
	}
}

func TestFutureControllerTimeout(t *testing.T) {
	fc := NewFutureController(context.Background(), 20*time.Millisecond)

	handle, err := fc.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	got := handle.Future().Wait(context.Background())
	if !errors.Is(got.Error, ErrFutureExceeded) {
		t.Fatalf("expected ErrFutureExceeded, got %v", got.Error)
	}

	if err := fc.Resolve(handle.Id(), async.NewResult("late", nil)); !errors.Is(err, ErrFutureExceeded) {
		t.Fatalf("expected late resolve to fail with ErrFutureExceeded, got %v", err)
	}
}

func TestFutureControllerClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fc := NewFutureController(ctx, 50*time.Millisecond)

	handle, err := fc.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	cancel()

	got := handle.Future().Wait(context.Background())
	if !errors.Is(got.Error, ErrFutureControllerClosed) && !errors.Is(got.Error, ErrFutureExceeded) {
		t.Fatalf("expected ErrFutureControllerClosed or ErrFutureExceeded, got %v", got.Error)
	}

	select {
	case <-fc.Terminated().Done():
	case <-time.After(time.Second):
		t.Fatal("controller did not terminate in time")
	}

	if _, err := fc.New(); !errors.Is(err, ErrFutureControllerClosed) {
		t.Fatalf("expected New after close to fail with ErrFutureControllerClosed, got %v", err)
	}
}

func TestFutureControllerResolveUnknownID(t *testing.T) {
	fc := NewFutureController(context.Background(), time.Second)

	if err := fc.Resolve(12345, async.NewResult(nil, nil)); !errors.Is(err, ErrFutureExceeded) {
		t.Fatalf("expected ErrFutureExceeded, got %v", err)
	}
}

func TestFutureControllerGenIDSkipsZero(t *testing.T) {
	fc := NewFutureController(context.Background(), time.Second)
	fc.idGen.Store(-1)

	if got := fc.genId(); got != 1 {
		t.Fatalf("unexpected generated id: got %d want 1", got)
	}
}
