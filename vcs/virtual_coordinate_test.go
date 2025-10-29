package vcs

import (
	"math"
	"net"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// TestGenerateVirtualCoordinate tests the GenerateVirtualCoordinate function.
func TestGenerateVirtualCoordinate(t *testing.T) {
	// モックデータを作成
	EnodeAddresses = map[enode.ID]*Addresses{
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
	MyEnode = enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")
	myIP := EnodeAddresses[MyEnode].IP
	VivaldiModels[MyEnode] = NewVivaldiModel(D, MyEnode, myIP, true, false, true)

	n := len(EnodeAddresses)
	D := 3 // Number of dimensions

	// Generate virtual coordinates for Vivaldi models
	err := GenerateVirtualCoordinate(D)
	if err != nil {
		t.Fatalf("error generating virtual coordinates: %v", err)
	}

	// Check if the number of generated Vivaldi models matches n
	if len(VivaldiModels) != n {
		t.Errorf("expected %d Vivaldi models, got %d", n, len(VivaldiModels))
	}

	// Check if the coordinates are not nil and have the correct dimension
	for i, model := range VivaldiModels {
		if len(model.Vector().V) != D {
			t.Errorf("expected model %d to have %d dimensions, but got %d", i, D, len(model.Vector().V))
		}
	}

	// Check if the coordinates are not a number(NaN)
	for _, model := range VivaldiModels {
		v0 := model.Vector().V[0]
		v1 := model.Vector().V[1]
		h := model.Coordinate().H
		if math.IsNaN(v0) || math.IsNaN(v1) || math.IsNaN(h) {
			t.Errorf("expected model is number, but got NaN")
		}
	}

	// Check if the file "vivaldi_node_position.txt" exists
	if _, err := os.Stat("vivaldi_node_position.txt"); os.IsNotExist(err) {
		t.Errorf("expected file vivaldi_node_position.txt to be created, but it does not exist")
	}
}
