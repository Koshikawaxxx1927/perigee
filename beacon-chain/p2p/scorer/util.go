package scorer

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ComputeRelativeSamplesFromSnapshot:
// given snapshot: map[blockID]map[peer.ID]time.Time
// For each block b, t_min = min_u t_{b,u,v}.
// For each peer u in that block, record relative = (t_{b,u,v} - t_min) in seconds (float64).
// Returns map[peer.ID][]float64 (relative seconds)
func ComputeRelativeSamplesFromSnapshot(snap map[string]map[peer.ID]time.Time) map[peer.ID][]float64 {
	perPeer := make(map[peer.ID][]float64)
	for _, m := range snap {
		if len(m) == 0 {
			continue
		}
		var tmin time.Time
		first := true
		for _, ts := range m {
			if first || ts.Before(tmin) {
				tmin = ts
				first = false
			}
		}
		for pid, ts := range m {
			relSec := ts.Sub(tmin).Seconds() // seconds as float64
			if relSec < 0 {
				relSec = 0
			}
			perPeer[pid] = append(perPeer[pid], relSec)
		}
	}
	return perPeer
}

// quickSelect: partition-based selection to find k-th smallest (0-indexed).
// It rearranges slice in-place. Returns value at index k after partitioning.
// Implementation uses randomized pivot.
// Note: this mutates the slice passed in.
func quickSelect(a []float64, k int) (float64, error) {
	n := len(a)
	if n == 0 {
		return 0, errors.New("empty slice")
	}
	if k < 0 || k >= n {
		return 0, errors.New("k out of range")
	}
	lo := 0
	hi := n - 1
	for {
		if lo == hi {
			return a[lo], nil
		}
		pivotIdx := lo + rand.Intn(hi-lo+1)
		pivot := a[pivotIdx]
		// move pivot to end
		a[pivotIdx], a[hi] = a[hi], a[pivotIdx]
		store := lo
		for i := lo; i < hi; i++ {
			if a[i] < pivot {
				a[store], a[i] = a[i], a[store]
				store++
			}
		}
		// move pivot to its final place
		a[store], a[hi] = a[hi], a[store]

		if k == store {
			return a[store], nil
		} else if k < store {
			hi = store - 1
		} else {
			lo = store + 1
		}
	}
}

// percentilePositional90 returns the element at pos = int(n * 0.9).
// It uses quickSelect (mutates input).
func percentilePositional90(arr []float64) (float64, error) {
	n := len(arr)
	if n == 0 {
		return 0, errors.New("empty slice")
	}
	pos := int(float64(n) * 0.9)
	if pos >= n {
		pos = n - 1
	}
	return quickSelect(arr, pos)
}

// quickSort wrapper for float64 slice (utility fallback)
func quickSort(a []float64) {
	if len(a) <= 1 {
		return
	}
	p := a[len(a)/2]
	i := 0
	j := len(a) - 1
	for i <= j {
		for a[i] < p {
			i++
		}
		for a[j] > p {
			j--
		}
		if i <= j {
			a[i], a[j] = a[j], a[i]
			i++
			j--
		}
	}
	if j > 0 {
		quickSort(a[:j+1])
	}
	if i < len(a) {
		quickSort(a[i:])
	}
}

// PercentileLinear kept for compatibility/testing (not used by default)
func PercentileLinear(data []float64, p float64) (float64, error) {
	if len(data) == 0 {
		return 0, errors.New("empty")
	}
	if p < 0 || p > 100 {
		return 0, errors.New("p out of range")
	}
	cpy := append([]float64(nil), data...)
	quickSort(cpy)
	if len(cpy) == 1 {
		return cpy[0], nil
	}
	pos := (p / 100.0) * float64(len(cpy)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return cpy[lo], nil
	}
	frac := pos - float64(lo)
	return cpy[lo]*(1-frac) + cpy[hi]*frac, nil
}
