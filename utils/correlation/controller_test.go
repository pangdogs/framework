package correlation

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.golaxy.org/core/utils/async"
)

func TestControllerResolve(t *testing.T) {
	controller := New(context.Background(), time.Second)
	t.Cleanup(func() { controller.Close() })

	id, future, err := controller.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	if id == 0 {
		t.Fatal("correlation ID is zero")
	}

	want := async.NewResult("ok", nil)
	if !controller.Resolve(id, want) {
		t.Fatal("Resolve failed")
	}

	got := future.Wait(context.Background())
	if got.Error != nil || got.Value != want.Value {
		t.Fatalf("unexpected result: %+v", got)
	}
	if controller.Resolve(id, want) {
		t.Fatal("second Resolve succeeded")
	}
}

func TestControllerCancel(t *testing.T) {
	controller := New(context.Background(), time.Second)
	t.Cleanup(func() { controller.Close() })

	id, future, err := controller.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	cause := errors.New("cancelled")
	if !controller.Cancel(id, cause) {
		t.Fatal("Cancel failed")
	}
	if got := future.Wait(context.Background()); !errors.Is(got.Error, cause) {
		t.Fatalf("unexpected cancellation error: %v", got.Error)
	}
	if controller.Cancel(id, cause) {
		t.Fatal("second Cancel succeeded")
	}
}

func TestControllerTimeout(t *testing.T) {
	controller := New(context.Background(), 20*time.Millisecond)
	t.Cleanup(func() { controller.Close() })

	id, future, err := controller.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	if got := future.Wait(context.Background()); !errors.Is(got.Error, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", got.Error)
	}
	if controller.Resolve(id, async.NewResult("late", nil)) {
		t.Fatal("late Resolve succeeded")
	}
}

func TestControllerClose(t *testing.T) {
	controller := New(context.Background(), time.Second)
	id, future, err := controller.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	if !controller.Close() {
		t.Fatal("first Close failed")
	}
	if !controller.Done().Completed() {
		t.Fatal("Done was not completed")
	}
	if got := future.Wait(context.Background()); !errors.Is(got.Error, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", got.Error)
	}
	if controller.Resolve(id, async.NewResult(nil, nil)) {
		t.Fatal("Resolve after Close succeeded")
	}
	if _, _, err := controller.Begin(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected Begin after Close to fail with ErrClosed, got %v", err)
	}
	if controller.Close() {
		t.Fatal("second Close succeeded")
	}
}

func TestControllerClosesWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	controller := New(ctx, time.Second)

	_, future, err := controller.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	cancel()

	waitCtx, stopWait := context.WithTimeout(context.Background(), time.Second)
	defer stopWait()
	if err := controller.Done().Wait(waitCtx); err != nil {
		t.Fatalf("Controller did not close in time: %v", err)
	}
	if got := future.Wait(context.Background()); !errors.Is(got.Error, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", got.Error)
	}
}

func TestControllerStartsClosedWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	controller := New(ctx, time.Second)
	if !controller.Done().Completed() {
		t.Fatal("Controller created with a cancelled context is not closed")
	}
	if _, _, err := controller.Begin(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected Begin to fail with ErrClosed, got %v", err)
	}
}

func TestControllerRejectsUnknownID(t *testing.T) {
	controller := New(context.Background(), time.Second)
	t.Cleanup(func() { controller.Close() })

	if controller.Resolve(12345, async.NewResult(nil, nil)) {
		t.Fatal("Resolve accepted an unknown ID")
	}
	if controller.Cancel(12345, nil) {
		t.Fatal("Cancel accepted an unknown ID")
	}
}
