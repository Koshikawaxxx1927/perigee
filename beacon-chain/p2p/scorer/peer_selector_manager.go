package scorer

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	field_params "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
)

// PeerSelectorManager は ObservationStore をラップし、
// 複数ラウンドでのピアスコア計算や最悪ピア選択を簡単に扱えるようにする。
type PeerSelectorManager struct {
	store *ObservationStore
	Delta time.Duration
}

// NewPeerSelectorManager creates a new manager with a set of outbound peers and Δ.
func NewPeerSelectorManager(Delta time.Duration) *PeerSelectorManager {
	return &PeerSelectorManager{
		store: NewObservationStore(),
		Delta: Delta,
	}
}

// RecordBlockDelivery records that peer p delivered block blk at timestamp ts.
func (m *PeerSelectorManager) RecordBlockDelivery(p peer.ID, blk [field_params.RootLength]byte, ts time.Time) {
	m.store.RecordBlkReceipt(p, blk, ts)
}

// SelectWorstPeer は現在のスナップショットで最もスコアの高いピアを返す
func (m *PeerSelectorManager) SelectWorstPeer(outboundPeers []peer.ID) (peer.ID, bool) {
	// ObservationStore の SelectWorstPeer を呼び出す
	return m.store.SelectWorstPeer(m.Delta, outboundPeers)
}
