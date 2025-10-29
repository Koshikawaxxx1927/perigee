package vcs

// テストコードが書けていない(ping用のモックを作らないといけないため)

import (
	"fmt"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/prysmaticlabs/prysm/v4/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var log = logrus.WithField("prefix", "vcs")

type ClusterMap struct {
	ClusterResult map[enode.ID]int
	ClusterList   map[int][]enode.ID
	Status        string
}

var (
	K                            int                                     // クラスターの数
	InnerCluster                 int                                     // 同じクラスター内のピアの数
	VCSPort                      uint                                    // VTSサーバーのポート番号
	VCSInterval                  int                                     // VCSの実行間隔
	calledCount                  = 0                                     // VCSを実行した回数
	isRunning                    = false                                 // VCSを実行中か(同時に実行を防ぐ)
	UseClusterBasedPeerSelection = false                                 // クラスター化に基づくピア選択を行うか
	EnableClusterBiasControl = false
	RunVCSChan                   = make(chan string, 1)                  // VCSを起動するときのチャネル
	DoneVCSChan                  = make(chan ClusterMap, 1)              //  VCSが終了したときのチャネル
	clusterResult                = make(map[enode.ID]int)                // ノードのenodeを指定して所属しているクラスターの番号を得る
	clusterList                  = make(map[int][]enode.ID)              // クラスター番号を指定して、そのクラスター番号に所属しているノードのenodeを格納した配列を得る
	useLowPassFilter             bool                                    // 観測したRTTに対して、ローパスフィルタを適用させるか
	filters                      = make(map[enode.ID]*ButterworthFilter) // ローパスフィルタ
)

func RegisterVCS(ctx *cli.Context) error {
	K = ctx.Int(cmd.NumberOfK.Name)
	InnerCluster = ctx.Int(cmd.InnerCluster.Name)
	VCSInterval = ctx.Int(cmd.VCSInterval.Name)
	VCSPort = 20000
	useLowPassFilter = ctx.Bool(cmd.UseLowPassFilter.Name)
	UseClusterBasedPeerSelection = ctx.Bool(cmd.UseClusterBasedPeerSelection.Name)
	EnableClusterBiasControl = ctx.Bool(cmd.EnableClusterBiasControl.Name)

	// fmt.Printf("K = %d\n", K)
	// fmt.Printf("VCS Interval = %d\n", VCSInterval)

	return nil
}

// SetVCSListenersはenrsの内容でvcsNodesを更新する関数
func SetVCSMyListener(myIP net.IP, myEnode enode.ID) {
	log.Debug("Setting the VCS information")
	MyEnode = myEnode
	EnodeAddresses[MyEnode] = &Addresses{IP: myIP, VCSPort: VCSPort}

	// if useLowPassFilter {
	// 	log.Debug("Initialized my filter")
	// 	filters[MyEnode] = NewButterworthFilter(b, a, zi)
	// }
}

// SetVCSListenersはenrsの内容でvcsNodesを更新する関数
func SetVCSListener(remoteEnode enode.ID, ip net.IP) {
	log.Info("Setting the VCS information")
	remoteVCSPort := uint(20000)
	EnodeAddresses[remoteEnode] = &Addresses{IP: ip, VCSPort: remoteVCSPort}

	if useLowPassFilter {
		log.Debug("Initialized my filter")
		filters[remoteEnode] = NewButterworthFilter(b, a, zi)
	}
}

func StartVCS() {
	var wg sync.WaitGroup
	wg.Add(1)
	calledCount++
	go func(calledCount int, wg *sync.WaitGroup) {
		if calledCount == 1 {
			// 最初だけサーバーを起動
			myIP := EnodeAddresses[MyEnode].IP
			err := StartVCSServer(myIP, VCSPort, wg)
			if err != nil {
				log.Error("Failed to start VCS server...")
			}
		} else {
			wg.Done()
		}
	}(calledCount, &wg)
	wg.Wait()
	if isRunning {
		return
	}
	isRunning = true
	if len(EnodeAddresses) == 0 || len(EnodeAddresses) < K {
		log.Infof("Cannot cluster nodes into %d when the found nodes are %d", K, len(EnodeAddresses))
		isRunning = false
		cm := ClusterMap{
			ClusterResult: nil,
			ClusterList:   nil,
			Status:        "Failed to run vcs",
		}
		log.Info("VCS has done but cannot cluster")
		DoneVCSChan <- cm
		return
	}
	var err error
	// _, _, err := VirtualCoordinateSystem(K) // VCSに基づいてピア選択を行う場合は上のコードを使う
	log.Info("Started setting the vritual coordinate system")
	clusterResult, clusterList, err = VirtualCoordinateSystem(K)
	if err != nil {
		// エラーが発生した場合、デフォルト値を設定
		fmt.Printf("Error in VirtualCoordinateSystem: %v", err)
	}
	isRunning = false
	cm := ClusterMap{
		ClusterResult: clusterResult,
		ClusterList:   clusterList,
		Status:        "Successfully ran vcs",
	}
	log.Info("VCS has done")
	DoneVCSChan <- cm
}

func GetClusterResult() (map[enode.ID]int, map[int][]enode.ID) {
	return clusterResult, clusterList
}
