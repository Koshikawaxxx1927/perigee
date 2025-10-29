package vcs

import (
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
)

func TestRandomBetween0And1(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := RandomBetween0And1()
		if val < 0.0 || val > 1.0 {
			t.Errorf("Expected value between 0 and 1, got %f", val)
		}
	}
}

func TestRandomBetweenMinus1And1(t *testing.T) {
	for i := 0; i < 100; i++ {
		val := randomBetweenMinus1And1()
		if val < -1.0 || val > 1.0 {
			t.Errorf("Expected value between -1 and 1, got %f", val)
		}
	}
}

func TestRandomBetween0And1_Distribution(t *testing.T) {
	const buckets = 10
	const samples = 1000
	bucketCounts := make([]int, buckets)

	for i := 0; i < samples; i++ {
		val := RandomBetween0And1()
		if val < 0.0 || val > 1.0 {
			t.Errorf("Expected value between 0 and 1, got %f", val)
		}
		bucketIndex := int(val * float64(buckets))
		if bucketIndex == buckets {
			bucketIndex = buckets - 1
		}
		bucketCounts[bucketIndex]++
	}

	for i, count := range bucketCounts {
		if count == 0 {
			t.Errorf("Bucket %d is empty, distribution may not be uniform", i)
		}
	}
}

func TestInverseLog(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.0, 1.0 / math.Log10(1.0+1.0)},   // Basic case
		{0.0, 1.0 / math.Log10(1.0+0.0)},   // Edge case: x = 0
		{10.0, 1.0 / math.Log10(1.0+10.0)}, // Larger input
		{-0.5, 1.0 / math.Log10(1.0-0.5)},  // Edge case: negative input, should handle or validate
	}

	for _, tt := range tests {
		result := inverseLog(tt.input)
		if math.Abs(result-tt.expected) > 1e-9 {
			t.Errorf("inverseLog(%v) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{7, 7, 7},
		{-1, 0, -1},
	}

	for _, tt := range tests {
		result := Min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Min(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestRandomShuffle(t *testing.T) {
	// Seed for deterministic test
	rand.Seed(42)

	original := []*peer.AddrInfo{
		{ID: "A"},
		{ID: "B"},
		{ID: "C"},
		{ID: "D"},
	}

	// Copy original
	copySlice := make([]*peer.AddrInfo, len(original))
	copy(copySlice, original)

	RandomShuffle(copySlice)

	// Same length
	if len(copySlice) != len(original) {
		t.Errorf("Expected same length after shuffle, got %d", len(copySlice))
	}

	// Same elements (order can differ)
	originalIDs := make(map[string]bool)
	for _, p := range original {
		originalIDs[string(p.ID)] = true
	}
	for _, p := range copySlice {
		if !originalIDs[string(p.ID)] {
			t.Errorf("Shuffled peer %s not in original list", p.ID)
		}
	}

	// Should not be in same order (most of the time)
	if reflect.DeepEqual(original, copySlice) {
		t.Logf("Warning: shuffle result is same as original (may occasionally happen)")
	}
}
