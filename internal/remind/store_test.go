package remind

import (
	"os"
	"sync"
	"testing"
)

func TestStoreSavesChanges(t *testing.T) {
	s := NewStore(t.TempDir())
	err := s.Update(func(l *List) error {
		l.Add(Reminder{Message: "stretch", Due: wed(9, 0), Repeat: &Repeat{Days: workDays, At: 9 * 60}})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	s.View(func(l *List) error {
		r, err := l.Get(1)
		if err != nil || r.Message != "stretch" || !r.Due.Equal(wed(9, 0)) || len(r.Repeat.Days) != 5 {
			t.Errorf("loaded %+v, %v", r, err)
		}
		return nil
	})
}

func TestStoreUpdatesDontCollide(t *testing.T) {
	s := NewStore(t.TempDir())
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if err := s.Update(func(l *List) error { l.Add(Reminder{}); return nil }); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	s.View(func(l *List) error {
		if len(l.Reminders) != 20 || l.NextID != 21 {
			t.Errorf("got %d reminders and next ID %d; want 20 and 21", len(l.Reminders), l.NextID)
		}
		return nil
	})
}

func TestStoreKeepsDamagedFile(t *testing.T) {
	s := NewStore(t.TempDir())
	os.WriteFile(s.Path(), []byte("{oops"), 0o600)
	if err := s.Update(func(*List) error { return nil }); err == nil {
		t.Fatal("expected an error for a damaged file")
	}
	if data, _ := os.ReadFile(s.Path()); string(data) != "{oops" {
		t.Error("the damaged file was overwritten")
	}
}
