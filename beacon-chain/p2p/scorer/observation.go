package scorer

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ObservationStore collects receipts in a round.
// Internally: map[blockID]map[peer.ID]time.Time
type ObservationStore struct {
	mu       sync.RWMutex
	receipts map[string]map[peer.ID]time.Time
}

// NewObservationStore creates a new store.
func NewObservationStore() *ObservationStore {
	return &ObservationStore{
		receipts: make(map[string]map[peer.ID]time.Time),
	}
}

// RecordReceipt records that a peer delivered a block at local time ts.
// If the same peer/block is recorded multiple times, earliest timestamp is kept.
func (s *ObservationStore) RecordReceipt(blockID string, pid peer.ID, ts time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.receipts[blockID]
	if !ok {
		m = make(map[peer.ID]time.Time)
		s.receipts[blockID] = m
	}
	prev, ok := m[pid]
	if !ok || ts.Before(prev) {
		m[pid] = ts
	}
}

// SnapshotAndClear returns a snapshot of current receipts and clears internal buffer.
// Returned map is the internal structure (caller MUST NOT modify it).
func (s *ObservationStore) SnapshotAndClear() map[string]map[peer.ID]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.receipts
	s.receipts = make(map[string]map[peer.ID]time.Time)
	return snap
}
