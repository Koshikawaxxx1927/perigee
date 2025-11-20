package scorer

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

func TestPeerSelectorManager_RecordAndWorst(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	peers := []peer.ID{p1, p2, p3}
	Delta := 10 * time.Second

	manager := NewPeerSelectorManager(Delta)

	root1 := fakeRoot(1)
	root2 := fakeRoot(2)

	// p1,p2 は配信済み
	manager.RecordBlockDelivery(p1, root1, time.Unix(100, 0))
	manager.RecordBlockDelivery(p2, root1, time.Unix(90, 0))
	manager.RecordBlockDelivery(p1, root2, time.Unix(200, 0))
	manager.RecordBlockDelivery(p2, root2, time.Unix(195, 0))
	// p3 は未配信（zero value）

	replacedNumber := 1
	worst, ok := manager.SelectWorstPeers(peers, replacedNumber)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	// p3 は未配信なので Δ が加算され、最悪スコアになる
	if worst[0] != p3 {
		t.Fatalf("expected p3 as worst peer, got %v", worst)
	}
}

func TestPeerSelectorManager_SelectWorstPeer_Empty(t *testing.T) {
	manager := NewPeerSelectorManager(10 * time.Second)
	replacedNumber := 1
	worst, ok := manager.SelectWorstPeers([]peer.ID{}, replacedNumber)
	if ok {
		t.Fatalf("expected no worst peer, got %v", worst)
	}
}

func TestPeerSelectorManager_AllFastDelivery(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	peers := []peer.ID{p1, p2}
	Delta := 10 * time.Second

	manager := NewPeerSelectorManager(Delta)
	root := fakeRoot(1)

	// 両者同じタイミングで配信
	manager.RecordBlockDelivery(p1, root, time.Unix(100, 0))
	manager.RecordBlockDelivery(p2, root, time.Unix(100, 0))

	replacedNumber := 1
	worst, ok := manager.SelectWorstPeers(peers, replacedNumber)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	// 両者同じタイミング → どちらでも正しい
	if worst[0] != p1 && worst[0] != p2 {
		t.Fatalf("expected p1 or p2 as worst peer, got %v", worst)
	}
}

func TestPeerSelectorManager_MultipleRounds(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	peers := []peer.ID{p1, p2, p3}
	Delta := 20 * time.Second

	manager := NewPeerSelectorManager(Delta)

	// ラウンド1
	root1 := fakeRoot(1)
	manager.RecordBlockDelivery(p1, root1, time.Unix(100, 0))
	manager.RecordBlockDelivery(p2, root1, time.Unix(90, 0))

	replacedNumber := 1
	worst1, _ := manager.SelectWorstPeers(peers, replacedNumber)
	if worst1[0] != p3 {
		t.Fatalf("expected p3 as worst in round 1, got %v", worst1)
	}

	// ラウンド2
	root2 := fakeRoot(2)
	manager.RecordBlockDelivery(p1, root2, time.Unix(200, 0))
	manager.RecordBlockDelivery(p3, root2, time.Unix(210, 0))

	replacedNumber = 1
	worst2, _ := manager.SelectWorstPeers(peers, replacedNumber)
	if worst2[0] != p2 {
		t.Fatalf("expected p2 as worst in round 2, got %v", worst2)
	}
}

func TestPeerSelectorManager_PartialDelivery(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	peers := []peer.ID{p1, p2, p3}
	Delta := 10 * time.Second

	manager := NewPeerSelectorManager(Delta)

	root := fakeRoot(1)
	// p1: 最速, p2: 遅延, p3: 未配信
	manager.RecordBlockDelivery(p1, root, time.Unix(100, 0))
	manager.RecordBlockDelivery(p2, root, time.Unix(105, 0))

	replacedNumber := 1
	worst, _ := manager.SelectWorstPeers(peers, replacedNumber)
	if worst[0] != p3 {
		t.Fatalf("expected p3 as worst due to no delivery, got %v", worst)
	}
}

// // ブロックが一切観測されなかったときのテスト (基本的にはない)
// func TestPeerSelectorManager_TieScores(t *testing.T) {
// 	p1 := mustPeer()
// 	p2 := mustPeer()
// 	peers := []peer.ID{p1, p2}
// 	Delta := 10 * time.Second

// 	manager := NewPeerSelectorManager(peers, Delta)

// 	// 両者未配信 → スコア Δ で同点
// 	worst, ok := manager.SelectWorstPeers()
// 	fmt.Printf("p1 %v\n", p1)
// 	fmt.Printf("p2 %v\n", p2)
// 	fmt.Printf("worst %v\n", worst)
// 	if !ok {
// 		t.Fatalf("expected a worst peer, got none")
// 	}
// 	if worst != p1 && worst != p2 {
// 		t.Fatalf("expected p1 or p2 as worst peer, got %v", worst)
// 	}
// }

