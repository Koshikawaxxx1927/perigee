package vcs

import (
	"net"
	"sync"
	"testing"
)

func TestSetVCSListenersForServerTest(t *testing.T) {
	// モックデータを作成
	testEnodeAddresses := map[enode.ID]*Addresses{
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
	myEnode := enode.HexID("2a8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec011")
	SetVCSMyListener(net.ParseIP("10.0.0.1"), myEnode)
	for remoteId, address := range testEnodeAddresses {
		SetVCSListener(remoteId, address.IP)
	}

	// エノードの確認
	for enodeID, address := range testEnodeAddresses {
		if addr, exists := EnodeAddresses[enodeID]; exists {
			if addr.IP.String() != address.IP.String() || addr.VCSPort != 20000 {
				t.Errorf("Expected address %v, got %v", address, addr)
			}
		} else {
			t.Errorf("Enode %v not found in EnodeAddresses", enodeID)
		}
	}
}

func TestVCSServerClient(t *testing.T) {
	ip := net.ParseIP("127.0.0.1") // ローカルIP
	port := uint(8080)             // テスト用のポート

	// VCSでCoordinateを通信するためのサーバを起動(サーバー側の操作)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := StartVCSServer(ip, port, &wg)
		if err != nil {
			t.Errorf("VCSサーバー起動中にエラー: %v", err)
		}
	}()
	wg.Wait()

	// 受信したレスポンスを解析(クライアント側の操作)
	coordinate, rtt, err := FetchCoordinateWithRTT(ip, port)
	if err != nil {
		t.Errorf("フェッチ中にエラー: %v", err)
	}
	// Coordinateレスポンスを準備
	expectedCoordinate := NewCoordinate(2)
	expectedRTT := 1.0

	if coordinate.H != expectedCoordinate.H {
		t.Errorf("Hが一致しません。期待値: %v, 実際の値: %v", expectedCoordinate.H, coordinate.H)
	}
	if coordinate.E != expectedCoordinate.E {
		t.Errorf("Eが一致しません。期待値: %v, 実際の値: %v", expectedCoordinate.E, coordinate.E)
	}
	if rtt != expectedRTT {
		t.Errorf("Expected RTT: %v, but got: %v", expectedRTT, rtt)
	}
}
