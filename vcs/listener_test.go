package vcs

import (
	"net"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/urfave/cli/v2"
)

func TestSetVCSListenersForListenerTest(t *testing.T) {
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

func TestRegisterVCS(t *testing.T) {
	// モック用の変数
	cmd := struct {
		NumberOfK   *cli.IntFlag
		VCSInterval *cli.IntFlag
	}{
		NumberOfK: &cli.IntFlag{
			Name:  "number-of-k",
			Usage: "The number of clusters on VCS",
			Value: 2, // デフォルト値
		},
		VCSInterval: &cli.IntFlag{
			Name:  "vcs-interval",
			Usage: "The interval of each vcs operation",
			Value: 5,
		},
	}

	// CLIアプリケーションを作成
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name: "vcs",
			Flags: []cli.Flag{
				cmd.NumberOfK,
				cmd.VCSInterval,
			},
			Action: RegisterVCS,
		},
	}

	// テスト1: デフォルト値
	t.Run("default value", func(t *testing.T) {
		// コマンドライン引数を指定
		args := []string{"app", "vcs"}

		// アプリケーション実行
		err := app.Run(args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// 値の確認
		if K != 2 {
			t.Errorf("expected K = 2, but got K = %d", K)
		}
		if VCSInterval != 5 {
			t.Errorf("expected VCSInterval = 5, but got VCSInterval = %d", VCSInterval)
		}
	})

	// テスト2: フラグを指定して実行
	t.Run("custom value", func(t *testing.T) {
		// コマンドライン引数を指定
		args := []string{"app", "vcs", "--number-of-k", "5", "--vcs-interval", "2"}

		// アプリケーション実行
		err := app.Run(args)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// 値の確認
		if K != 5 {
			t.Errorf("expected K = 5, but got K = %d", K)
		}
		if VCSInterval != 2 {
			t.Errorf("expected VCSInterval = 2, but got VCSInterval = %d", VCSInterval)
		}
	})
}
