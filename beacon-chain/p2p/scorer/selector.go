package scorer

import (
	"errors"
	"math"
	"math/rand"
	"sort"

	"github.com/libp2p/go-libp2p/core/peer"
)

type lcbucb struct {
	LCB float64
	UCB float64
}

type Selector struct{}

func NewSelector() *Selector {
	return &Selector{}
}

// SelectWithLCBReplacement selects up to dout peers.
// current: 現在の近隣
// bounds: 各ピアのLCB/UCB
// candidatePool: 置換候補
func (sel *Selector) SelectWithLCBReplacement(
	bounds map[peer.ID]lcbucb,
	current []peer.ID,
	dout int,
	candidatePool []peer.ID,
) ([]peer.ID, error) {
	if dout <= 0 {
		return nil, errors.New("dout must be > 0")
	}

	selected := make([]peer.ID, 0, dout)

	// --- 1. Top-K選択 (current内でLCBが小さい順) ---
	type kv struct {
		id  peer.ID
		lcb float64
	}
	currBounds := make([]kv, 0, len(current))
	for _, pid := range current {
		if b, ok := bounds[pid]; ok {
			currBounds = append(currBounds, kv{id: pid, lcb: b.LCB})
		}
	}
	sort.Slice(currBounds, func(i, j int) bool { return currBounds[i].lcb < currBounds[j].lcb })

	for i := 0; i < len(currBounds) && len(selected) < dout; i++ {
		selected = append(selected, currBounds[i].id)
	}

	// --- 2. LCB/UCB置換 ---
	if len(selected) > 0 && len(candidatePool) > 0 {
		// selected内の worstPeer を探す
		var worstPeer peer.ID
		maxLCB := -math.Inf(1)
		minUCB := math.Inf(1)
		for _, pid := range selected {
			b := bounds[pid]
			if b.LCB > maxLCB {
				maxLCB = b.LCB
				worstPeer = pid
			}
			if b.UCB < minUCB {
				minUCB = b.UCB
			}
		}

		// maxLCB > minUCB の場合のみ置換
		if maxLCB > minUCB {
			randIndex := rand.Intn(len(candidatePool))
			replacement := candidatePool[randIndex]
			for i, pid := range selected {
				if pid == worstPeer {
					selected[i] = replacement
					break
				}
			}
		}
	}

	// --- 3. dout に満たない場合は候補から追加 ---
	for _, pid := range candidatePool {
		if len(selected) >= dout {
			break
		}
		already := false
		for _, s := range selected {
			if s == pid {
				already = true
				break
			}
		}
		if !already {
			selected = append(selected, pid)
		}
	}

	return selected, nil
}
