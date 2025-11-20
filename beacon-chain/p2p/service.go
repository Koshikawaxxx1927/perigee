// Package p2p defines the network protocol implementation for Ethereum consensus
// used by beacon nodes, including peer discovery using discv5, gossip-sub
// using libp2p, and handing peer lifecycles + handshakes.
package p2p

import (
	"context"
	"crypto/ecdsa"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/async"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/p2p/encoder"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/p2p/peers"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/p2p/peers/scorers"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/p2p/scorer"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/p2p/types"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	leakybucket "github.com/prysmaticlabs/prysm/v4/container/leaky-bucket"
	prysmnetwork "github.com/prysmaticlabs/prysm/v4/network"
	"github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1/metadata"
	"github.com/prysmaticlabs/prysm/v4/runtime"
	"github.com/prysmaticlabs/prysm/v4/time/slots"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

var _ runtime.Service = (*Service)(nil)

// In the event that we are at our peer limit, we
// stop looking for new peers and instead poll
// for the current peer limit status for the time period
// defined below.
var pollingPeriod = 6 * time.Second

// Refresh rate of ENR set at twice per slot.
var refreshRate = slots.DivideSlotBy(2)

// maxBadResponses is the maximum number of bad responses from a peer before we stop talking to it.
const maxBadResponses = 5

// maxDialTimeout is the timeout for a single peer dial.
var maxDialTimeout = params.BeaconConfig().RespTimeoutDuration()

// For Perigee project ////////////////////////////
// スコアの計算をするエンティティ
var manager *scorer.PeerSelectorManager

////////////////////////////////////////////////////

// Service for managing peer to peer (p2p) networking.
type Service struct {
	started               bool
	isPreGenesis          bool
	pingMethod            func(ctx context.Context, id peer.ID) error
	cancel                context.CancelFunc
	cfg                   *Config
	peers                 *peers.Status
	addrFilter            *multiaddr.Filters
	ipLimiter             *leakybucket.Collector
	privKey               *ecdsa.PrivateKey
	metaData              metadata.Metadata
	pubsub                *pubsub.PubSub
	joinedTopics          map[string]*pubsub.Topic
	joinedTopicsLock      sync.RWMutex
	subnetsLock           map[uint64]*sync.RWMutex
	subnetsLockLock       sync.Mutex // Lock access to subnetsLock
	initializationLock    sync.Mutex
	dv5Listener           Listener
	startupErr            error
	ctx                   context.Context
	host                  host.Host
	genesisTime           time.Time
	genesisValidatorsRoot []byte
	activeValidatorCount  uint64

	// For Perigee project ////////////////////////////
	mu sync.Mutex
	// ピアとして接続を拒否するピアのマップ
	blockList map[peer.ID]time.Time
	///////////////////////////////////////////////////
}

// NewService initializes a new p2p service compatible with shared.Service interface. No
// connections are made until the Start function is called during the service registry startup.
func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	var err error
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop().

	s := &Service{
		ctx:          ctx,
		cancel:       cancel,
		cfg:          cfg,
		isPreGenesis: true,
		joinedTopics: make(map[string]*pubsub.Topic, len(gossipTopicMappings)),
		subnetsLock:  make(map[uint64]*sync.RWMutex),
	}

	s.cfg = validateConfig(s.cfg)

	dv5Nodes := parseBootStrapAddrs(s.cfg.BootstrapNodeAddr)

	cfg.Discv5BootStrapAddr = dv5Nodes

	ipAddr := prysmnetwork.IPAddr()
	s.privKey, err = privKey(s.cfg)
	if err != nil {
		log.WithError(err).Error("Failed to generate p2p private key")
		return nil, err
	}
	s.metaData, err = metaDataFromConfig(s.cfg)
	if err != nil {
		log.WithError(err).Error("Failed to create peer metadata")
		return nil, err
	}
	s.addrFilter, err = configureFilter(s.cfg)
	if err != nil {
		log.WithError(err).Error("Failed to create address filter")
		return nil, err
	}
	s.ipLimiter = leakybucket.NewCollector(ipLimit, ipBurst, 30*time.Second, true /* deleteEmptyBuckets */)

	opts := s.buildOptions(ipAddr, s.privKey)
	h, err := libp2p.New(opts...)
	if err != nil {
		log.WithError(err).Error("Failed to create p2p host")
		return nil, err
	}

	s.host = h
	// Gossipsub registration is done before we add in any new peers
	// due to libp2p's gossipsub implementation not taking into
	// account previously added peers when creating the gossipsub
	// object.
	psOpts := s.pubsubOptions()
	// Set the pubsub global parameters that we require.
	setPubSubParameters()
	// Reinitialize them in the event we are running a custom config.
	attestationSubnetCount = params.BeaconConfig().AttestationSubnetCount
	syncCommsSubnetCount = params.BeaconConfig().SyncCommitteeSubnetCount

	gs, err := pubsub.NewGossipSub(s.ctx, s.host, psOpts...)
	if err != nil {
		log.WithError(err).Error("Failed to start pubsub")
		return nil, err
	}
	s.pubsub = gs

	s.peers = peers.NewStatus(ctx, &peers.StatusConfig{
		PeerLimit: int(s.cfg.MaxPeers),
		ScorerParams: &scorers.Config{
			BadResponsesScorerConfig: &scorers.BadResponsesScorerConfig{
				Threshold:     maxBadResponses,
				DecayInterval: time.Hour,
			},
		},
	})

	// Initialize Data maps.
	types.InitializeDataMaps()

	// For Perigee project ////////////////////////////
	// Deltaを指定してスコアの上限の指定して、Perigeeをセットアップ
	delta := 360 * time.Millisecond
	manager = scorer.NewPeerSelectorManager(delta)
	s.blockList = make(map[peer.ID]time.Time)
	////////////////////////////////////////////////////

	return s, nil
}

