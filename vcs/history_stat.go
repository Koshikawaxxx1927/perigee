package vcs

import (
	"fmt"
	"math"
	"sort"
)

// HistoryStat struct for float64 values
type HistoryStat struct {
	maxDataPoint   int
	historyCounter int
	q              []float64
	median         float64
	absMedianDev   float64
	UPDATE_ROUND   int
}

// NewHistoryStat constructor
func NewHistoryStat(maxDataPoint int) *HistoryStat {
	return &HistoryStat{
		maxDataPoint:   maxDataPoint,
		historyCounter: 0,
		q:              make([]float64, 0, maxDataPoint),
		median:         0,
		absMedianDev:   0,
		UPDATE_ROUND:   int(math.Min(50, float64(maxDataPoint))),
	}
}

// Observe adds new data point and recalculates statistics
func (hs *HistoryStat) Observe(newData float64) {
	// Remove the oldest data if the queue is full
	if len(hs.q) == hs.maxDataPoint {
		hs.q = hs.q[1:]
	}
	// Add new data to the queue
	hs.q = append(hs.q, newData)

	hs.historyCounter++
	// Recalculate the median and absolute median deviation at intervals
	if hs.historyCounter <= hs.UPDATE_ROUND || hs.historyCounter%hs.UPDATE_ROUND == 1 {
		tmp := make([]float64, len(hs.q))
		copy(tmp, hs.q)
		sort.Float64s(tmp)

		// Compute the median
		mid := len(tmp) / 2
		if len(tmp)%2 == 0 {
			hs.median = (tmp[mid-1] + tmp[mid]) / 2
		} else {
			hs.median = tmp[mid]
		}

		// Compute the absolute median deviation
		hs.absMedianDev = 0
		for _, v := range tmp {
			hs.absMedianDev += math.Abs(v - hs.median)
		}
		hs.absMedianDev /= float64(len(tmp))
	}
}

// GetMedian returns the current median
func (hs *HistoryStat) GetMedian() float64 {
	if len(hs.q) == 0 {
		return 0
	}
	return hs.median
}

// GetMedianDev returns the current absolute median deviation
func (hs *HistoryStat) GetMedianDev() float64 {
	if len(hs.q) == 0 {
		return 0
	}
	return hs.absMedianDev
}

// CollectedDataNum returns the number of collected data points
func (hs *HistoryStat) CollectedDataNum() int {
	return len(hs.q)
}

// IsFull checks if the queue is full
func (hs *HistoryStat) IsFull() bool {
	return len(hs.q) == hs.maxDataPoint
}

// Show prints the median and deviation
func (hs *HistoryStat) Show() {
	fmt.Printf("median = %.2f, dev = %.2f\n", hs.median, hs.absMedianDev)

	tmp := make([]float64, len(hs.q))
	copy(tmp, hs.q)
	sort.Float64s(tmp)

	fmt.Println("Sorted data:", tmp)
}
