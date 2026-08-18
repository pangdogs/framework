package fanout

import "testing"

func TestBroadcasterSubscribeSnapshotUnsubscribe(t *testing.T) {
	b := NewBroadcaster[string, int]()

	if got := b.Snapshot(); got != nil {
		t.Fatalf("expected nil initial snapshot, got %v", got)
	}

	a := b.Subscribe("a", 1)
	bsub := b.Subscribe("b", 2)

	if a.Handler != "a" || cap(a.Inbox) != 1 {
		t.Fatalf("unexpected first subscription: handler %q capacity %d", a.Handler, cap(a.Inbox))
	}

	snapshot := b.Snapshot()
	if len(snapshot) != 2 || snapshot[0] != a || snapshot[1] != bsub {
		t.Fatalf("unexpected subscription snapshot: %v", snapshot)
	}

	b.Unsubscribe(a)
	snapshot = b.Snapshot()
	if len(snapshot) != 1 || snapshot[0] != bsub {
		t.Fatalf("unexpected snapshot after unsubscribe: %v", snapshot)
	}

	b.Unsubscribe(a)
	snapshot = b.Snapshot()
	if len(snapshot) != 1 || snapshot[0] != bsub {
		t.Fatalf("unexpected snapshot after repeated unsubscribe: %v", snapshot)
	}
}

func TestBroadcasterBroadcast(t *testing.T) {
	var b Broadcaster[string, int]
	a := b.Subscribe("a", 1)
	bsub := b.Subscribe("b", 1)

	if dropped := b.Broadcast(7); dropped != 0 {
		t.Fatalf("unexpected dropped count: got %d want 0", dropped)
	}

	for name, sub := range map[string]*Subscription[string, int]{"a": a, "b": bsub} {
		select {
		case got := <-sub.Inbox:
			if got != 7 {
				t.Fatalf("unexpected message for %s: got %d want 7", name, got)
			}
		default:
			t.Fatalf("expected message for %s", name)
		}
	}
}

func TestBroadcasterBroadcastDropsFullInbox(t *testing.T) {
	var b Broadcaster[string, int]
	a := b.Subscribe("a", 1)
	bsub := b.Subscribe("b", 1)
	a.Inbox <- 1

	if dropped := b.Broadcast(2); dropped != 1 {
		t.Fatalf("unexpected dropped count: got %d want 1", dropped)
	}

	if got := <-a.Inbox; got != 1 {
		t.Fatalf("unexpected retained message for a: got %d want 1", got)
	}
	if got := <-bsub.Inbox; got != 2 {
		t.Fatalf("unexpected broadcast message for b: got %d want 2", got)
	}
}