// Start the p2p service.
func (s *Service) Start() {
	if s.started {
		log.Error("Attempted to start p2p service when it was already started")
		return
	}

	// Waits until the state is initialized via an event feed.
	// Used for fork-related data when connecting peers.
	s.awaitStateInitialized()
	s.isPreGenesis = false

	var relayNodes []string
	if s.cfg.RelayNodeAddr != "" {
		relayNodes = append(relayNodes, s.cfg.RelayNodeAddr)
		if err := dialRelayNode(s.ctx, s.host, s.cfg.RelayNodeAddr); err != nil {
			log.WithError(err).Errorf("Could not dial relay node")
		}
	}

	if !s.cfg.NoDiscovery {
		ipAddr := prysmnetwork.IPAddr()
		listener, err := s.startDiscoveryV5(
			ipAddr,
			s.privKey,
		)
		if err != nil {
			log.WithError(err).Fatal("Failed to start discovery")
			s.startupErr = err
			return
		}
		err = s.connectToBootnodes()
		if err != nil {
			log.WithError(err).Error("Could not add bootnode to the exclusion list")
			s.startupErr = err
			return
		}
		s.dv5Listener = listener
		go s.listenForNewNodes()
	}

	s.started = true

	if len(s.cfg.StaticPeers) > 0 {
		addrs, err := PeersFromStringAddrs(s.cfg.StaticPeers)
		if err != nil {
			log.WithError(err).Error("Could not connect to static peer")
		}
		// Set trusted peers for those that are provided as static addresses.
		pids := peerIdsFromMultiAddrs(addrs)
		s.peers.SetTrustedPeers(pids)
		s.connectWithAllTrustedPeers(addrs)
	}
	// Initialize metadata according to the
	// current epoch.
	s.RefreshENR()

	// Periodic functions.
	async.RunEvery(s.ctx, params.BeaconConfig().TtfbTimeoutDuration(), func() {
		ensurePeerConnections(s.ctx, s.host, s.peers, relayNodes...)
	})
	async.RunEvery(s.ctx, 30*time.Minute, s.Peers().Prune)
	async.RunEvery(s.ctx, time.Duration(params.BeaconConfig().RespTimeout)*time.Second, s.updateMetrics)
	async.RunEvery(s.ctx, refreshRate, s.RefreshENR)
	async.RunEvery(s.ctx, 1*time.Minute, func() {
		log.WithFields(logrus.Fields{
			"inbound":     len(s.peers.InboundConnected()),
			"outbound":    len(s.peers.OutboundConnected()),
			"activePeers": len(s.peers.Active()),
			"numOfConns":  len(s.host.Network().Peers()),
		}).Info("Peer summary")
		for _, pid := range s.Peers().Connected() {
			addr, err := s.Peers().Address(pid)
			if err != nil {
				log.WithError(err).Warnf("could not get address for peer %s", pid)
				continue
			}
			// マルチアドレスから IP とポートを分離する例
			parts := strings.Split(addr.String(), "/")
			// 例: parts = ["", "ip4", "192.168.1.2", "tcp", "13000", "p2p", "<peerID>"]
			for i := 0; i < len(parts)-1; i++ {
				if parts[i] == "ip4" || parts[i] == "ip6" {
					ip := parts[i+1]
					log.Infof("Peer %s is at IP %s", pid, ip)
				}
			}
		}
	})

	multiAddrs := s.host.Network().ListenAddresses()
	logIPAddr(s.host.ID(), multiAddrs...)

	p2pHostAddress := s.cfg.HostAddress
	p2pTCPPort := s.cfg.TCPPort

	if p2pHostAddress != "" {
		logExternalIPAddr(s.host.ID(), p2pHostAddress, p2pTCPPort)
		verifyConnectivity(p2pHostAddress, p2pTCPPort, "tcp")
	}

	p2pHostDNS := s.cfg.HostDNS
	if p2pHostDNS != "" {
		logExternalDNSAddr(s.host.ID(), p2pHostDNS, p2pTCPPort)
	}
	go s.forkWatcher()

	// For Perigee project ////////////////////////////
	// 決まった時間ごとにピアの切断を呼び出すための関数
	s.StartPerigeePeerSelection()
	////////////////////////////////////////////////////
}

