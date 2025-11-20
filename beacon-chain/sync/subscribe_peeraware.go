// remove unused import warning by actually using pubsub package
package sync

// import (
//     "context"
//     // "time"
//     // pubsub "github.com/libp2p/go-libp2p-pubsub"
//     "github.com/libp2p/go-libp2p/core/peer"
//     "google.golang.org/protobuf/proto"

//     ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
//     // "github.com/prysmaticlabs/prysm/v4/beacon-chain/p2p/scorer"
//     "github.com/sirupsen/logrus"
// )

// // subscribePeerAwareDirect subscribes directly using libp2p pubsub.
// // Replace s.pubsub below with the actual pubsub instance in your Service.
// // For example: s.cfg.p2p.PubSub or s.p2p.PubSub depending on your Service definition.
// func (s *Service) subscribePeerAwareDirect(topic string) error {
//     // Replace `s.pubsub` with your actual PubSub object:
//     sub, err := s.pubsub.Subscribe(topic) // <-- ADAPT this line to your Service field
//     if err != nil {
//         return err
//     }

//     go func() {
//         for {
//             m, err := sub.Next(s.ctx)
//             if err != nil {
//                 return
//             }

//             from := m.ReceivedFrom

//             // Unmarshal into concrete proto type used by Prysm — adapt if different
//             var pb ethpb.SignedBeaconBlock
//             if err := proto.Unmarshal(m.Data, &pb); err != nil {
//                 logrus.WithError(err).Debug("could not unmarshal incoming beacon block proto")
//                 continue
//             }

//             go func(pm proto.Message, pid peer.ID) {
//                 if err := s.beaconBlockSubscriberWithPeer(context.Background(), pm, pid); err != nil {
//                     logrus.WithError(err).Debug("beaconBlockSubscriberWithPeer returned error")
//                 }
//             }(&pb, from)
//         }
//     }()

//     s.subscriptions = append(s.subscriptions, sub) // <- ensure Service has subscriptions field
//     return nil
// }
