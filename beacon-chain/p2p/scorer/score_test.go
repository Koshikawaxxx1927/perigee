package scorer

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	field_params "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
)

func TestCalculateScoreForPeer_SingleBlock(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	root := fakeRoot(1)
	Delta := 10 * time.Second

	snapshot := map[[field_params.RootLength]byte]map[peer.ID]time.Time{
		root: {
			p1: time.Unix(100, 0),
			p2: time.Unix(90, 0), // fastest
		},
	}

	score := CalculateScoreForPeer(snapshot, p1, Delta)
	expected := 10 * time.Second
	if score != expected {
		t.Fatalf("expected %v, got %v", expected, score)
	}

	score2 := CalculateScoreForPeer(snapshot, p2, Delta)
	if score2 != 0 {
		t.Fatalf("expected 0, got %v", score2)
	}
}

func TestCalculateScoreForPeer_MultipleBlocks(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	root1 := fakeRoot(1)
	root2 := fakeRoot(2)
	Delta := 10 * time.Second

	snapshot := map[[field_params.RootLength]byte]map[peer.ID]time.Time{
		root1: {
			p1: time.Unix(100, 0),
			p2: time.Unix(90, 0),
			p3: time.Time{}, // 未配信
		},
		root2: {
			p1: time.Unix(200, 0),
			p2: time.Unix(195, 0),
			p3: time.Time{}, // 未配信
		},
	}

	score1 := CalculateScoreForPeer(snapshot, p1, Delta)
	expected1 := (10*time.Second + 5*time.Second) / 2
	if score1 != expected1 {
		t.Fatalf("expected %v, got %v", expected1, score1)
	}

	score2 := CalculateScoreForPeer(snapshot, p2, Delta)
	expected2 := 0
	if score2 != time.Duration(expected2) {
		t.Fatalf("expected 0, got %v", expected2)
	}

	score3 := CalculateScoreForPeer(snapshot, p3, Delta)
	if score3 != Delta {
		t.Fatalf("expected %v, got %v", Delta, score3)
	}
}

func TestCompleteScoresForAllPeers(t *testing.T) {
	p1 := mustPeer()
	p2 := mustPeer()
	p3 := mustPeer()
	root := fakeRoot(1)
	Delta := 10 * time.Second

	snapshot := map[[field_params.RootLength]byte]map[peer.ID]time.Time{
		root: {
			p1: time.Unix(100, 0),
			p2: time.Unix(90, 0),
			p3: time.Time{}, // 未配信
		},
	}

	scores := CompleteScoresForAllPeers(snapshot, Delta)

	if scores[p1] != 10*time.Second {
		t.Fatalf("p1 expected 10s, got %v", scores[p1])
	}
	if scores[p2] != 0 {
		t.Fatalf("p2 expected 0s, got %v", scores[p2])
	}
	if scores[p3] != Delta {
		t.Fatalf("p3 expected Delta, got %v", scores[p3])
	}
}
