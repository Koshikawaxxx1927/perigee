package scorer

import (
	"math"
	"math/rand"
	"testing"
	"time"
	"sync"
	"github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p/core/peer"
)

// // --- ユニットテスト: ComputeRelativeSamplesFromSnapshot ---
// func TestComputeRelativeSamplesFromSnapshot(t *testing.T) {
// 	base := time.Unix(0, 0)
// 	snap := map[string]map[peer.ID]time.Time{
// 		"b1": {
// 			peer.ID("p1"): base,
// 			peer.ID("p2"): base.Add(100 * time.Millisecond),
// 			peer.ID("p3"): base.Add(300 * time.Millisecond),
// 		},
// 		"b2": {
// 			peer.ID("p2"): base.Add(500 * time.Millisecond),
// 			peer.ID("p1"): base.Add(700 * time.Millisecond),
// 		},
// 	}
// 	per := ComputeRelativeSamplesFromSnapshot(snap)

// 	if len(per[peer.ID("p1")]) != 2 {
// 		t.Fatalf("expected p1 len 2, got %d", len(per[peer.ID("p1")]))
// 	}
// 	if len(per[peer.ID("p2")]) != 2 {
// 		t.Fatalf("expected p2 len 2, got %d", len(per[peer.ID("p2")]))
// 	}
// 	if len(per[peer.ID("p3")]) != 1 {
// 		t.Fatalf("expected p3 len 1, got %d", len(per[peer.ID("p3")]))
// 	}
// 	if math.Abs(per[peer.ID("p3")][0]-0.3) > 1e-6 {
// 		t.Fatalf("expected p3 sample ~0.3 got %v", per[peer.ID("p3")][0])
// 	}
// }

// --- ユニットテスト: percentilePositional90 ---
func TestPercentilePositional90(t *testing.T) {
	arr := []float64{5, 1, 9, 0, 3, 7, 2, 8, 4, 6}
	cp := append([]float64(nil), arr...)
	v, err := percentilePositional90(cp)
	if err != nil {
		t.Fatalf("percentile error: %v", err)
	}
	if v != 9 {
		t.Fatalf("expected 9 got %v", v)
	}
}

// --------------------------
// 単体テスト: ComputeRelativeSamplesFromSnapshot
// --------------------------
func TestComputeRelativeSamplesFromSnapshot(t *testing.T) {
	base := time.Unix(0, 0)
	snap := map[string]map[peer.ID]time.Time{
		"b1": {
			peer.ID("p1"): base,
			peer.ID("p2"): base.Add(100 * time.Millisecond),
			peer.ID("p3"): base.Add(300 * time.Millisecond),
		},
		"b2": {
			peer.ID("p2"): base.Add(500 * time.Millisecond),
			peer.ID("p1"): base.Add(700 * time.Millisecond),
		},
	}
	per := ComputeRelativeSamplesFromSnapshot(snap)

	if len(per[peer.ID("p1")]) != 2 {
		t.Fatalf("expected p1 len 2, got %d", len(per[peer.ID("p1")]))
	}
	if len(per[peer.ID("p2")]) != 2 {
		t.Fatalf("expected p2 len 2, got %d", len(per[peer.ID("p2")]))
	}
	if len(per[peer.ID("p3")]) != 1 {
		t.Fatalf("expected p3 len 1, got %d", len(per[peer.ID("p3")]))
	}
	if math.Abs(per[peer.ID("p3")][0]-0.3) > 1e-6 {
		t.Fatalf("expected p3 sample ~0.3 got %v", per[peer.ID("p3")][0])
	}
}

