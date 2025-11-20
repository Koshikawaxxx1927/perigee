package scorer

import (
	"time"
	"sort"

	"github.com/libp2p/go-libp2p/core/peer"
)

// SelectWorstPeer retrieves snapshot, computes φ(v) for all peers,
// and returns the peerID with the highest (worst) score.
// If no peers exist, returns (zeroID, false).
// SelectWorstPeers retrieves snapshot, computes φ(v) for all peers,
// and returns the top N peerIDs with the highest (worst) scores.
// If no peers exist, returns (nil, false).
func (s *ObservationStore) SelectWorstPeers(
    Delta time.Duration,
    outbound []peer.ID,
    topN int,
) ([]peer.ID, bool) {

    // 1. Snapshot and clear
    snapshot := s.SnapshotAndClear(outbound)
    if len(snapshot) == 0 {
        return []peer.ID{}, false
    }

    // 2. Compute all peer scores
    scoreMap := CompleteScoresForAllPeers(snapshot, Delta)

    // 3. Convert to slice for sorting
    type kv struct {
        Peer  peer.ID
        Score time.Duration
    }
    var scores []kv
    for pid, score := range scoreMap {
        scores = append(scores, kv{Peer: pid, Score: score})
    }

    // 4. Sort by score descending (worst first)
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].Score > scores[j].Score
    })

    // 5. Determine actual number to return
    if topN > len(scores) {
        topN = len(scores)
    }
    if topN <= 0 {
        return []peer.ID{}, false
    }

    // 6. Extract worst peers
    worst := make([]peer.ID, topN)
    for i := 0; i < topN; i++ {
        worst[i] = scores[i].Peer
    }
    return worst, true
}

