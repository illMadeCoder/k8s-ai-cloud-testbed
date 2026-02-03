package main

import (
	"sync"
	"testing"
)

func TestMemStore_PutGet(t *testing.T) {
	s := NewMemStore()
	s.Put("k1", []byte("v1"))

	v, ok := s.Get("k1")
	if !ok {
		t.Fatal("expected key k1 to exist")
	}
	if string(v) != "v1" {
		t.Fatalf("expected v1, got %s", v)
	}
}

func TestMemStore_GetMissing(t *testing.T) {
	s := NewMemStore()
	_, ok := s.Get("nope")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestMemStore_Delete(t *testing.T) {
	s := NewMemStore()
	s.Put("k1", []byte("v1"))

	if !s.Delete("k1") {
		t.Fatal("expected delete to return true")
	}
	if s.Delete("k1") {
		t.Fatal("expected second delete to return false")
	}
	if _, ok := s.Get("k1"); ok {
		t.Fatal("expected key to be gone after delete")
	}
}

func TestMemStore_Len(t *testing.T) {
	s := NewMemStore()
	if s.Len() != 0 {
		t.Fatal("expected empty store")
	}
	s.Put("a", []byte("1"))
	s.Put("b", []byte("2"))
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
	s.Delete("a")
	if s.Len() != 1 {
		t.Fatalf("expected 1, got %d", s.Len())
	}
}

func TestMemStore_Overwrite(t *testing.T) {
	s := NewMemStore()
	s.Put("k", []byte("first"))
	s.Put("k", []byte("second"))

	v, _ := s.Get("k")
	if string(v) != "second" {
		t.Fatalf("expected second, got %s", v)
	}
	if s.Len() != 1 {
		t.Fatal("overwrite should not change length")
	}
}

func TestMemStore_Concurrent(t *testing.T) {
	s := NewMemStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + n%26))
			s.Put(key, []byte("val"))
			s.Get(key)
			s.Len()
		}(i)
	}
	wg.Wait()
}

func BenchmarkMemStore_Put(b *testing.B) {
	s := NewMemStore()
	val := []byte("benchmark-value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Put("key", val)
	}
}

func BenchmarkMemStore_Get(b *testing.B) {
	s := NewMemStore()
	s.Put("key", []byte("benchmark-value"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get("key")
	}
}
