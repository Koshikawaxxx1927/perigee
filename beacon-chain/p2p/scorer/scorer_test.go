package scorer

import (
	"math"
	"math/rand"
	"testing"
	"time"

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
	ucb := NewUCBScorer(90.0, 125, 5)

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

// --------------------------
// 単体テスト: Selector with LCB置換
// --------------------------
func TestSelectorWithLCBReplacement_AllCases(t *testing.T) {
	rand.Seed(42)
	sel := NewSelector()

	t.Run("replacement_needed", func(t *testing.T) {
		bounds := map[peer.ID]lcbucb{
			peer.ID("p_worst"): {LCB: 5.0, UCB: 7.0},
			peer.ID("p_good"):  {LCB: 1.0, UCB: 2.0},
			peer.ID("p_mid"):   {LCB: 2.0, UCB: 3.0},
		}
		current := []peer.ID{peer.ID("p_worst"), peer.ID("p_good")}
		candidates := []peer.ID{peer.ID("p_new")}

		selected, err := sel.SelectWithLCBReplacement(bounds, current, 2, candidates)
		if err != nil {
			t.Fatalf("SelectWithLCBReplacement failed: %v", err)
		}

		// p_worst は置換されているはず
		for _, p := range selected {
			if p == peer.ID("p_worst") {
				t.Fatalf("expected p_worst to be replaced, still selected")
			}
		}
		foundNew := false
		for _, p := range selected {
			if p == peer.ID("p_new") {
				foundNew = true
			}
		}
		if !foundNew {
			t.Fatalf("expected p_new to be selected, got %v", selected)
		}
	})

	t.Run("no_replacement_needed", func(t *testing.T) {
		bounds := map[peer.ID]lcbucb{
			peer.ID("p1"): {LCB: 1.0, UCB: 5.0},
			peer.ID("p2"): {LCB: 2.0, UCB: 6.0},
		}
		current := []peer.ID{peer.ID("p1"), peer.ID("p2")}
		candidates := []peer.ID{peer.ID("p_new")}

		selected, err := sel.SelectWithLCBReplacement(bounds, current, 2, candidates)
		if err != nil {
			t.Fatalf("SelectWithLCBReplacement failed: %v", err)
		}

		// 置換不要なので p1 と p2 が残るはず
		found1, found2 := false, false
		for _, p := range selected {
			if p == peer.ID("p1") {
				found1 = true
			}
			if p == peer.ID("p2") {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Fatalf("expected p1 and p2 to remain, got %v", selected)
		}
	})

	t.Run("dout_smaller_than_total", func(t *testing.T) {
		bounds := map[peer.ID]lcbucb{
			peer.ID("p1"): {LCB: 1.0, UCB: 2.0},
			peer.ID("p2"): {LCB: 2.0, UCB: 3.0},
			peer.ID("p3"): {LCB: 3.0, UCB: 4.0},
		}
		current := []peer.ID{peer.ID("p1"), peer.ID("p2"), peer.ID("p3")}
		candidates := []peer.ID{peer.ID("p_new")}

		selected, err := sel.SelectWithLCBReplacement(bounds, current, 2, candidates)
		if err != nil {
			t.Fatalf("SelectWithLCBReplacement failed: %v", err)
		}
		if len(selected) != 2 {
			t.Fatalf("expected selected length 2, got %d", len(selected))
		}
	})

	t.Run("multiple_candidates", func(t *testing.T) {
		bounds := map[peer.ID]lcbucb{
			peer.ID("p1"): {LCB: 5.0, UCB: 6.0},
			peer.ID("p2"): {LCB: 1.0, UCB: 2.0},
		}
		current := []peer.ID{peer.ID("p1"), peer.ID("p2")}
		candidates := []peer.ID{peer.ID("p_new1"), peer.ID("p_new2")}

		selected, err := sel.SelectWithLCBReplacement(bounds, current, 2, candidates)
		if err != nil {
			t.Fatalf("SelectWithLCBReplacement failed: %v", err)
		}

		// p1 は置換され、候補のどちらかが入る
		found := false
		for _, p := range selected {
			if p == peer.ID("p_new1") || p == peer.ID("p_new2") {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected one candidate to be selected, got %v", selected)
		}
	})

	t.Run("no_candidates", func(t *testing.T) {
		bounds := map[peer.ID]lcbucb{
			peer.ID("p1"): {LCB: 5.0, UCB: 6.0},
			peer.ID("p2"): {LCB: 1.0, UCB: 2.0},
		}
		current := []peer.ID{peer.ID("p1"), peer.ID("p2")}
		candidates := []peer.ID{}

		selected, err := sel.SelectWithLCBReplacement(bounds, current, 2, candidates)
		if err != nil {
			t.Fatalf("SelectWithLCBReplacement failed: %v", err)
		}

		// 候補なしでも dout <= len(current) なので元のノードが残る
		found1, found2 := false, false
		for _, p := range selected {
			if p == peer.ID("p1") {
				found1 = true
			}
			if p == peer.ID("p2") {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Fatalf("expected p1 and p2 to remain, got %v", selected)
		}
	})
}

func TestEndToEndIntegration(t *testing.T) {
	rand.Seed(7)

	// --- 1. ObservationStoreにサンプルを記録 ---
	store := NewObservationStore()
	base := time.Unix(0, 0)

	// 近隣ノードの受信タイムスタンプを模擬
	store.RecordReceipt("b1", peer.ID("p1"), base)
	store.RecordReceipt("b1", peer.ID("p2"), base.Add(100*time.Millisecond))
	store.RecordReceipt("b2", peer.ID("p1"), base.Add(200*time.Millisecond))
	store.RecordReceipt("b2", peer.ID("p2"), base.Add(500*time.Millisecond))

	// --- 2. スナップショットを取得してクリア ---
	snap := store.SnapshotAndClear()

	// --- 3. UCBScorerでLCB/UCBを計算 ---
	ucb := NewUCBScorer(90.0, 125, 2)
	ucb.AddRoundSamples(snap)
	bounds, err := ucb.ComputeLCBAndUCB()
	if err != nil {
		t.Fatalf("ComputeLCBAndUCB error: %v", err)
	}

	// p1の方が速い（LCB小さい）ことを確認
	if bounds[peer.ID("p1")].LCB >= bounds[peer.ID("p2")].LCB {
		t.Fatalf("expected p1 LCB < p2 LCB, got p1=%v p2=%v",
			bounds[peer.ID("p1")].LCB, bounds[peer.ID("p2")].LCB)
	}

	// --- 4. Selectorで選択 ---
	sel := NewSelector()
	currentOutgoing := []peer.ID{peer.ID("p1"), peer.ID("p2")}
	candidatePool := []peer.ID{peer.ID("p3")}
	dout := 2

	selected, err := sel.SelectWithLCBReplacement(bounds, currentOutgoing, dout, candidatePool)
	if err != nil {
		t.Fatalf("SelectWithLCBReplacement failed: %v", err)
	}

	// --- 5. 選択結果の確認 ---
	if len(selected) != dout {
		t.Fatalf("expected %d peers selected, got %d", dout, len(selected))
	}

	// p1は必ず選ばれる（速い）
	foundP1 := false
	for _, pid := range selected {
		if pid == peer.ID("p1") {
			foundP1 = true
			break
		}
	}
	if !foundP1 {
		t.Fatalf("expected p1 to be selected, got %v", selected)
	}

	// maxLCB > minUCB の場合、最も遅い p2 が p3 に置換されているか確認
	maxLCB := -1.0
	var worstPeer peer.ID
	for _, pid := range selected {
		if b, ok := bounds[pid]; ok && b.LCB > maxLCB {
			maxLCB = b.LCB
			worstPeer = pid
		}
	}
	if worstPeer == peer.ID("p2") {
		t.Fatalf("expected p2 to be replaced by candidate, got %v", selected)
	}
}
