package vcs

import "testing"

func TestRunPingCommand(t *testing.T) {
	// テスト用の IP アドレス
	ip := "192.168.1.1"

	// runPingCommand を呼び出す
	rtt, err := runPingCommand(ip)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// テストモードでは、RTT が 1.0 になることを期待
	expectedRTT := 1.0

	// RTT の値をチェック
	if rtt != expectedRTT {
		t.Errorf("Test failed: expected RTT %v, but got %v", expectedRTT, rtt)
	} else {
		t.Logf("Test passed: RTT is as expected (%v)", rtt)
	}
}