// --------------------------
// 単体テスト: ComputeLCBAndUCB
// --------------------------
func TestComputeLCBAndUCB_Basic(t *testing.T) {
	rand.Seed(12345)
	ucb := NewUCBScorer(90.0, 0.125, 5)

	base := time.Unix(0, 0)
	round1 := map[string]map[peer.ID]time.Time{
		"b1": {peer.ID("A"): base, peer.ID("B"): base.Add(100 * time.Millisecond)},
		"b2": {peer.ID("A"): base.Add(10 * time.Millisecond), peer.ID("B"): base.Add(200 * time.Millisecond)},
	}
	round2 := map[string]map[peer.ID]time.Time{
		"b3": {peer.ID("A"): base.Add(300 * time.Millisecond), peer.ID("B"): base.Add(700 * time.Millisecond)},
		"b4": {peer.ID("A"): base.Add(400 * time.Millisecond), peer.ID("B"): base.Add(900 * time.Millisecond)},
	}

	ucb.AddRoundSamples(round1)
	ucb.AddRoundSamples(round2)

	bounds, err := ucb.ComputeLCBAndUCB()
	if err != nil {
		t.Fatalf("ComputeLCBAndUCB error: %v", err)
	}
	if !(bounds[peer.ID("A")].LCB < bounds[peer.ID("B")].LCB) {
		t.Fatalf("expected LCB(A) < LCB(B); got A=%v B=%v", bounds[peer.ID("A")].LCB, bounds[peer.ID("B")].LCB)
	}
}

// // --------------------------
// // 単体テスト: Selector with LCB置換
// // --------------------------
func TestSelectWithLCBReplacement_SingleReplacement(t *testing.T) {
	sel := NewSelector()

	tests := []struct {
		name        string
		bounds      map[peer.ID]lcbucb
		current     []peer.ID
		wantReplace peer.ID
	}{
		{
			name: "replacement_needed",
			bounds: map[peer.ID]lcbucb{
				peer.ID("p_worst"): {LCB: 5.0, UCB: 7.0},
				peer.ID("p_good"):  {LCB: 1.0, UCB: 2.0},
				peer.ID("p_mid"):   {LCB: 2.0, UCB: 3.0},
			},
			current:     []peer.ID{peer.ID("p_worst"), peer.ID("p_good")},
			wantReplace: peer.ID("p_worst"),
		},
		{
			name: "no_replacement_needed",
			bounds: map[peer.ID]lcbucb{
				peer.ID("p1"): {LCB: 1.0, UCB: 5.0},
				peer.ID("p2"): {LCB: 2.0, UCB: 6.0},
			},
			current:     []peer.ID{peer.ID("p1"), peer.ID("p2")},
			wantReplace: "",
		},
		{
			name:        "empty_current",
			bounds:      map[peer.ID]lcbucb{},
			current:     []peer.ID{},
			wantReplace: "",
		},
		{
			name: "all_same_bounds",
			bounds: map[peer.ID]lcbucb{
				peer.ID("p1"): {LCB: 2.0, UCB: 2.0},
				peer.ID("p2"): {LCB: 2.0, UCB: 2.0},
			},
			current:     []peer.ID{peer.ID("p1"), peer.ID("p2")},
			wantReplace: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sel.SelectWithLCBReplacement(tt.bounds, tt.current)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantReplace {
				t.Errorf("expected replaced peer %v, got %v", tt.wantReplace, got)
			}
		})
	}
}

// --------------------------
// End-to-End テスト
// --------------------------
func TestEndToEnd_ObservationStore_UCB_Selector(t *testing.T) {
	rand.Seed(42)
	base := time.Unix(0, 0)

	// 1. ObservationStore にサンプル記録
	store := NewObservationStore()
	store.RecordReceipt("b1", peer.ID("p1"), base)
	store.RecordReceipt("b1", peer.ID("p2"), base.Add(100*time.Millisecond))
	store.RecordReceipt("b2", peer.ID("p1"), base.Add(200*time.Millisecond))
	store.RecordReceipt("b2", peer.ID("p2"), base.Add(500*time.Millisecond))

	snapshot := store.SnapshotAndClear()

	// 2. UCBScorer で LCB/UCB 計算
	ucb := NewUCBScorer(90.0, 0.125, 5) // 秒単位
	ucb.AddRoundSamples(snapshot)

	bounds, err := ucb.ComputeLCBAndUCB()
	if err != nil {
		t.Fatalf("ComputeLCBAndUCB error: %v", err)
	}

	// p1 の方が速い（LCB が小さい）ことを確認
	if bounds[peer.ID("p1")].LCB >= bounds[peer.ID("p2")].LCB {
		t.Fatalf("expected LCB(p1) < LCB(p2), got p1=%v, p2=%v", bounds[peer.ID("p1")].LCB, bounds[peer.ID("p2")].LCB)
	}

	// 3. Selector で置換が必要な peer を取得
	sel := NewSelector()
	current := []peer.ID{peer.ID("p1"), peer.ID("p2")}
	replaced, err := sel.SelectWithLCBReplacement(bounds, current)
	if err != nil {
		t.Fatalf("Selector error: %v", err)
	}

	// 4. 遅い p2 が置換対象になることを確認
	if replaced != peer.ID("p2") {
		t.Fatalf("expected p2 to be replaced, got %v", replaced)
	}
}

