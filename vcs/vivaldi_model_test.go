package vcs

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func DistanceSquared(v1, v2 EuclideanVector) float64 {
	// Simple Euclidean distance squared calculation for the test
	sum := 0.0
	for i := range v1.V {
		diff := v1.V[i] - v2.V[i]
		sum += diff * diff
	}
	return sum
}

var testEnodeAddresses map[enode.ID]*Addresses

var testMyEnode = enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")
var testRemoteEnode = enode.HexID("8f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")

func createEnodeAddresses() {
	// モックデータを作成
	testEnodeAddresses = map[enode.ID]*Addresses{
		enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.1"), VCSPort: 30303},
		enode.HexID("7f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.2"), VCSPort: 30304},
		enode.HexID("8f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.3"), VCSPort: 30305},
		enode.HexID("9f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.4"), VCSPort: 30306},
		enode.HexID("1f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.5"), VCSPort: 30307},
		enode.HexID("2f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.6"), VCSPort: 30308},
		enode.HexID("3f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.7"), VCSPort: 30309},
		enode.HexID("4f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.8"), VCSPort: 30310},
		enode.HexID("0f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.9"), VCSPort: 30311},
	}
}

func TestNewVivaldiModel(t *testing.T) {
	createEnodeAddresses()
	vm := NewVivaldiModel(3, testMyEnode, testEnodeAddresses[testMyEnode].IP, true, false, true)

	if vm.D != 3 {
		t.Errorf("expected D to be 3, got %d", vm.D)
	}
	if vm.SelfID != testMyEnode {
		t.Errorf("expected SelfID to be 1, got %d", vm.SelfID)
	}
	if vm.EnableIN1 != true {
		t.Errorf("expected EnableIN1 to be true, got %v", vm.EnableIN1)
	}
	if vm.HaveEnoughPeer != false {
		t.Errorf("expected HaveEnoughPeer to be false, got %v", vm.HaveEnoughPeer)
	}
}

func TestObserve(t *testing.T) {
	createEnodeAddresses()
	vm := NewVivaldiModel(3, testMyEnode, testEnodeAddresses[testMyEnode].IP, true, false, true)

	remoteCoord := Coordinate{
		V: NewEuclideanVectorFromArray([]float64{1.0, 2.0, 3.0}),
		E: 0.5,
		H: 1.0,
	}

	vm.Observe(testRemoteEnode, remoteCoord, 100.0)

	// Check if the coordinates are not a number(NaN)
	v0 := vm.Vector().V[0]
	v1 := vm.Vector().V[1]
	v2 := vm.Vector().V[2]
	if math.IsNaN(v0) || math.IsNaN(v1) || math.IsNaN(v2) {
		t.Errorf("expected model is number, but got NaN")
	}
	if len(vm.PeerCoord) != 1 {
		t.Errorf("expected 1 peer, got %d", len(vm.PeerCoord))
	}
	if _, exists := vm.ReceivedRTT[testRemoteEnode]; !exists {
		t.Errorf("expected ReceivedRTT for remoteID 2 to exist")
	}
	if len(vm.ReceivedCoord[testRemoteEnode]) != 1 {
		t.Errorf("expected ReceivedCoord[2] to have 1 coordinate, got %d", len(vm.ReceivedCoord[testRemoteEnode]))
	}

	// Test observing a blacklisted peer
	vm.Blacklist[testRemoteEnode] = struct{}{}
	vm.Observe(testRemoteEnode, remoteCoord, 100.0) // This should not add the peer
	if len(vm.PeerCoord) != 1 {
		t.Errorf("expected 1 peer after blacklisting, got %d", len(vm.PeerCoord))
	}
}

func TestGetRTT(t *testing.T) {
	createEnodeAddresses()
	vm := NewVivaldiModel(3, testMyEnode, testEnodeAddresses[testMyEnode].IP, true, false, true)

	remoteCoord := Coordinate{
		V: NewEuclideanVectorFromArray([]float64{1.0, 2.0, 3.0}),
		E: 0.5,
		H: 1.0,
	}
	vm.Observe(testRemoteEnode, remoteCoord, 100.0)

	if rttMap := vm.GetRTT(); len(rttMap) != 1 {
		t.Errorf("expected RTT map to have 1 entry, got %d", len(rttMap))
	}
}

func TestObserveMultiplePeers(t *testing.T) {
	createEnodeAddresses()
	vm := NewVivaldiModel(3, testMyEnode, testEnodeAddresses[testMyEnode].IP, true, false, true)

	for i := 0; i < 20; i++ {
		remoteCoord := Coordinate{
			V: NewEuclideanVectorFromArray([]float64{float64(i), float64(i), float64(i)}),
			E: 0.5,
			H: 1.0,
		}

		// Ensure that the hex string has an even length
		remoteID := fmt.Sprintf("%064x", i) // Ensure the string is 64 characters long
		vm.Observe(enode.HexID(remoteID), remoteCoord, 100.0)
	}

	if len(vm.PeerCoord) != RANDOM_NEIGHBOR_NUM {
		t.Errorf("expected %d peers, got %d", RANDOM_NEIGHBOR_NUM, len(vm.PeerCoord))
	}
}

func TestObserveZeroRTT(t *testing.T) {
	createEnodeAddresses()
	vm := NewVivaldiModel(3, testMyEnode, testEnodeAddresses[testMyEnode].IP, true, false, true)

	remoteCoord := Coordinate{
		V: NewEuclideanVectorFromArray([]float64{1.0, 2.0, 3.0}),
		E: 0.5,
		H: 1.0,
	}

	vm.Observe(testRemoteEnode, remoteCoord, 0.0) // Test with RTT = 0

	// Ensure the state is valid and the model doesn't fail
	if len(vm.PeerCoord) != 1 {
		t.Errorf("expected 1 peer after RTT 0, got %d", len(vm.PeerCoord))
	}
}
