package scorer

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	field_params "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
)

// ObservationStore collects block delivery timestamps in a round.
// Internally:
//
//	map[blockRoot([32]byte)] → map[peer.ID]time.Time
type ObservationStore struct {
	mu       sync.RWMutex
	receipts map[[field_params.RootLength]byte]map[peer.ID]time.Time
}

// NewObservationStore creates an empty store and registers all outbound peers.
func NewObservationStore() *ObservationStore {

	return &ObservationStore{
		receipts: make(map[[field_params.RootLength]byte]map[peer.ID]time.Time),
	}
}

// RecordBlkReceipt records that peer p delivered block blk at timestamp ts.
// The earliest timestamp wins.
func (s *ObservationStore) RecordBlkReceipt(
	p peer.ID,
	blkRoot [field_params.RootLength]byte,
	ts time.Time,
) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create entry for this block root if needed
	peerMap, ok := s.receipts[blkRoot]
	if !ok {
		peerMap = make(map[peer.ID]time.Time)
		s.receipts[blkRoot] = peerMap
	}

	// Store earliest timestamp only
	prev, ok := peerMap[p]
	if !ok || ts.Before(prev) {
		peerMap[p] = ts
	}

	return nil
}

// SnapshotAndClear returns all stored receipts and resets the store.
// It also ensures that all outbound peers are represented, even if they didn't send any block.
func (s *ObservationStore) SnapshotAndClear(outboundPeers []peer.ID) map[[field_params.RootLength]byte]map[peer.ID]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	snap := s.receipts
	newSnap := make(map[[field_params.RootLength]byte]map[peer.ID]time.Time)

	// Ensure all outbound peers are present in each block
	for blk, peerMap := range snap {
		newPeerMap := make(map[peer.ID]time.Time)
		for _, pid := range outboundPeers {
			if ts, ok := peerMap[pid]; ok {
				newPeerMap[pid] = ts
			} else {
				// 記録なしピアは存在だけ確保（スコア計算時に Δ を足す）
				newPeerMap[pid] = time.Time{} // zero value を使う
			}
		}
		newSnap[blk] = newPeerMap
	}

	// クリア
	s.receipts = make(map[[field_params.RootLength]byte]map[peer.ID]time.Time)

	return newSnap
}
