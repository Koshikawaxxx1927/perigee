package vcs

import (
	"math/rand"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// Node represents a host on the network.
type Addresses struct {
	IP      net.IP
	VCSPort uint
}

var (
	VivaldiModels  = make(map[enode.ID]*VivaldiModel)
	EnodeAddresses = make(map[enode.ID]*Addresses)
	MyEnode        enode.ID
)

// ランダムに要素を取得する関数
func RandomEnodeAddresses() (enode.ID, *Addresses) {
	// マップのキーをスライスに変換
	keys := make([]enode.ID, 0, len(EnodeAddresses))
	for key := range EnodeAddresses {
		keys = append(keys, key)
	}

	// マップが空でないことを確認
	if len(keys) == 0 {
		return enode.HexID(""), nil
	}

	// ランダムにキーを選択
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomKey := keys[r.Intn(len(keys))]

	// ランダムに選ばれたキーに対応する値を取得
	randomValue := EnodeAddresses[randomKey]

	return randomKey, randomValue
}
