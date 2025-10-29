package vcs

import (
	"fmt"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	COORDINATE_UPDATE_ROUND = 10 // Set this to your desired number of rounds
)

// observeNeighbor は yenode を引数に受け取り、RTT 計測を行って Observe 処理を実行します
func observeNeighbor(yenode enode.ID, wg *sync.WaitGroup, xenode enode.ID, mutex *sync.Mutex) {
	defer wg.Done() // Goルーチンが終了したことを通知
	if VivaldiModels[xenode] == nil {
		fmt.Printf("VivaldiModel for my enode: %v is nil\n", xenode)
		return
	}
	if VivaldiModels[yenode] == nil {
		fmt.Printf("VivaldiModel for remote enode: %v is nil\n", yenode)
		return
	}

	_, found := EnodeAddresses[yenode]

	if !found {
		// yenode に該当するノードが見つからなかった場合の処理
		fmt.Println("yenode not found in vcsNodes")
		return // 見つからなかった場合は処理を中断
	}

	// インデックスが見つかったので、rtt を計測する
	yenodeIP := EnodeAddresses[yenode].IP
	yenodePort := EnodeAddresses[yenode].VCSPort
	newCoord, rtt, err := FetchCoordinateWithRTT(yenodeIP, yenodePort)

	if err != nil && rtt == 0 {
		fmt.Printf("error when measuring rtt for yenode %v: %v\n", yenode, err)
		return
	}

	if newCoord != nil {
		// log.Infof("New Coord %v", newCoord)
		VivaldiModels[yenode].SetCoordinate(*newCoord)
	}

	// ローパスフィルタによる処理
	if useLowPassFilter {
		filtered_rtt := filters[yenode].LFilter(rtt)
		log.Infof("Filter rtt to %v: %f to %f\n", yenodeIP, rtt, filtered_rtt)
		rtt = filtered_rtt
	}

	// 排他制御：Observe を呼び出す前にロック
	mutex.Lock()
	defer mutex.Unlock()

	// RTT 計測後、Observe 処理を実行
	VivaldiModels[xenode].Observe(yenode, VivaldiModels[yenode].Coordinate(), rtt)
}

// GenerateVirtualCoordinate generates virtual coordinates for Vivaldi models
func GenerateVirtualCoordinate(D int) error {
	log.Info("Started generating virtual coordinate")
	for id, address := range EnodeAddresses {
		if id == MyEnode {
			continue
		}
		if VivaldiModels[id] == nil {
			VivaldiModels[id] = NewVivaldiModel(D, id, address.IP, true, false, true)
		}
	}

	// Open output file
	// output, err := os.Create("vivaldi_node_position.txt")
	// ファイルを追記モードで開く（存在しない場合は新規作成）
	output, err := os.OpenFile("vivaldi_node_position.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("cannot open file: %s\n", "vivaldi_node_position.txt")
		return err
	}

	defer func() {
		if err := output.Close(); err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}()

	// Update coordinates over a number of rounds
	for round := 0; round < COORDINATE_UPDATE_ROUND; round++ {
		// Goルーチンを使って observeNeighbor を並行処理
		var wg sync.WaitGroup
		// Mutexを使って計測したRTTの記録を排他制御にする
		var mutex sync.Mutex

		x_enode := MyEnode

		for y_enode := range VivaldiModels {
			if x_enode == y_enode {
				continue
			}
			wg.Add(1)
			go observeNeighbor(y_enode, &wg, x_enode, &mutex)
		}
		wg.Wait()

		if _, err := fmt.Fprintf(output, "Round %d-%d\n", calledCount, round+1); err != nil {
			fmt.Printf("error writing to file at round %d: %v\n", round, err)
			return err
		}
		log.Infof("Vivaldi update: Round %d-%d", calledCount, round+1)
		for j := range VivaldiModels {
			ip := EnodeAddresses[j].IP
			v := VivaldiModels[j].Vector().V
			h := VivaldiModels[j].Coordinate().H
			e := VivaldiModels[j].Coordinate().E
			if _, err := fmt.Fprintf(output, "%s %s %.2f %.2f %.2f %f\n", j, ip, v[0], v[1], h, e); err != nil {
				fmt.Printf("error writing to file at round %d: %v\n", round, err)
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(output); err != nil {
		fmt.Printf("error writing newline to file: %v\n", err)
		return err
	}
	log.Info("Generated virtual coordinate")
	return nil
}
