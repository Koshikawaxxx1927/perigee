package vcs

import (
	"math"
	"testing"
)

func TestNewHistoryStat(t *testing.T) {
	hs := NewHistoryStat(5)

	if hs.maxDataPoint != 5 {
		t.Errorf("expected maxDataPoint to be 5, got %d", hs.maxDataPoint)
	}

	if hs.historyCounter != 0 {
		t.Errorf("expected historyCounter to be 0, got %d", hs.historyCounter)
	}

	if len(hs.q) != 0 {
		t.Errorf("expected initial queue length to be 0, got %d", len(hs.q))
	}

	if hs.median != 0 {
		t.Errorf("expected median to be 0, got %f", hs.median)
	}

	if hs.absMedianDev != 0 {
		t.Errorf("expected absMedianDev to be 0, got %f", hs.absMedianDev)
	}

	if hs.UPDATE_ROUND != int(math.Min(50, float64(hs.maxDataPoint))) {
		t.Errorf("expected UPDATE_ROUND to be %d, got %d", int(math.Min(50, float64(hs.maxDataPoint))), hs.UPDATE_ROUND)
	}
}

func TestObserveAndMedian(t *testing.T) {
	hs := NewHistoryStat(3)

	hs.Observe(1.0)
	hs.Observe(2.0)
	hs.Observe(3.0)

	// Check median after 3 observations
	expectedMedian := 2.0
	if hs.GetMedian() != expectedMedian {
		t.Errorf("expected median to be %f, got %f", expectedMedian, hs.GetMedian())
	}

	// Add another data point (which should replace the oldest one)
	hs.Observe(4.0)

	// Check median after observing the new data point
	expectedMedian = 3.0
	if hs.GetMedian() != expectedMedian {
		t.Errorf("expected median to be %f, got %f", expectedMedian, hs.GetMedian())
	}
}

func TestAbsMedianDev(t *testing.T) {
	hs := NewHistoryStat(3)

	hs.Observe(1.0)
	hs.Observe(2.0)
	hs.Observe(3.0)

	// Check absolute median deviation after 3 observations
	median := 2.0
	expectedAbsMedianDev := (math.Abs(1.0-median) + math.Abs(2.0-median) + math.Abs(3.0-median)) / 3.0
	if hs.GetMedianDev() != expectedAbsMedianDev {
		t.Errorf("expected absolute median deviation to be %f, got %f", expectedAbsMedianDev, hs.GetMedianDev())
	}

	// Add another data point and check again
	hs.Observe(4.0)
	median = 3.0
	expectedAbsMedianDev = (math.Abs(2.0-median) + math.Abs(3.0-median) + math.Abs(4.0-median)) / 3.0
	if hs.GetMedianDev() != expectedAbsMedianDev {
		t.Errorf("expected absolute median deviation to be %f, got %f", expectedAbsMedianDev, hs.GetMedianDev())
	}
}

func TestCollectedDataNum(t *testing.T) {
	hs := NewHistoryStat(3)

	hs.Observe(1.0)
	if hs.CollectedDataNum() != 1 {
		t.Errorf("expected collected data number to be 1, got %d", hs.CollectedDataNum())
	}

	hs.Observe(2.0)
	hs.Observe(3.0)
	if hs.CollectedDataNum() != 3 {
		t.Errorf("expected collected data number to be 3, got %d", hs.CollectedDataNum())
	}

	hs.Observe(4.0)
	if hs.CollectedDataNum() != 3 { // Should still be 3 due to maxDataPoint limit
		t.Errorf("expected collected data number to be 3, got %d", hs.CollectedDataNum())
	}
}

func TestIsFull(t *testing.T) {
	hs := NewHistoryStat(3)

	if hs.IsFull() {
		t.Errorf("expected IsFull to return false, but got true")
	}

	hs.Observe(1.0)
	hs.Observe(2.0)
	hs.Observe(3.0)

	if !hs.IsFull() {
		t.Errorf("expected IsFull to return true, but got false")
	}

	hs.Observe(4.0) // Should replace the oldest data, still full
	if !hs.IsFull() {
		t.Errorf("expected IsFull to return true after adding 4th data point, but got false")
	}
}

func TestHistoryStatShow(t *testing.T) {
	hs := NewHistoryStat(3)

	hs.Observe(1.0)
	hs.Observe(2.0)
	hs.Observe(3.0)

	// Just run Show() to ensure no panic occurs
	hs.Show()
}

func TestObserveSingleDataPoint(t *testing.T) {
	hs := NewHistoryStat(1) // maxDataPoint = 1
	hs.Observe(10.0)

	if hs.GetMedian() != 10.0 {
		t.Errorf("expected median to be 10.0, got %f", hs.GetMedian())
	}

	if hs.GetMedianDev() != 0 {
		t.Errorf("expected absolute median deviation to be 0, got %f", hs.GetMedianDev())
	}
}

func TestObserveWithLargeAndNegativeValues(t *testing.T) {
	hs := NewHistoryStat(5)
	hs.Observe(-100.0)
	hs.Observe(0.0)
	hs.Observe(100.0)

	expectedMedian := 0.0
	expectedAbsMedianDev := (100.0 + 0.0 + 100.0) / 3.0

	if hs.GetMedian() != expectedMedian {
		t.Errorf("expected median to be %f, got %f", expectedMedian, hs.GetMedian())
	}

	if math.Abs(hs.GetMedianDev()-expectedAbsMedianDev) > 1e-9 {
		t.Errorf("expected absolute median deviation to be %f, got %f", expectedAbsMedianDev, hs.GetMedianDev())
	}
}

func TestShowWithEmptyQueue(t *testing.T) {
	hs := NewHistoryStat(3)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Show() panicked with empty queue: %v", r)
		}
	}()
	hs.Show() // Ensure no panic occurs
}
