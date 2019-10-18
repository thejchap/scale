package store

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/msmedes/scale/internal/pkg/keyspace"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	key1 := keyspace.StringToKey("hello")
	val1 := []byte("world")

	store.Set(key1, val1)
	got := store.Get(key1)

	if !bytes.Equal(got, val1) {
		t.Errorf("expected %x got %x", val1, got)
	}

	keys := store.Keys()

	if !keyspace.Equal(keys[0], key1) {
		t.Errorf("expected keys[0] to be %x got %x", key1, keys[0])
	}

	store.Del(key1)
	got = store.Get(key1)

	if !bytes.Equal(got, nil) {
		t.Errorf("expected %x got %x", val1, got)
	}
}

func TestMemoryStoreThreadSafety(t *testing.T) {
	store := NewMemoryStore()
	key := keyspace.StringToKey("key")

	var w sync.WaitGroup
	w.Add(3)

	for i := 0; i < 3; i++ {
		go func(j int) {
			defer w.Done()
			time.Sleep(time.Duration(j) * 10 * time.Millisecond)
			str := fmt.Sprintf("val-%d", j)
			store.Set(key, []byte(str))
		}(i)
	}

	w.Wait()

	got := string(store.Get(key))

	if got != "val-2" {
		t.Errorf("expected val-2 got %s", got)
	}
}
