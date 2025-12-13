package analytics

import (
	"shiroxy/pkg/logger"
	"sync"
	"testing"
	"time"
)

func getTestLogger() *logger.Logger {
	logHandler, _ := logger.StartLogger(nil)
	return logHandler
}

func TestStartAnalytics(t *testing.T) {
	var wg sync.WaitGroup
	logHandler := getTestLogger()

	analytics, err := StartAnalytics(100*time.Millisecond, logHandler, &wg)
	if err != nil {
		t.Fatalf("failed to start analytics: %v", err)
	}
	defer analytics.StopAnalytics()

	if analytics == nil {
		t.Fatal("analytics configuration is nil")
	}

	if analytics.RequestAnalytics == nil {
		t.Error("RequestAnalytics channel is nil")
	}
	if analytics.ReadAnalyticsData == nil {
		t.Error("ReadAnalyticsData channel is nil")
	}
}

func TestCollectAnalytics(t *testing.T) {
	var wg sync.WaitGroup
	logHandler := getTestLogger()

	analytics, err := StartAnalytics(100*time.Millisecond, logHandler, &wg)
	if err != nil {
		t.Fatalf("failed to start analytics: %v", err)
	}
	defer analytics.StopAnalytics()

	// Wait for first collection
	time.Sleep(200 * time.Millisecond)

	data, err := analytics.ReadAnalytics(false)
	if err != nil {
		t.Fatalf("failed to read analytics: %v", err)
	}

	if data == nil {
		t.Fatal("analytics data is nil")
	}

	// Verify memory stats are collected (values >= 0 is acceptable)
	if data.Memory_SYS < 0 {
		t.Error("Memory_SYS should be non-negative")
	}
	if data.GC_Count < 0 {
		t.Error("GC_Count should be non-negative")
	}
	// CPU usage can be 0 during tests, so just check it's a valid number
	if data.CPU_Usage < 0 {
		t.Error("CPU_Usage should be non-negative")
	}
}

func TestReadAnalyticsForcedCollection(t *testing.T) {
	var wg sync.WaitGroup
	logHandler := getTestLogger()

	analytics, err := StartAnalytics(10*time.Second, logHandler, &wg) // Long interval
	if err != nil {
		t.Fatalf("failed to start analytics: %v", err)
	}
	defer analytics.StopAnalytics()

	// Force immediate collection
	data, err := analytics.ReadAnalytics(true)
	if err != nil {
		t.Fatalf("failed to read analytics: %v", err)
	}

	// Give it time to collect
	time.Sleep(100 * time.Millisecond)

	data, err = analytics.ReadAnalytics(false)
	if data == nil {
		t.Fatal("forced analytics collection failed")
	}
}

func TestUpdateTriggerInterval(t *testing.T) {
	var wg sync.WaitGroup
	logHandler := getTestLogger()

	analytics, err := StartAnalytics(1*time.Second, logHandler, &wg)
	if err != nil {
		t.Fatalf("failed to start analytics: %v", err)
	}
	defer analytics.StopAnalytics()

	// Update interval
	err = analytics.UpdateTriggerInterval(500 * time.Millisecond)
	if err != nil {
		t.Fatalf("failed to update trigger interval: %v", err)
	}

	// Verify it doesn't panic
	time.Sleep(100 * time.Millisecond)
}

func TestStopAnalytics(t *testing.T) {
	var wg sync.WaitGroup
	logHandler := getTestLogger()

	analytics, err := StartAnalytics(100*time.Millisecond, logHandler, &wg)
	if err != nil {
		t.Fatalf("failed to start analytics: %v", err)
	}

	analytics.StopAnalytics()

	// Wait for goroutine to exit
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success - goroutine exited
	case <-time.After(2 * time.Second):
		t.Error("analytics goroutine did not stop in time")
	}
}

func TestBToMb(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected uint64
	}{
		{0, 0},
		{1024 * 1024, 1},
		{2 * 1024 * 1024, 2},
		{1024, 0}, // Less than 1MB
		{10 * 1024 * 1024, 10},
	}

	for _, tt := range tests {
		result := bToMb(tt.bytes)
		if result != tt.expected {
			t.Errorf("bToMb(%d) = %d; want %d", tt.bytes, result, tt.expected)
		}
	}
}

func TestConcurrentReadAnalytics(t *testing.T) {
	var wg sync.WaitGroup
	logHandler := getTestLogger()

	analytics, err := StartAnalytics(100*time.Millisecond, logHandler, &wg)
	if err != nil {
		t.Fatalf("failed to start analytics: %v", err)
	}
	defer analytics.StopAnalytics()

	// Wait for first collection
	time.Sleep(200 * time.Millisecond)

	// Concurrent reads should not cause race conditions
	var readWg sync.WaitGroup
	for i := 0; i < 10; i++ {
		readWg.Add(1)
		go func() {
			defer readWg.Done()
			_, err := analytics.ReadAnalytics(false)
			if err != nil {
				t.Errorf("concurrent read failed: %v", err)
			}
		}()
	}

	readWg.Wait()
}
