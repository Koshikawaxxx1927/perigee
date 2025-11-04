// peer_selector_manager.go
package scorer

import (
	"sync"
	"time"
	"encoding/hex"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus" // Prysm のロギングに合わせる
)

type PeerSelectorManager struct {
	obs        *ObservationStore
	ucb        *UCBScorer
	sel        *Selector
	disconnect func(pid peer.ID) error // DI: s.Disconnect
	// blockList: pid -> unblock time
	blockMu   sync.RWMutex
	blockList map[peer.ID]time.Time

	// 設定
	roundInterval time.Duration
	blockDuration time.Duration
	minPeers      int // 最低維持ピア数
	maxReplace    int // 1回のラウンドで最大置換数（dout相当）

	logger *logrus.Entry
}

// NewPeerSelectorManager constructs manager
func NewPeerSelectorManager(disconnect func(pid peer.ID) error, logger *logrus.Entry) *PeerSelectorManager {
	return &PeerSelectorManager{
		obs:           NewObservationStore(),
		ucb:           NewUCBScorer(90.0, 125.0, 5), // C++互換パラメータ
		sel:           NewSelector(),
		disconnect:    disconnect,
		blockList:     make(map[peer.ID]time.Time),
		roundInterval: 120 * time.Second, // 任意調整
		blockDuration: 5 * time.Minute,  // 切断後ブロック期間
		minPeers:      2,                // 例: 最低 8 ピアは残す
		maxReplace:    1,                // 例: 1ラウンドで最大4件を置換
		logger:        logger,
	}
}

// Run starts periodic selection loop. Call in its own goroutine.
func (m *PeerSelectorManager) Run(getCurrentPeers func() []peer.ID, stopCh <-chan struct{}) {
	ticker := time.NewTicker(m.roundInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.runOneRound(getCurrentPeers)
		case <-stopCh:
			m.logger.Info("PeerSelectorManager stopping")
			return
		}
	}
}

func (m *PeerSelectorManager) runOneRound(getCurrentPeers func() []peer.ID) {
	// 1) snapshot samples and add to ucb
	m.logger.Info("Starting peer selection round")
	snap := m.obs.SnapshotAndClear()
	m.ucb.AddRoundSamples(snap)

	// 2) compute bounds
	m.logger.Info("Computing UCB/LBC bounds for peer selection")
	bounds, err := m.ucb.ComputeLCBAndUCB()
	if err != nil {
		m.logger.WithError(err).Warn("ComputeLCBAndUCB failed")
		return
	}

	// 3) build current peer set (filter out blocked)
	m.logger.Info("Building current peer set for selection")
	current := getCurrentPeers()
	filtered := make([]peer.ID, 0, len(current))
	now := time.Now()
	for _, pid := range current {
		if m.isBlocked(pid, now) {
			// skip blocked peers from consideration for replacement
			continue
		}
		filtered = append(filtered, pid)
	}

	// 4) enforce minPeers guard
	if len(filtered) <= m.minPeers {
		m.logger.Infof("Skipping replacement: only %d peers (min %d)", len(filtered), m.minPeers)
		return
	}

	// 5) attempt up to maxReplace replacements
	replaceCount := 0
	for replaceCount < m.maxReplace {
		replaced, err := m.sel.SelectWithLCBReplacement(bounds, filtered)
		if err != nil {
			m.logger.WithError(err).Warn("Selector failed")
			break
		}
		if replaced == "" {
			// nothing to replace
			break
		}

		// Ensure we don't accidentally remove below minPeers
		if len(filtered)-1 < m.minPeers {
			m.logger.Info("Not replacing to avoid falling below minPeers")
			break
		}

		// call Disconnect (synchronous here)
		if err := m.disconnect(replaced); err != nil {
			m.logger.WithField("peer", replaced).WithError(err).Warn("Failed to disconnect peer")
			// If disconnect failed, avoid marking blocked; but continue
			break
		}
		m.logger.WithField("peer", replaced).Info("Disconnected peer")
		// record to blockList
		m.addBlocked(replaced, now.Add(m.blockDuration))
		m.logger.WithField("peer", replaced).Infof("Disconnected and blocked until %s", now.Add(m.blockDuration).Format(time.RFC3339))

		// remove replaced from filtered slice
		newFiltered := make([]peer.ID, 0, len(filtered)-1)
		for _, p := range filtered {
			if p != replaced {
				newFiltered = append(newFiltered, p)
			}
		}
		filtered = newFiltered

		replaceCount++
	}
}

// isBlocked checks blockList and clears expired entries lazily.
func (m *PeerSelectorManager) isBlocked(pid peer.ID, now time.Time) bool {
	m.blockMu.RLock()
	until, ok := m.blockList[pid]
	m.blockMu.RUnlock()
	if !ok {
		return false
	}
	if now.After(until) {
		// expire
		m.blockMu.Lock()
		delete(m.blockList, pid)
		m.blockMu.Unlock()
		return false
	}
	return true
}

func (m *PeerSelectorManager) addBlocked(pid peer.ID, until time.Time) {
	m.blockMu.Lock()
	m.blockList[pid] = until
	m.blockMu.Unlock()
}

func (m *PeerSelectorManager) CanConnect(pid peer.ID) bool {
    return !m.isBlocked(pid, time.Now())
}

// AddBlock records that we observed blockRoot from peer `from` at time `ts`.
// Internally it converts the root to a string and delegates to ObservationStore.RecordReceipt.
// It's safe to call concurrently. If the manager's ObservationStore is nil, this is a no-op.
func (m *PeerSelectorManager) AddBlock(blockRoot [32]byte, from peer.ID, ts time.Time) {
	if m == nil || m.obs == nil {
		return
	}
	id := hex.EncodeToString(blockRoot[:])
	m.obs.RecordReceipt(id, from, ts)
}

// AddBlockNow is a convenience wrapper that records the observation with the current time.
func (m *PeerSelectorManager) AddBlockNow(blockRoot [32]byte, from peer.ID) {
	m.AddBlock(blockRoot, from, time.Now())
}