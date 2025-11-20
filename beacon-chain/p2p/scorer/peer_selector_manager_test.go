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

	worst, ok := manager.SelectWorstPeer(peers)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	// p3 は未配信なので Δ が加算され、最悪スコアになる
	if worst != p3 {
		t.Fatalf("expected p3 as worst peer, got %v", worst)
	}
}

func TestPeerSelectorManager_SelectWorstPeer_Empty(t *testing.T) {
	manager := NewPeerSelectorManager(10 * time.Second)
	worst, ok := manager.SelectWorstPeer([]peer.ID{})
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

	worst, ok := manager.SelectWorstPeer(peers)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	// 両者同じタイミング → どちらでも正しい
	if worst != p1 && worst != p2 {
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

	worst1, _ := manager.SelectWorstPeer(peers)
	if worst1 != p3 {
		t.Fatalf("expected p3 as worst in round 1, got %v", worst1)
	}

	// ラウンド2
	root2 := fakeRoot(2)
	manager.RecordBlockDelivery(p1, root2, time.Unix(200, 0))
	manager.RecordBlockDelivery(p3, root2, time.Unix(210, 0))

	worst2, _ := manager.SelectWorstPeer(peers)
	if worst2 != p2 {
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

	worst, _ := manager.SelectWorstPeer(peers)
	if worst != p3 {
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
// 	worst, ok := manager.SelectWorstPeer()
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