//////////////
// --- Helpers for tests -----------------------------------------------------

// a simple no-op disconnect that records calls
type recordingDisconnect struct {
	mu       sync.Mutex
	called   []peer.ID
	failOnce bool
}

func (r *recordingDisconnect) fn(pid peer.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.called = append(r.called, pid)
	if r.failOnce {
		// simulate an error once
		r.failOnce = false
		return &mockErr{"simulated disconnect error"}
	}
	return nil
}

type mockErr struct{ s string }

func (m *mockErr) Error() string { return m.s }

// --- Tests ----------------------------------------------------------------

func TestNewPeerSelectorManagerDefaults(t *testing.T) {
	// We can pass nil for complex dependencies for this simple constructor test,
	// because we only assert defaults on the manager struct itself.
	logger := logrus.NewEntry(logrus.New())
	m := NewPeerSelectorManager(nil, nil, nil, func(pid peer.ID) error { return nil }, logger)

	if m.roundInterval <= 0 {
		t.Fatalf("expected positive roundInterval, got %v", m.roundInterval)
	}
	if m.blockDuration <= 0 {
		t.Fatalf("expected positive blockDuration, got %v", m.blockDuration)
	}
	if m.minPeers <= 0 {
		t.Fatalf("expected positive minPeers, got %d", m.minPeers)
	}
	if m.maxReplace <= 0 {
		t.Fatalf("expected positive maxReplace, got %d", m.maxReplace)
	}
	if m.blockList == nil {
		t.Fatalf("expected blockList to be initialized")
	}
}

func TestAddAndIsBlocked(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	m := NewPeerSelectorManager(nil, nil, nil, func(pid peer.ID) error { return nil }, logger)

	// make up a peer ID
	var pid peer.ID = peer.ID("peer-A")
	now := time.Now()
	until := now.Add(1 * time.Minute)

	// initially not blocked
	if m.isBlocked(pid, now) {
		t.Fatalf("peer should not be blocked initially")
	}

	// add blocked and check
	m.addBlocked(pid, until)
	if !m.isBlocked(pid, now) {
		t.Fatalf("peer should be blocked after addBlocked")
	}

	// after expiry it should be removed (isBlocked does lazy expiry)
	later := until.Add(1 * time.Second)
	if m.isBlocked(pid, later) {
		t.Fatalf("peer should not be blocked after expiry")
	}

	// ensure it was removed from map
	m.blockMu.RLock()
	_, exists := m.blockList[pid]
	m.blockMu.RUnlock()
	if exists {
		t.Fatalf("expired peer should have been removed from blockList")
	}
}

func TestRunStopsImmediatelyWhenStopClosed(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	// create manager with short interval so test is fast
	m := NewPeerSelectorManager(nil, nil, nil, func(pid peer.ID) error { return nil }, logger)
	m.roundInterval = 10 * time.Millisecond

	stopCh := make(chan struct{})
	// close immediately so Run should exit quickly
	close(stopCh)

	// provide a dummy getCurrentPeers to satisfy signature
	getCurrentPeers := func() []peer.ID { return []peer.ID{} }

	// run in goroutine and ensure it returns quickly
	done := make(chan struct{})
	go func() {
		m.Run(getCurrentPeers, stopCh)
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("Run did not stop promptly after stopCh closed")
	}
}