func TestSelectWorstPeer_ReplacedNumber2Plus(t *testing.T) {
    // ピアを複数生成
    p1 := mustPeer()
    p2 := mustPeer()
    p3 := mustPeer()
    p4 := mustPeer()
    outbound := []peer.ID{p1, p2, p3, p4}

    store := NewObservationStore()
    manager := &PeerSelectorManager{
        store: store,
        Delta: 10 * time.Second,
    }
    root := fakeRoot(1)

    t.Run("AllFastDelivery_SameScore", func(t *testing.T) {
        // 全員同じ受信時間 → スコア同点
        store.RecordBlkReceipt(p1, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p2, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p3, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p4, root, time.Unix(100, 0))

        replacedNumber := 3
        worst, ok := manager.SelectWorstPeers(outbound, replacedNumber)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != replacedNumber {
            t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
        }

        // 全員同点なので順序は問わない
        validSet := map[peer.ID]bool{p1: true, p2: true, p3: true, p4: true}
        for _, pid := range worst {
            if !validSet[pid] {
                t.Errorf("unexpected peer %v returned", pid)
            }
        }
    })

    t.Run("DifferentDeliveryTimes", func(t *testing.T) {
        // 受信時間に差をつける → スコアに差が出る
        store.RecordBlkReceipt(p1, root, time.Unix(100, 0)) // 最速
        store.RecordBlkReceipt(p2, root, time.Unix(105, 0))
        store.RecordBlkReceipt(p3, root, time.Unix(110, 0))
        store.RecordBlkReceipt(p4, root, time.Unix(120, 0)) // 最遅

        replacedNumber := 2
        worst, ok := manager.SelectWorstPeers(outbound, replacedNumber)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != replacedNumber {
            t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
        }

        // スコア降順で上位2つを確認
        expectedSet := map[peer.ID]bool{p4: true, p3: true} // p4最悪, p3次
        for _, pid := range worst {
            if !expectedSet[pid] {
                t.Errorf("unexpected peer %v returned", pid)
            }
        }
    })

    t.Run("TieAmongTopPeers", func(t *testing.T) {
        // p3, p4 同点で最悪
        store.RecordBlkReceipt(p1, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p2, root, time.Unix(105, 0))
        store.RecordBlkReceipt(p3, root, time.Unix(120, 0))
        store.RecordBlkReceipt(p4, root, time.Unix(120, 0))

        replacedNumber := 2
        worst, ok := manager.SelectWorstPeers(outbound, replacedNumber)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != replacedNumber {
            t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
        }

        // 同点なので p3, p4 どちらの組み合わせでもOK
        validSet := map[peer.ID]bool{p3: true, p4: true}
        for _, pid := range worst {
            if !validSet[pid] {
                t.Errorf("unexpected peer %v returned", pid)
            }
        }
    })

    t.Run("TopNExceedsPeerCount", func(t *testing.T) {
        // topN > ピア数 → 全員返る
        store.RecordBlkReceipt(p1, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p2, root, time.Unix(105, 0))

        replacedNumber := 5 // ピア数より多い
        worst, ok := manager.SelectWorstPeers(outbound[:2], replacedNumber)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != 2 {
            t.Fatalf("expected 2 peers returned, got %d: %v", len(worst), worst)
        }
    })
}

