package scorer

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// SelectWorstPeer retrieves snapshot, computes φ(v) for all peers,
// and returns the peerID with the highest (worst) score.
// If no peers exist, returns (zeroID, false).
func (s *ObservationStore) SelectWorstPeer(
	Delta time.Duration,
	outbound []peer.ID,
) (peer.ID, bool) {

	// 1. Snapshot and clear
	snapshot := s.SnapshotAndClear(outbound)

	if len(snapshot) == 0 {
		return peer.ID(""), false
	}

	// 2. Compute all peer scores
	scoreMap := CompleteScoresForAllPeers(snapshot, Delta)

	// 3. Select the peer with the highest score φ(v)
	var worst peer.ID
	var worstScore time.Duration
	found := false

	for pid, score := range scoreMap {
		if !found || score > worstScore {
			found = true
			worst = pid
			worstScore = score
		}
	}

	return worst, found
}
