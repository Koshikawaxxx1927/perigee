package scorer

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	field_params "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
)

// helper: generate peer.ID from string (stable)
func mustPeer() peer.ID {
	priv, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		panic(err)
	}
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		panic(err)
	}
	return pid
}

// helper: fixed block root
func fakeRoot(b byte) [field_params.RootLength]byte {
	var r [field_params.RootLength]byte
	for i := 0; i < field_params.RootLength; i++ {
		r[i] = b
	}
	return r
}

func TestObservationStore_RecordAndSnapshot(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	outbound := []peer.ID{p1, p2, p3}

	store := NewObservationStore()

	root1 := fakeRoot(1)
	root2 := fakeRoot(2)

	// p1,p2 だけ配信
	ts1 := time.Unix(100, 0)
	ts2 := time.Unix(110, 0)
	store.RecordBlkReceipt(p1, root1, ts1)
	store.RecordBlkReceipt(p2, root1, ts2)
	store.RecordBlkReceipt(p1, root2, ts2)

	// SnapshotAndClear
	snap := store.SnapshotAndClear(outbound)

	// 各ブロックのピア数
	if len(snap[root1]) != len(outbound) {
		t.Fatalf("expected %d peers in root1, got %d", len(outbound), len(snap[root1]))
	}
	if len(snap[root2]) != len(outbound) {
		t.Fatalf("expected %d peers in root2, got %d", len(outbound), len(snap[root2]))
	}

	// 配信したピアは timestamp が入っている
	if snap[root1][p1] != ts1 {
		t.Fatalf("root1 p1 timestamp mismatch")
	}
	if snap[root1][p2] != ts2 {
		t.Fatalf("root1 p2 timestamp mismatch")
	}
	if snap[root2][p1] != ts2 {
		t.Fatalf("root2 p1 timestamp mismatch")
	}

	// 配信していないピアは zero value
	if !snap[root1][p3].IsZero() {
		t.Fatalf("root1 p3 should be zero")
	}
	if !snap[root2][p2].IsZero() {
		t.Fatalf("root2 p2 should be zero")
	}
	if !snap[root2][p3].IsZero() {
		t.Fatalf("root2 p3 should be zero")
	}

	// store はクリアされている
	snap2 := store.SnapshotAndClear(outbound)
	if len(snap2) != 0 {
		t.Fatalf("store should be empty after SnapshotAndClear")
	}
}

func TestObservationStore_EarliestWins(t *testing.T) {
	p1 := mustPeer()
	outbound := []peer.ID{p1}
	store := NewObservationStore()

	root := fakeRoot(3)
	tsEarly := time.Unix(100, 0)
	tsLate := time.Unix(200, 0)

	// late first
	store.RecordBlkReceipt(p1, root, tsLate)
	// early second
	store.RecordBlkReceipt(p1, root, tsEarly)

	snap := store.SnapshotAndClear(outbound)
	got := snap[root][p1]
	if got != tsEarly {
		t.Fatalf("earlier timestamp should win: expected %v, got %v", tsEarly, got)
	}
}

func TestObservationStore_AllPeersRegistered(t *testing.T) {
	// outbound に登録されていないピアはスコア計算に含まれないことを確認
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer() // 未登録
	outbound := []peer.ID{p1, p2}

	store := NewObservationStore()

	root := fakeRoot(4)
	store.RecordBlkReceipt(p1, root, time.Unix(100, 0))
	store.RecordBlkReceipt(p2, root, time.Unix(110, 0))
	store.RecordBlkReceipt(p3, root, time.Unix(120, 0)) // 登録されていないピア

	snap := store.SnapshotAndClear(outbound)

	if _, ok := snap[root][p1]; !ok {
		t.Fatalf("p1 should exist")
	}
	if _, ok := snap[root][p2]; !ok {
		t.Fatalf("p2 should exist")
	}
	if _, ok := snap[root][p3]; ok {
		t.Fatalf("p3 should NOT exist (not in outboundPeers)")
	}
}
