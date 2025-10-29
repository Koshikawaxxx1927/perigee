package vcs

import (
	"net"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestKMeansBasedOnVirtualCoordinate(t *testing.T) {
	D := 3 // Dimensions
	K := 4

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
	var err error
	err = GenerateVirtualCoordinate(D)
	if err != nil {
		t.Fatalf("error generating virtual coordinates: %v", err)
	}

	// Ensure VivaldiModels is properly populated
	for _, model := range VivaldiModels {
		if model == nil {
			t.Fatalf("VivaldiModel is nil for a node")
		}
	}

	// Run K-means clustering
	clusterResult, clusterList, err := KMeansBasedOnVirtualCoordinate(K)
	if err != nil {
		t.Errorf("Failed to cluster on virtual coordinate: %s", err)
	}

	// Check the results
	clusterCount := make([]int, K)
	for _, clusterIndex := range clusterResult {
		if clusterIndex < 0 || clusterIndex >= K {
			t.Errorf("Invalid cluster index: %d", clusterIndex)
		}
		clusterCount[clusterIndex]++
	}

	// Ensure each cluster has at least one member
	for i, count := range clusterCount {
		if count == 0 {
			t.Errorf("Cluster %d has no members", i)
		}
	}

	// Check that the clusterList is populated
	for i := 0; i < K; i++ {
		if len(clusterList[i]) == 0 {
			t.Errorf("Cluster %d is empty after clustering", i)
		}
	}

	// Verify file output
	if _, err := os.Stat("node_by_cluster.txt"); os.IsNotExist(err) {
		t.Errorf("File node_by_cluster.txt was not created")
	}
}
