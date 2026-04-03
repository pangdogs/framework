package concurrent

import "testing"

func TestNewListener(t *testing.T) {
	l := NewListener[string, int]("handler", 2)

	if l.Handler != "handler" {
		t.Fatalf("unexpected handler: got %q want %q", l.Handler, "handler")
	}
	if cap(l.Inbox) != 2 {
		t.Fatalf("unexpected inbox size: got %d want %d", cap(l.Inbox), 2)
	}
}

func TestListenersAddLoadDelete(t *testing.T) {
	ls := NewListeners[string, int]()

	if got := ls.Load(); got != nil {
		t.Fatalf("expected nil initial snapshot, got %v", got)
	}

	a := ls.Add("a", 1)
	b := ls.Add("b", 1)

	snap := ls.Load()
	if len(snap) != 2 {
		t.Fatalf("unexpected listener count: got %d want 2", len(snap))
	}
	if snap[0] != a || snap[1] != b {
		t.Fatal("unexpected listener order in snapshot")
	}

	ls.Delete(a)

	snap = ls.Load()
	if len(snap) != 1 || snap[0] != b {
		t.Fatalf("unexpected snapshot after delete: %v", snap)
	}

	ls.Delete(a)
	snap = ls.Load()
	if len(snap) != 1 || snap[0] != b {
		t.Fatalf("unexpected snapshot after deleting missing listener: %v", snap)
	}
}

func TestListenersBroadcast(t *testing.T) {
	ls := NewListeners[string, int]()
	a := ls.Add("a", 1)
	b := ls.Add("b", 1)

	if rejected := ls.Broadcast(7); rejected != 0 {
		t.Fatalf("unexpected rejected count: got %d want 0", rejected)
	}

	select {
	case got := <-a.Inbox:
		if got != 7 {
			t.Fatalf("unexpected message for a: got %d want 7", got)
		}
	default:
		t.Fatal("expected message for a")
	}

	select {
	case got := <-b.Inbox:
		if got != 7 {
			t.Fatalf("unexpected message for b: got %d want 7", got)
		}
	default:
		t.Fatal("expected message for b")
	}
}

func TestListenersBroadcastRejectsFullInbox(t *testing.T) {
	ls := NewListeners[string, int]()
	a := ls.Add("a", 1)
	b := ls.Add("b", 1)

	a.Inbox <- 1

	rejected := ls.Broadcast(2)
	if rejected != 1 {
		t.Fatalf("unexpected rejected count: got %d want 1", rejected)
	}

	select {
	case got := <-a.Inbox:
		if got != 1 {
			t.Fatalf("unexpected retained message for a: got %d want 1", got)
		}
	default:
		t.Fatal("expected retained message for a")
	}

	select {
	case got := <-b.Inbox:
		if got != 2 {
			t.Fatalf("unexpected broadcast message for b: got %d want 2", got)
		}
	default:
		t.Fatal("expected message for b")
	}
}
