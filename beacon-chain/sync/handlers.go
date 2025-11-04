// For Perigee project //////////////////////////
package sync

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"

	"github.com/prysmaticlabs/prysm/v4/consensus-types/blocks"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/blockchain"
    "github.com/prysmaticlabs/prysm/v4/beacon-chain/core/transition/interop"

	// "github.com/prysmaticlabs/prysm/v4/beacon-chain/p2p"
)

func (s *Service) beaconBlockSubscriberWithPeer(ctx context.Context, msg proto.Message, from peer.ID) error {
	signed, err := blocks.NewSignedBeaconBlock(msg)
	if err != nil {
		return err
	}
	if err := blocks.BeaconBlockIsNil(signed); err != nil {
		return err
	}

	// ここで送信元ピア情報を記録できる（非同期推奨）
	go func(root [32]byte, pid peer.ID) {
		// ✅ ここでブロック観測を記録
		if mgr := s.cfg.p2p.PeerSelectorManager(); mgr != nil {
    		go mgr.AddBlockNow(root, pid)
		}
	}(func() [32]byte {
		r, _ := signed.Block().HashTreeRoot()
		return r
	}(), from)

	// 以下は既存処理
	s.setSeenBlockIndexSlot(signed.Block().Slot(), signed.Block().ProposerIndex())

	block := signed.Block()
	root, err := block.HashTreeRoot()
	if err != nil {
		return err
	}

	if err := s.cfg.chain.ReceiveBlock(ctx, signed, root, nil); err != nil {
		// エラーハンドリング（既存のまま）
		if blockchain.IsInvalidBlock(err) {
			r := blockchain.InvalidBlockRoot(err)
			if r != [32]byte{} {
				s.setBadBlock(ctx, r)
			} else {
				interop.WriteBlockToDisk(signed, true /*failed*/)
				s.setBadBlock(ctx, root)
			}
		}
		for _, root := range blockchain.InvalidAncestorRoots(err) {
			s.setBadBlock(ctx, root)
		}
		return err
	}
	return nil
}
// ///////////////////////////////////////////