// Stop the p2p service and terminate all peer connections.
func (s *Service) Stop() error {
	defer s.cancel()
	s.started = false
	if s.dv5Listener != nil {
		s.dv5Listener.Close()
	}
	return nil
}

// Status of the p2p service. Will return an error if the service is considered unhealthy to
// indicate that this node should not serve traffic until the issue has been resolved.
func (s *Service) Status() error {
	if s.isPreGenesis {
		return nil
	}
	if !s.started {
		return errors.New("not running")
	}
	if s.startupErr != nil {
		return s.startupErr
	}
	if s.genesisTime.IsZero() {
		return errors.New("no genesis time set")
	}
	return nil
}

// Started returns true if the p2p service has successfully started.
func (s *Service) Started() bool {
	return s.started
}

// Encoding returns the configured networking encoding.
func (_ *Service) Encoding() encoder.NetworkEncoding {
	return &encoder.SszNetworkEncoder{}
}

// PubSub returns the p2p pubsub framework.
func (s *Service) PubSub() *pubsub.PubSub {
	return s.pubsub
}

// Host returns the currently running libp2p
// host of the service.
func (s *Service) Host() host.Host {
	return s.host
}

// SetStreamHandler sets the protocol handler on the p2p host multiplexer.
// This method is a pass through to libp2pcore.Host.SetStreamHandler.
func (s *Service) SetStreamHandler(topic string, handler network.StreamHandler) {
	s.host.SetStreamHandler(protocol.ID(topic), handler)
}

// PeerID returns the Peer ID of the local peer.
func (s *Service) PeerID() peer.ID {
	return s.host.ID()
}

