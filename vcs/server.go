package vcs

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"sync"
)

// StartVCSServerはUDPサーバーを開始し、リクエストを受信するとCoordinateを返します
func StartVCSServer(_ip net.IP, port uint, wg *sync.WaitGroup) error {
	ip := _ip.String()
	// IPアドレスとポート番号を結合してアドレスを生成
	addr := fmt.Sprintf("%s:%d", ip, port)

	myIP := EnodeAddresses[MyEnode].IP
	VivaldiModels[MyEnode] = NewVivaldiModel(D, MyEnode, myIP, true, false, true)

	// UDPアドレスを解決
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve the address: %w", err)
	}

	// 指定されたUDPアドレスでリッスンを開始
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("Error when starting VCS server: %w", err)
	}
	// connがnilでない場合にCloseを遅延呼び出し
	if conn != nil {
		defer func() {
			if err := conn.Close(); err != nil {
				fmt.Println("Error when connecting the VCS server:", err)
			}
		}()
	}
	log.Infof("Started VCS server on %s:%d", ip, port)

	var coordinate Coordinate
	// テスト環境かどうかをチェック
	isTesting := flag.Lookup("test.v") != nil
	wg.Done()
	for {
		// 受信データを格納するバッファ
		buf := make([]byte, 1024)

		// リクエストを受信
		_, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Failed to read data from UDP:", err)
			continue
		}

		// Coordinateレスポンスを準備
		if isTesting {
			coordinate = NewCoordinate(2)
		} else {
			coordinate = VivaldiModels[MyEnode].Coordinate()
		}

		// Coordinate構造体をJSON形式に変換
		response, err := json.Marshal(coordinate)
		if err != nil {
			fmt.Println("レスポンスのマーシャリング中にエラー:", err)
			continue
		}

		// クライアントにレスポンスを送信
		_, err = conn.WriteToUDP(response, clientAddr)
		if err != nil {
			fmt.Println("レスポンスの送信中にエラー:", err)
			continue
		}
	}
}
