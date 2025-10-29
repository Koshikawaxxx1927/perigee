package vcs

import (
	"fmt"
	"math"
	"net"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	RANDOM_NEIGHBOR_NUM = 16
	ERROR_LIMIT         = 0.25
	MIN_ERROR           = 0.1
	ADDPATIVE_TIMESTEP  = 0.25
	GRAVITY_RHO         = 500
	CENTROID_DRIFT      = 50
)

// VivaldiModel struct
type VivaldiModel struct {
	EnableIN1      bool
	EnableIN2      bool
	EnableIN3      bool
	HaveEnoughPeer bool
	D              int

	HistoryForceStat HistoryStat
	SelfID           enode.ID
	SelfIP           net.IP
	HistoryCounter   int

	PeerCoord     map[enode.ID]Coordinate
	ReceivedForce map[enode.ID]EuclideanVector
	ReceivedCoord map[enode.ID][]Coordinate // Using slice instead of deque
	ReceivedRTT   map[enode.ID]*HistoryStat
	Blacklist     map[enode.ID]struct{}

	LocalCoord    Coordinate
	RandomPeerSet map[enode.ID]struct{}
}

// NewVivaldiModel creates a new VivaldiModel instance
func NewVivaldiModel(D int, id enode.ID, ip net.IP, in1, in2, in3 bool) *VivaldiModel {
	return &VivaldiModel{
		D:                D,
		HistoryForceStat: *NewHistoryStat(100),
		SelfID:           id,
		SelfIP:           ip,
		HistoryCounter:   0,
		EnableIN1:        in1,
		EnableIN2:        in2,
		EnableIN3:        in3,
		PeerCoord:        make(map[enode.ID]Coordinate),
		ReceivedForce:    make(map[enode.ID]EuclideanVector),
		ReceivedRTT:      make(map[enode.ID]*HistoryStat),
		ReceivedCoord:    make(map[enode.ID][]Coordinate),
		Blacklist:        make(map[enode.ID]struct{}),
		RandomPeerSet:    make(map[enode.ID]struct{}),
		LocalCoord:       Coordinate{V: NewEuclideanVector(D), H: 0, E: 2.0},
		HaveEnoughPeer:   false,
	}
}

// Coordinate returns the local coordinate
func (vm *VivaldiModel) Coordinate() Coordinate {
	return vm.LocalCoord
}

// Set the new coordinate
func (vm *VivaldiModel) SetCoordinate(newCoord Coordinate) {
	vm.LocalCoord = newCoord
}

// Vector returns the local vector
func (vm *VivaldiModel) Vector() EuclideanVector {
	return vm.LocalCoord.V
}

// BelongToPeerSet checks if the given id belongs to the peer set
func (vm *VivaldiModel) BelongToPeerSet(id enode.ID) bool {
	_, exists := vm.PeerCoord[id]
	return exists
}