// Disconnect from a peer.
func (s *Service) Disconnect(pid peer.ID) error {
	return s.host.Network().ClosePeer(pid)
}

// Connect to a specific peer.
func (s *Service) Connect(pi peer.AddrInfo) error {
	return s.host.Connect(s.ctx, pi)
}

// Peers returns the peer status interface.
func (s *Service) Peers() *peers.Status {
	return s.peers
}

// ENR returns the local node's current ENR.
func (s *Service) ENR() *enr.Record {
	if s.dv5Listener == nil {
		return nil
	}
	return s.dv5Listener.Self().Record()
}

// DiscoveryAddresses represents our enr addresses as multiaddresses.
func (s *Service) DiscoveryAddresses() ([]multiaddr.Multiaddr, error) {
	if s.dv5Listener == nil {
		return nil, nil
	}
	return convertToUdpMultiAddr(s.dv5Listener.Self())
}

// Metadata returns a copy of the peer's metadata.
func (s *Service) Metadata() metadata.Metadata {
	return s.metaData.Copy()
}

// MetadataSeq returns the metadata sequence number.
func (s *Service) MetadataSeq() uint64 {
	return s.metaData.SequenceNumber()
}

// AddPingMethod adds the metadata ping rpc method to the p2p service, so that it can
// be used to refresh ENR.
func (s *Service) AddPingMethod(reqFunc func(ctx context.Context, id peer.ID) error) {
	s.pingMethod = reqFunc
}

func (s *Service) pingPeers() {
	if s.pingMethod == nil {
		return
	}
	for _, pid := range s.peers.Connected() {
		go func(id peer.ID) {
			if err := s.pingMethod(s.ctx, id); err != nil {
				log.WithField("peer", id).WithError(err).Debug("Failed to ping peer")
			}
		}(pid)
	}
}

// Waits for the beacon state to be initialized, important
// for initializing the p2p service as p2p needs to be aware
// of genesis information for peering.
func (s *Service) awaitStateInitialized() {
	s.initializationLock.Lock()
	defer s.initializationLock.Unlock()
	if s.isInitialized() {
		return
	}
	clock, err := s.cfg.ClockWaiter.WaitForClock(s.ctx)
	if err != nil {
		log.WithError(err).Fatal("failed to receive initial genesis data")
	}
	s.genesisTime = clock.GenesisTime()
	gvr := clock.GenesisValidatorsRoot()
	s.genesisValidatorsRoot = gvr[:]
	_, err = s.currentForkDigest() // initialize fork digest cache
	if err != nil {
		log.WithError(err).Error("Could not initialize fork digest")
	}
}

func (s *Service) connectWithAllTrustedPeers(multiAddrs []multiaddr.Multiaddr) {
	addrInfos, err := peer.AddrInfosFromP2pAddrs(multiAddrs...)
	if err != nil {
		log.WithError(err).Error("Could not convert to peer address info's from multiaddresses")
		return
	}
	for _, info := range addrInfos {
		// add peer into peer status
		s.peers.Add(nil, info.ID, info.Addrs[0], network.DirUnknown)
		// make each dial non-blocking
		go func(info peer.AddrInfo) {
			if err := s.connectWithPeer(s.ctx, info); err != nil {
				log.WithError(err).Tracef("Could not connect with peer %s", info.String())
			}
		}(info)
	}
}

func (s *Service) connectWithAllPeers(multiAddrs []multiaddr.Multiaddr) {
	addrInfos, err := peer.AddrInfosFromP2pAddrs(multiAddrs...)
	if err != nil {
		log.WithError(err).Error("Could not convert to peer address info's from multiaddresses")
		return
	}
	for _, info := range addrInfos {
		// make each dial non-blocking
		go func(info peer.AddrInfo) {
			if err := s.connectWithPeer(s.ctx, info); err != nil {
				log.WithError(err).Tracef("Could not connect with peer %s", info.String())
			}
		}(info)
	}
}

