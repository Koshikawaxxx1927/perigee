package scorer

import (
	"math"
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
) (peer.ID, error) {
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

	selected := make([]peer.ID, 0)
	for i := 0; i < len(currBounds); i++ {
		selected = append(selected, currBounds[i].id)
	}

	// --- 2. LCB/UCB置換 ---
	var replaced peer.ID
	// if len(selected) > 0 && len(candidatePool) > 0 {
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

		if maxLCB > minUCB {
			replaced = worstPeer
		}
	// }

	// --- 3. dout に満たない場合の補充は無視 ---
	// 置換されたノードだけを返すため、追加処理は行わない

	return replaced, nil
}