func (vm *VivaldiModel) Observe(remoteID enode.ID, remoteCoord Coordinate, rtt float64) {
	if _, blacklisted := vm.Blacklist[remoteID]; blacklisted {
		fmt.Printf("blocked: self id = %s, remote id = %d\n", vm.SelfID, remoteID)
		return
	}

	if !vm.HaveEnoughPeer {
		if !vm.BelongToPeerSet(remoteID) {
			vm.PeerCoord[remoteID] = remoteCoord
			vm.ReceivedForce[remoteID] = EuclideanVector{}
			vm.ReceivedRTT[remoteID] = NewHistoryStat(10)
			vm.ReceivedCoord[remoteID] = []Coordinate{}

			vm.RandomPeerSet[remoteID] = struct{}{}
			if len(vm.RandomPeerSet) == RANDOM_NEIGHBOR_NUM {
				vm.HaveEnoughPeer = true
			}
		}
	}

	if rtt == 0 {
		fmt.Println("RTT = 0")
		return
	}

	if vm.BelongToPeerSet(remoteID) {
		vm.ReceivedRTT[remoteID].Observe(rtt)
		med := vm.ReceivedRTT[remoteID].GetMedian()
		rtt = med

		if pastCoordQueue, exists := vm.ReceivedCoord[remoteID]; exists && len(pastCoordQueue) > 0 {
			lastCoord := pastCoordQueue[len(pastCoordQueue)-1]
			if lastCoord.E <= 0.3 && remoteCoord.E <= 0.3 {
				coordDiff := Sub(lastCoord.V, remoteCoord.V).Magnitude()
				if coordDiff > 75 {
					return
				}
			}
		}

		if len(vm.ReceivedCoord[remoteID]) == 10 {
			vm.ReceivedCoord[remoteID] = vm.ReceivedCoord[remoteID][1:]
		}
		vm.ReceivedCoord[remoteID] = append(vm.ReceivedCoord[remoteID], remoteCoord)
	}

	weight := vm.LocalCoord.E / (vm.LocalCoord.E + remoteCoord.E)

	predictRTT := EstimateRTT(vm.LocalCoord, remoteCoord)
	relativeError := math.Abs(predictRTT-rtt) / rtt

	weightedError := ERROR_LIMIT * weight
	newError := relativeError*weightedError + vm.LocalCoord.E*(1.0-weightedError)
	if newError < MIN_ERROR {
		newError = MIN_ERROR
	}

	adaptiveTimestep := ADDPATIVE_TIMESTEP * weight

	weightedForceMagnitude := adaptiveTimestep * (rtt - predictRTT)
	if weightedForceMagnitude > 100 {
		return
	}

	v := Sub(vm.LocalCoord.V, remoteCoord.V)
	var unitV EuclideanVector
	if v.IsZero() {
		unitV = RandomUnitVector(vm.D)
	} else {
		unitV = Div(v, predictRTT)
	}

	newHeight := vm.LocalCoord.H
	if v.IsZero() {
		newHeight += weightedForceMagnitude
	} else {
		newHeight += (vm.LocalCoord.H + remoteCoord.H) * weightedForceMagnitude / predictRTT
	}

	force := Mul(unitV, weightedForceMagnitude)

	if newHeight < 0 {
		newHeight = 0
	}

	historyMedian := vm.HistoryForceStat.GetMedian()
	medianDev := vm.HistoryForceStat.GetMedianDev()

	if vm.EnableIN3 && vm.HistoryCounter > 20 &&
		math.Abs(weightedForceMagnitude) > 20 &&
		math.Abs(weightedForceMagnitude) > historyMedian+8*medianDev {
		return
	}

	if vm.EnableIN3 {
		vm.HistoryForceStat.Observe(force.Magnitude())
	}

	if vm.EnableIN1 && vm.BelongToPeerSet(remoteID) {
		vm.ReceivedForce[remoteID] = Add(vm.ReceivedForce[remoteID], force)
	}

	newCoord := Add(vm.LocalCoord.V, force)
	newCoordMag := newCoord.Magnitude()
	unitDir := Div(newCoord, newCoordMag)
	gravityWeight := math.Pow(newCoordMag/GRAVITY_RHO, 2)
	newCoord = Sub(newCoord, Mul(unitDir, (gravityWeight)))

	vm.LocalCoord = Coordinate{V: newCoord, H: newHeight, E: newError}

	vm.HistoryCounter++
	if vm.EnableIN1 && vm.HistoryCounter > 20 {
		vm.SecurityIn1()
	}
}

func (vm *VivaldiModel) GetRTT() map[enode.ID]*HistoryStat {
	return vm.ReceivedRTT
}

func (vm *VivaldiModel) SecurityIn1() {
	var centroid EuclideanVector
	centroid = Add(centroid, vm.LocalCoord.V)

	for _, coord := range vm.PeerCoord {
		centroid = Add(centroid, coord.V)
	}

	peerCount := float64(len(vm.PeerCoord) + 1)
	centroid = Mul(centroid, 1.0/peerCount)

	if centroid.Magnitude() > CENTROID_DRIFT {
		var force EuclideanVector
		maxForceMag := 0.0
		// maliciousID := 0
		var maliciousID enode.ID

		for id, f := range vm.ReceivedForce {
			p := Dot(f, centroid)
			if p > maxForceMag {
				force = f
				maxForceMag = p
				maliciousID = id
			}
		}

		newCoord := Sub(vm.LocalCoord.V, force)
		newHeight := vm.LocalCoord.H
		newErr := vm.LocalCoord.E
		vm.LocalCoord = Coordinate{V: newCoord, H: newHeight, E: newErr}

		delete(vm.RandomPeerSet, maliciousID)
		delete(vm.PeerCoord, maliciousID)
		delete(vm.ReceivedForce, maliciousID)
		vm.HaveEnoughPeer = false

		vm.Blacklist[maliciousID] = struct{}{}
	}
}
