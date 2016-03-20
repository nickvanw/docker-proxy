package dockerproxy

import (
	"testing"
	"time"
)

func TestSingleDebounce(t *testing.T) {
	ch := make(chan struct{})
	debounce := debounceChannel(100*time.Millisecond, ch)
	ch <- struct{}{}
	start := time.Now()
	select {
	case <-time.After(150 * time.Millisecond):
		t.Fatalf("debounce channel took too long to send")
	case <-debounce:
		if time.Since(start) < 100*time.Millisecond {
			t.Fatalf("debounce channel didn't wait long enough to send")
		}
	}
	t.Logf("debouncing wait: %v", time.Since(start))
}

func TestBounceOnce(t *testing.T) {
	ch := make(chan struct{})
	debounce := debounceChannel(100*time.Millisecond, ch)
	go func() { ch <- struct{}{}; time.Sleep(50 * time.Millisecond); ch <- struct{}{} }()
	start := time.Now()
	select {
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("debounce channel took too long to send")
	case <-debounce:
		if time.Since(start) < 150*time.Millisecond {
			t.Fatalf("debounce channel didn't wait long enough to send")
		}
	}
	t.Logf("debouncing wait: %v", time.Since(start))
}

func TestMultipleBouncing(t *testing.T) {
	ch := make(chan struct{})
	debounce := debounceChannel(50*time.Millisecond, ch)
	go func() {
		ch <- struct{}{}
		for x := 0; x < 5; x++ {
			time.Sleep(30 * time.Millisecond)
			ch <- struct{}{}
		}
	}()
	start := time.Now()
	select {
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("debounce channel took too long to send")
	case <-debounce:
		if time.Since(start) < 200*time.Millisecond {
			t.Fatalf("debounce channel didn't wait long enough to send")
		}
	}
	t.Logf("debouncing wait: %v", time.Since(start))
}
