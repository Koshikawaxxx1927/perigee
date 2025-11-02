package scorer

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// UCBScorer holds history and parameters for UCB scoring.
//
// Observations (samples) are expected in seconds (float64), i.e. produced by
// time.Since(...).Seconds() or by ComputeRelativeSamplesFromSnapshot (which returns seconds).
//
// To match the original C++ bias (bias = 125.0 * sqrt(log(n)/(2n)) where C++ used ms),
// set biasScale = 125.0 / 1000.0 = 125 when constructing.
// Example:
//   ucb := NewUCBScorer(90.0, 125, 5)
type UCBScorer struct {
	Percentile    float64 // keep for compatibility; p=90 expected
	BiasScale     float64 // bias scale in same unit as samples (seconds). For C++ parity use 125 for seconds.
	HistoryRounds int     // number of past rounds to keep (0 = keep all)

	mu      sync.RWMutex
	history map[peer.ID][][]float64 // peerID -> list of rounds (each round: slice of samples in seconds)
}

// NewUCBScorer constructs a UCBScorer.
//
// percentile (e.g. 90.0), biasScale (unit: same as samples: seconds), historyRounds
func NewUCBScorer(percentile, biasScale float64, historyRounds int) *UCBScorer {
	return &UCBScorer{
		Percentile:    percentile,
		BiasScale:     biasScale,
		HistoryRounds: historyRounds,
		history:       make(map[peer.ID][][]float64),
	}
}

// AddRoundSamples merges snapshot into history.
// snapshot: map[blockID]map[peer.ID]time.Time (uses ComputeRelativeSamplesFromSnapshot to get seconds)
func (s *UCBScorer) AddRoundSamples(snapshot map[string]map[peer.ID]time.Time) {
	roundSamples := ComputeRelativeSamplesFromSnapshot(snapshot) // map[peer.ID][]float64 (seconds)
	s.mu.Lock()
	defer s.mu.Unlock()
	for pid, samples := range roundSamples {
		if len(samples) == 0 {
			continue
		}
		s.history[pid] = append(s.history[pid], samples)
		if s.HistoryRounds > 0 && len(s.history[pid]) > s.HistoryRounds {
			start := len(s.history[pid]) - s.HistoryRounds
			s.history[pid] = s.history[pid][start:]
		}
	}
}

// ComputeLCBAndUCB computes for each peer:
//  - per90 by positional index pos = int(n*0.9) (C++ style)
//  - bias = BiasScale * sqrt(log(n) / (2*n))
// Returns map[peer.ID]{LCB, UCB} in seconds.
// Peers with no samples get LCB=+Inf, UCB=-Inf.
func (s *UCBScorer) ComputeLCBAndUCB() (map[peer.ID]lcbucb, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[peer.ID]lcbucb, len(s.history))
	for pid, rounds := range s.history {
		merged := make([]float64, 0)
		for _, r := range rounds {
			merged = append(merged, r...)
		}
		n := len(merged)
		if n == 0 {
			out[pid] = lcbucb{LCB: math.Inf(1), UCB: -math.Inf(1)}
			continue
		}
		// copy before selection/sort to avoid mutating history
		tmp := append([]float64(nil), merged...)
		var per90 float64
		// heuristic: use quickSelect for larger slices, sort for small
		if n > 200 {
			v, err := quickSelect(tmp, int(float64(n)*0.9))
			if err != nil {
				// fallback
				sort.Float64s(tmp)
				pos := int(float64(n) * 0.9)
				if pos >= n {
					pos = n - 1
				}
				per90 = tmp[pos]
			} else {
				per90 = v
			}
		} else {
			sort.Float64s(tmp)
			pos := int(float64(n) * 0.9)
			if pos >= n {
				pos = n - 1
			}
			per90 = tmp[pos]
		}
		// bias formula (seconds)
		denom := float64(n)
		if denom <= 1 {
			denom = 2.0 // avoid log(1)=0 and /0 issues
		}
		bias := s.BiasScale * math.Sqrt(math.Log(denom)/(2.0*denom))
		lcb := per90 - bias
		ucb := per90 + bias
		if lcb < 0 {
			lcb = 0
		}
		out[pid] = lcbucb{LCB: lcb, UCB: ucb}
	}
	return out, nil
}

// GetHistoryForPeer returns deep copy for inspection/testing.
func (s *UCBScorer) GetHistoryForPeer(pid peer.ID) [][]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rounds, ok := s.history[pid]
	if !ok {
		return nil
	}
	out := make([][]float64, len(rounds))
	for i, r := range rounds {
		cp := append([]float64(nil), r...)
		out[i] = cp
	}
	return out
}