func (s *Service) connectWithPeer(ctx context.Context, info peer.AddrInfo) error {
	ctx, span := trace.StartSpan(ctx, "p2p.connectWithPeer")
	defer span.End()

	if info.ID == s.host.ID() {
		return nil
	}
	if s.Peers().IsBad(info.ID) {
		return errors.New("refused to connect to bad peer")
	}
	ctx, cancel := context.WithTimeout(ctx, maxDialTimeout)
	defer cancel()
	if err := s.host.Connect(ctx, info); err != nil {
		s.Peers().Scorers().BadResponsesScorer().Increment(info.ID)
		return err
	}
	return nil
}

func (s *Service) connectToBootnodes() error {
	nodes := make([]*enode.Node, 0, len(s.cfg.Discv5BootStrapAddr))
	for _, addr := range s.cfg.Discv5BootStrapAddr {
		bootNode, err := enode.Parse(enode.ValidSchemes, addr)
		if err != nil {
			return err
		}
		// do not dial bootnodes with their tcp ports not set
		if err := bootNode.Record().Load(enr.WithEntry("tcp", new(enr.TCP))); err != nil {
			if !enr.IsNotFound(err) {
				log.WithError(err).Error("Could not retrieve tcp port")
			}
			continue
		}
		nodes = append(nodes, bootNode)
	}
	multiAddresses := convertToMultiAddr(nodes)
	s.connectWithAllPeers(multiAddresses)
	return nil
}

// Returns true if the service is aware of the genesis time and genesis validators root. This is
// required for discovery and pubsub validation.
func (s *Service) isInitialized() bool {
	return !s.genesisTime.IsZero() && len(s.genesisValidatorsRoot) == 32
}

// For Virtual Coordinate System code /////////////
func (s *Service) DisconnectAllOutboundPeers() {
	all := s.Peers().OutboundConnected() // 接続中のピア一覧
	for _, pid := range all {
		if err := s.Disconnect(pid); err != nil {
			log.WithError(err).Warnf("Failed to disconnect peer %s", pid)
		} else {
			log.Infof("Disconnected outbound peer %s", pid)
		}
	}
}

////////////////////////////////////////////////////

// For Perigee project ////////////////////////////
func (s *Service) StartPerigeePeerSelection() {
	log.Info("Peer selection with Perigee started")
	perigeeDisconnectionInterval := 30 * time.Minute
	ticker := time.NewTicker(perigeeDisconnectionInterval)
	go func() {
		for range ticker.C {
			s.perigeePeerSelection()
		}
	}()
}

func (s *Service) perigeePeerSelection() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// blockListを更新（360秒経過したものは削除）
	now := time.Now()
	for pid, t := range s.blockList {
		if now.Sub(t) >= 360*time.Second {
			delete(s.blockList, pid)
		}
	}

	// 接続中のアウトバウンドピアを取得
	outboundConnected := s.peers.OutboundConnected()

	// blockListにいるピアは除外
	filteredPeers := make([]peer.ID, 0, len(outboundConnected))
	for _, pid := range outboundConnected {
		if _, blocked := s.blockList[pid]; !blocked {
			filteredPeers = append(filteredPeers, pid)
		}
	}

	if len(filteredPeers) == 0 {
		return
	}

	// 最悪スコアのピアを選択
	worstPeer, found := manager.SelectWorstPeer(filteredPeers)
	if !found {
		return
	}

	// ピアを切断
	if err := s.Disconnect(worstPeer); err != nil {
		log.Infof("failed to disconnect peer %s: %v", worstPeer, err)
	} else {
		log.Infof("disconnected worst peer %s", worstPeer)
		// blockListに追加
		s.blockList[worstPeer] = now
	}
}

func PeerSelectorManager() *scorer.PeerSelectorManager { return manager }

// ////////////////////////////////////////////////////
