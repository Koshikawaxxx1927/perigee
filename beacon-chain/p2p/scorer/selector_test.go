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

	worst, ok := store.SelectWorstPeer(Delta, outbound)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	// p3 は未配信なので Δ が加算され、最悪スコアになる
	if worst != p3 {
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

	worst, ok := store.SelectWorstPeer(Delta, outbound)
	if !ok {
		t.Fatalf("expected a worst peer, got none")
	}

	if worst != p1 {
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

	worst, _ := store.SelectWorstPeer(Delta, outbound)

	// 未配信ピアは Δ が加算されるため最悪
	if worst != p3 {
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

	worst, _ := store.SelectWorstPeer(Delta, outbound)

	if worst != p3 {
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

	worst, _ := store.SelectWorstPeer(Delta, outbound)
	// 両者同じタイミング → どちらでも OK
	if worst != p1 && worst != p2 {
		t.Fatalf("expected p1 or p2 as worst peer, got %v", worst)
	}
}

func TestSelectWorstPeer_EmptyStore(t *testing.T) {
	store := NewObservationStore()
	Delta := 10 * time.Second
	outbound := []peer.ID{}

	worst, ok := store.SelectWorstPeer(Delta, outbound)
	if ok {
		t.Fatalf("expected no worst peer, got %v", worst)
	}
}

func TestSelectWorstPeer_Tie(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	outbound := []peer.ID{p1, p2}

	store := NewObservationStore()
	// root := fakeRoot(1)
	Delta := 10 * time.Second

	// 両者未配信 → スコアは同じ Δ
	// どちらが選ばれるかは map の走査順に依存
	worst, _ := store.SelectWorstPeer(Delta, outbound)
	if worst == p1 || worst == p2 {
		t.Fatalf("expected p1 or p2 as worst peer, got %v", worst)
	}
}
