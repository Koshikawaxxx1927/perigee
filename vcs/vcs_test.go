package vcs

import (
	"net"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestVirtualCoordinateSystem(t *testing.T) {
	MyEnode = enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")
	EnodeAddresses = map[enode.ID]*Addresses{
		enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.1"), VCSPort: 30301},
		enode.HexID("7f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.2"), VCSPort: 30302},
		enode.HexID("8f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): {IP: net.ParseIP("192.168.1.3"), VCSPort: 30303},
		enode.HexID("9f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec022"): {IP: net.ParseIP("192.168.1.4"), VCSPort: 30304},
		enode.HexID("1f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec022"): {IP: net.ParseIP("192.168.1.5"), VCSPort: 30305},
		enode.HexID("2f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec022"): {IP: net.ParseIP("192.168.1.6"), VCSPort: 30306},
		enode.HexID("3f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec033"): {IP: net.ParseIP("192.168.1.7"), VCSPort: 30307},
		enode.HexID("4f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec033"): {IP: net.ParseIP("192.168.1.8"), VCSPort: 30308},
		enode.HexID("0f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec033"): {IP: net.ParseIP("192.168.1.9"), VCSPort: 30309},
		enode.HexID("5f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec044"): {IP: net.ParseIP("192.168.1.10"), VCSPort: 30310},
		enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec044"): {IP: net.ParseIP("192.168.1.11"), VCSPort: 30311},
		enode.HexID("7f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec044"): {IP: net.ParseIP("192.168.1.12"), VCSPort: 30312},
		enode.HexID("8f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec055"): {IP: net.ParseIP("192.168.1.13"), VCSPort: 30313},
		enode.HexID("9f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec055"): {IP: net.ParseIP("192.168.1.14"), VCSPort: 30314},
		enode.HexID("1f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec055"): {IP: net.ParseIP("192.168.1.15"), VCSPort: 30315},
		enode.HexID("2f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec066"): {IP: net.ParseIP("192.168.1.16"), VCSPort: 30316},
		enode.HexID("3f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec066"): {IP: net.ParseIP("192.168.1.17"), VCSPort: 30317},
		enode.HexID("4f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec066"): {IP: net.ParseIP("192.168.1.18"), VCSPort: 30318},
		enode.HexID("5f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec077"): {IP: net.ParseIP("192.168.1.19"), VCSPort: 30319},
		enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec077"): {IP: net.ParseIP("192.168.1.20"), VCSPort: 30320},
		enode.HexID("7f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec077"): {IP: net.ParseIP("192.168.1.21"), VCSPort: 30321},
		// 必要に応じてさらにデータを追加
	}

	// テストケースの定義
	tests := []struct {
		K                     int
		expectedErr           bool
		expectedResult        bool
		expectedClusterResult map[enode.ID]int
		expectedClusterList   map[int][]enode.ID
	}{
		// Kが1と20の範囲内のテストケース
		{K: 5, expectedErr: false, expectedResult: true, expectedClusterResult: map[enode.ID]int{enode.HexID("0f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): 1}, expectedClusterList: map[int][]enode.ID{1: {enode.HexID("0f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")}}},
		{K: 20, expectedErr: false, expectedResult: true, expectedClusterResult: map[enode.ID]int{enode.HexID("0f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011"): 1}, expectedClusterList: map[int][]enode.ID{1: {enode.HexID("0f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")}}},
		// Kが範囲外のテストケース
		{K: 0, expectedErr: true, expectedResult: false, expectedClusterResult: nil, expectedClusterList: nil},
		{K: 21, expectedErr: true, expectedResult: false, expectedClusterResult: nil, expectedClusterList: nil},
	}

	// テストの実行
	for _, test := range tests {
		t.Run("TestKValue", func(t *testing.T) {
			MyEnode = enode.HexID("6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")
			myIP := EnodeAddresses[MyEnode].IP
			VivaldiModels[MyEnode] = NewVivaldiModel(D, MyEnode, myIP, true, false, true)
			clusterResult, clusterList, err := VirtualCoordinateSystem(test.K)

			// エラーが予期される場合、エラーが発生したかをチェック
			if test.expectedErr {
				if err == nil {
					t.Errorf("expected error for K=%d, but got nil", test.K)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for K=%d: %v", test.K, err)
				}
			}

			// 結果が予期される場合、結果をチェック
			if test.expectedResult {
				if len(clusterResult) != len(EnodeAddresses) {
					t.Errorf("expected len(clusterResult) to be %d, but got %d", len(EnodeAddresses), len(clusterResult))
				}
				if len(clusterList) != test.K {
					t.Errorf("expected len(clusterList) to be %d, but got %d", test.K, len(clusterList))
				}
			} else {
				if len(clusterResult) != 0 {
					t.Errorf("expected empty clusterResult for K=%d, but got non-empty", test.K)
				}
				if len(clusterList) != 0 {
					t.Errorf("expected empty clusterList for K=%d, but got non-empty", test.K)
				}
			}
		})
	}
}

// テスト関数
func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		input    []float64
		expected float64
	}{
		{[]float64{}, 0},                // 空のスライス
		{[]float64{1}, 1},               // 要素が1つ
		{[]float64{1, 2, 3}, 2},         // 奇数個の要素
		{[]float64{1, 2, 3, 4}, 2.5},    // 偶数個の要素
		{[]float64{10, 20, 30}, 20},     // 大きい値を含む
		{[]float64{1.1, 2.2, 3.3}, 2.2}, // 小数点を含む
	}

	for i, test := range tests {
		result := calculateMedian(test.input)
		if result != test.expected {
			t.Errorf("Test %d failed: input=%v, expected=%v, got=%v", i, test.input, test.expected, result)
		}
	}
}