func TestSelectWorstPeer_WithMissingBlocks(t *testing.T) {
    // ピアを複数生成
    p1 := mustPeer()
    p2 := mustPeer()
    p3 := mustPeer()
    outbound := []peer.ID{p1, p2, p3}

    store := NewObservationStore()
    manager := &PeerSelectorManager{
        store: store,
        Delta: 10 * time.Second,
    }
    root1 := fakeRoot(1)
    root2 := fakeRoot(2)
    root3 := fakeRoot(3)

    // --------------------------------------------------
    // 1回目のブロック配信
    store.RecordBlkReceipt(p1, root1, time.Unix(100, 0))
    store.RecordBlkReceipt(p2, root1, time.Unix(101, 0))
    // p3 は未受信

    replacedNumber := 2
    worst, ok := manager.SelectWorstPeers(outbound, replacedNumber)
    if !ok {
        t.Fatalf("expected ok=true")
    }
    if len(worst) != replacedNumber {
        t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
    }

    // p3 はブロック未受信 → 最悪スコアになっているはず
    validSet := map[peer.ID]bool{p3: true, p2: true, p1: true}
    for _, pid := range worst {
        if !validSet[pid] {
            t.Errorf("unexpected peer %v returned", pid)
        }
    }

    // --------------------------------------------------
    // 2回目のブロック配信（p1 が遅延）
    store.RecordBlkReceipt(p1, root2, time.Unix(120, 0)) // 遅い
    store.RecordBlkReceipt(p2, root2, time.Unix(110, 0))
    store.RecordBlkReceipt(p3, root2, time.Unix(115, 0))

    replacedNumber = 2
    worst, ok = manager.SelectWorstPeers(outbound, replacedNumber)
    if !ok {
        t.Fatalf("expected ok=true")
    }
    if len(worst) != replacedNumber {
        t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
    }

    // p1 が遅延、p3 はまだ未受信または遅延 → 上位2ピアに入るはず
    expectedSet := map[peer.ID]bool{p1: true, p3: true}
    for _, pid := range worst {
        if !expectedSet[pid] {
            t.Errorf("unexpected peer %v returned", pid)
        }
    }

    // --------------------------------------------------
    // 3回目のブロック配信（p2 が未受信）
    store.RecordBlkReceipt(p1, root3, time.Unix(130, 0))
    store.RecordBlkReceipt(p3, root3, time.Unix(140, 0))
    // p2 は未受信

    replacedNumber = 2
    worst, ok = manager.SelectWorstPeers(outbound, replacedNumber)
    if !ok {
        t.Fatalf("expected ok=true")
    }
    if len(worst) != replacedNumber {
        t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
    }

    // p2 が未受信 → 最悪スコア
    expectedSet = map[peer.ID]bool{p2: true, p3: true} // p2最悪、p3次
    for _, pid := range worst {
        if !expectedSet[pid] {
            t.Errorf("unexpected peer %v returned", pid)
        }
    }
}


func TestSelectWorstPeer_SamePeerMultipleBlocks(t *testing.T) {
    // ピアを複数生成
    p1 := mustPeer()
    p2 := mustPeer()
    p3 := mustPeer()
    outbound := []peer.ID{p1, p2, p3}

    store := NewObservationStore()
    manager := &PeerSelectorManager{
        store: store,
        Delta: 10 * time.Second,
    }

    // 複数ブロック生成
    root1 := fakeRoot(1)
    root2 := fakeRoot(2)
    root3 := fakeRoot(3)

    // --------------------------------------------------
    // p1 が複数のブロックを受信、他は1回ずつ
	// Delta = 10s
    store.RecordBlkReceipt(p1, root1, time.Unix(100, 0))
    store.RecordBlkReceipt(p1, root2, time.Unix(105, 0))
    store.RecordBlkReceipt(p1, root3, time.Unix(110, 0))
	// p1 score = 0s + 0s + 0s = 0s

    store.RecordBlkReceipt(p2, root1, time.Unix(102, 0))
	// p2 score = 2s + Δ + Δ = 2s + 10s + 10s = 22s
    store.RecordBlkReceipt(p3, root1, time.Unix(103, 0))
	// p3 score = 3s + Δ + Δ = 3s + 10s + 10s = 23s

    replacedNumber := 2
    worst, ok := manager.SelectWorstPeers(outbound, replacedNumber)
    if !ok {
        t.Fatalf("expected ok=true")
    }
    if len(worst) != replacedNumber {
        t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
    }

    // スコア計算は受信時間に依存するため、p1 は複数回受信しても最後の受信時間でスコアが決まることが多い
    // worst には遅延が大きい順に上位2ピアが入るはず
    expectedSet := map[peer.ID]bool{p2: true, p3: true}
    for _, pid := range worst {
        if !expectedSet[pid] {
            t.Errorf("unexpected peer %v returned", pid)
        }
    }

    // --------------------------------------------------
    // さらに別のブロックで p2 が遅延
    root4 := fakeRoot(4)
	// p1 score = delta = 10s
    store.RecordBlkReceipt(p2, root4, time.Unix(120, 0)) // p2 遅延
	// p2 score = 5s
    store.RecordBlkReceipt(p3, root4, time.Unix(115, 0)) // p3 遅延
	// p3 score = 0s
    // p1 はまだ早い
    replacedNumber = 2
    worst, ok = manager.SelectWorstPeers(outbound, replacedNumber)
    if !ok {
        t.Fatalf("expected ok=true")
    }
    if len(worst) != replacedNumber {
        t.Fatalf("expected %d worst peers, got %d: %v", replacedNumber, len(worst), worst)
    }

    // p2 が未配信、p1 が次
    expectedSet = map[peer.ID]bool{p1: true, p2: true}
    for _, pid := range worst {
        if !expectedSet[pid] {
            t.Errorf("unexpected peer %v returned", pid)
        }
    }
}
