package scorer

import (
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	field_params "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
)

// CalculateScoreForPeer computes φ(v) exactly as defined in the Peri paper.
//
//	φ(v) = (1/|M_v|) * Σ_m min( T_v(m) - T(m), Δ )
//
// snapshot = ObservationStore.SnapshotAndClear() の結果
func CalculateScoreForPeer(
	snapshot map[[field_params.RootLength]byte]map[peer.ID]time.Time,
	target peer.ID,
	Delta time.Duration,
) time.Duration {

	var sum time.Duration
	count := 0

	for _, peerMap := range snapshot {
		// ブロックごとにループ

		// T_v(m) が存在するか
		Tv := peerMap[target]

		// Zero value なら Delta を使う
		if Tv.IsZero() {
			sum += Delta
			count++
			continue
		}

		// T(m) = 最速の配信
		var Tm time.Time
		for _, t := range peerMap {
			if !t.IsZero() && (Tm.IsZero() || t.Before(Tm)) {
				Tm = t
			}
		}

		// diff = T_v(m) - T(m)
		diff := Tv.Sub(Tm)
		if diff < 0 {
			diff = 0
		}

		// min( diff, Δ )
		if diff > Delta {
			diff = Delta
		}

		sum += diff
		count++
	}

	if count == 0 {
		return Delta // M_v が空 → is_excused
	}

	return sum / time.Duration(count)
}

// CompleteScoresForAllPeers calculates φ(v) for every peer observed
// in the snapshot produced by ObservationStore.SnapshotAndClear().
func CompleteScoresForAllPeers(
	snapshot map[[field_params.RootLength]byte]map[peer.ID]time.Time,
	Delta time.Duration,
) map[peer.ID]time.Duration {

	scoreMap := make(map[peer.ID]time.Duration)

	// Collect all peers that appear at least once
	peerSet := make(map[peer.ID]struct{})

	for _, peerMap := range snapshot {
		for pid := range peerMap {
			peerSet[pid] = struct{}{}
		}
	}

	// Compute φ(v) for every peer found
	for pid := range peerSet {
		score := CalculateScoreForPeer(snapshot, pid, Delta)
		scoreMap[pid] = score
	}

	// ログ出力
	for pid, score := range scoreMap {
		log.Printf("peer: %s, score: %v\n", pid, score)
	}

	return scoreMap
}
