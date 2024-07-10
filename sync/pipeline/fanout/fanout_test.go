package fanout

import (
	"context"
	"sync"
	"testing"
)

func TestFanout_Do(t *testing.T) {
	ca := New("cache", WithWorker(1), WithBuffer(1024))
	var run bool
	wg := sync.WaitGroup{}
	wg.Add(1)
	_ = ca.Do(context.Background(), func(_ context.Context) {
		defer wg.Done()
		run = true
		panic("error")
	})
	wg.Wait()
	t.Log("not panic")
	if !run {
		t.Error("expect run be true")
	}
}

func TestFanout_Close(t *testing.T) {
	ca := New("cache", WithWorker(1), WithBuffer(1024))
	_ = ca.Close()
	err := ca.Do(context.Background(), func(_ context.Context) {})
	if err == nil {
		t.Error("expect get err")
	}
}
