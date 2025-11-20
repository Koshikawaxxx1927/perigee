package scorer

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

func TestSelectWorstPeer_Basic(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	outbound := []peer.ID{p1, p2, p3}

	store := NewObservationStore()

	root1 := fakeRoot(1)
	root2 := fakeRoot(2)
	Delta := 10 * time.Second

	// ブロック配信状況
	store.RecordBlkReceipt(p1, root1, time.Unix(100, 0))
	store.RecordBlkReceipt(p2, root1, time.Unix(90, 0))
	store.RecordBlkReceipt(p1, root2, time.Unix(200, 0))
	store.RecordBlkReceipt(p2, root2, time.Unix(195, 0))
	// p3 は未配信（zero value）

	replacedNumber := 1
	worst, ok := store.SelectWorstPeers(Delta, outbound, replacedNumber)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	// p3 は未配信なので Δ が加算され、最悪スコアになる
	if worst[0] != p3 {
		t.Fatalf("expected p3 as worst peer, got %v", worst)
	}
}

func TestSelectWorstPeer_DelayedPeer(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	outbound := []peer.ID{p1, p2}

	store := NewObservationStore()
	root := fakeRoot(1)
	Delta := 10 * time.Second

	// p1 は遅れて配信
	store.RecordBlkReceipt(p1, root, time.Unix(200, 0))
	store.RecordBlkReceipt(p2, root, time.Unix(100, 0))

	replacedNumber := 1
	worst, ok := store.SelectWorstPeers(Delta, outbound, replacedNumber)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	if worst[0] != p1 {
		t.Fatalf("expected p1 as worst peer due to delay, got %v", worst)
	}
}

func TestSelectWorstPeer_PartialDelivery(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	outbound := []peer.ID{p1, p2, p3}

	store := NewObservationStore()
	root := fakeRoot(1)
	Delta := 10 * time.Second

	// p1,p2 配信、p3 未配信
	store.RecordBlkReceipt(p1, root, time.Unix(105, 0))
	store.RecordBlkReceipt(p2, root, time.Unix(100, 0))

	replacedNumber := 1
	worst, _ := store.SelectWorstPeers(Delta, outbound, replacedNumber)

	// 未配信ピアは Δ が加算されるため最悪
	if worst[0] != p3 {
		t.Fatalf("expected p3 as worst peer due to no delivery, got %v", worst)
	}
}

func TestSelectWorstPeer_MultipleBlocksAverage(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	outbound := []peer.ID{p1, p2, p3}

	store := NewObservationStore()
	root1 := fakeRoot(1)
	root2 := fakeRoot(2)
	Delta := 10 * time.Second

	store.RecordBlkReceipt(p1, root1, time.Unix(100, 0))
	store.RecordBlkReceipt(p2, root1, time.Unix(90, 0))
	store.RecordBlkReceipt(p1, root2, time.Unix(200, 0))
	store.RecordBlkReceipt(p2, root2, time.Unix(195, 0))
	// p3 は未配信

	replacedNumber := 1
	worst, _ := store.SelectWorstPeers(Delta, outbound, replacedNumber)

	if worst[0] != p3 {
		t.Fatalf("expected p3 as worst peer due to all missing deliveries, got %v", worst)
	}
}

func TestSelectWorstPeer_AllFastDelivery(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	outbound := []peer.ID{p1, p2}

	store := NewObservationStore()
	root := fakeRoot(1)
	Delta := 10 * time.Second

	store.RecordBlkReceipt(p1, root, time.Unix(100, 0))
	store.RecordBlkReceipt(p2, root, time.Unix(100, 0))

	replacedNumber := 1
	worst, _ := store.SelectWorstPeers(Delta, outbound, replacedNumber)
	// 両者同じタイミング → どちらでも OK
	if worst[0] != p1 && worst[0] != p2 {
		t.Fatalf("expected p1 or p2 as worst peer, got %v", worst)
	}
}

func TestSelectWorstPeer_EmptyStore(t *testing.T) {
	store := NewObservationStore()
	Delta := 10 * time.Second
	outbound := []peer.ID{}
	replacedNumber := 1
	worst, ok := store.SelectWorstPeers(Delta, outbound, replacedNumber)
	if ok {
		t.Fatalf("expected no worst peer, got %v", worst)
	}
}

// 二つのピアで二つのピアからブロックが送信されなかった場合
// func TestSelectWorstPeer_Tie(t *testing.T) {
// 	p1 := mustPeer()
// 	p2 := mustPeer()
// 	outbound := []peer.ID{p1, p2}

// 	store := NewObservationStore()
// 	// root := fakeRoot(1)
// 	Delta := 10 * time.Second

// 	// 両者未配信 → スコアは同じ Δ
// 	// どちらが選ばれるかは map の走査順に依存
// 	replacedNumber := 1
// 	worst, _ := store.SelectWorstPeers(Delta, outbound, replacedNumber)
// 	if len(worst) != 2 {
// 		t.Fatalf("expected both peers as worst due to tie, got %v", worst)
// 	}
// 	if worst[0] == p1 || worst[0] == p2 {
// 		t.Fatalf("expected p1 or p2 as worst peer, got %v", worst)
// 	}
// }


func TestSelectWorstPeers_TopN_MultipleCases(t *testing.T) {
    // ピアを複数生成
    p1 := mustPeer()
    p2 := mustPeer()
    p3 := mustPeer()
    p4 := mustPeer()
    outbound := []peer.ID{p1, p2, p3, p4}

    store := NewObservationStore()
    root := fakeRoot(1)
    Delta := 10 * time.Second

    t.Run("AllFastDelivery_SameScore", func(t *testing.T) {
        // 全員同じ受信時間 → スコア同点
        store.RecordBlkReceipt(p1, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p2, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p3, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p4, root, time.Unix(100, 0))

        topN := 3
        worst, ok := store.SelectWorstPeers(Delta, outbound, topN)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != topN {
            t.Fatalf("expected %d worst peers, got %d: %v", topN, len(worst), worst)
        }

        // 全員同点なのでどの組み合わせでも OK
        for _, pid := range worst {
            found := false
            for _, p := range outbound {
                if pid == p {
                    found = true
                    break
                }
            }
            if !found {
                t.Fatalf("unexpected peer %v returned", pid)
            }
        }
    })

    t.Run("DifferentDeliveryTimes", func(t *testing.T) {
        // 受信時間に差をつける → スコアに差が出る
        store.RecordBlkReceipt(p1, root, time.Unix(100, 0)) // 最速
        store.RecordBlkReceipt(p2, root, time.Unix(105, 0))
        store.RecordBlkReceipt(p3, root, time.Unix(110, 0))
        store.RecordBlkReceipt(p4, root, time.Unix(120, 0)) // 最遅

        topN := 2
        worst, ok := store.SelectWorstPeers(Delta, outbound, topN)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != topN {
            t.Fatalf("expected %d worst peers, got %d: %v", topN, len(worst), worst)
        }

		expectedSet := map[peer.ID]bool{p3: true, p4: true}
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

        topN := 2
        worst, ok := store.SelectWorstPeers(Delta, outbound, topN)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != topN {
            t.Fatalf("expected %d worst peers, got %d: %v", topN, len(worst), worst)
        }

        // 最悪は p3, p4 のどちらでもOK（同点）
        valid := map[peer.ID]bool{p3: true, p4: true}
        for _, pid := range worst {
            if !valid[pid] {
                t.Fatalf("unexpected peer %v returned", pid)
            }
        }
    })

    t.Run("TopNExceedsPeerCount", func(t *testing.T) {
        // topN > ピア数 → 全員返る
        store.RecordBlkReceipt(p1, root, time.Unix(100, 0))
        store.RecordBlkReceipt(p2, root, time.Unix(105, 0))

        topN := 5 // ピア数より多い
        worst, ok := store.SelectWorstPeers(Delta, outbound[:2], topN)
        if !ok {
            t.Fatalf("expected ok=true")
        }
        if len(worst) != 2 {
            t.Fatalf("expected 2 peers returned, got %d: %v", len(worst), worst)
        }
    })
}
