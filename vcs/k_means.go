package vcs

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	maxIter = 100 // 最大イテレーション数
)

// kMeansBasedOnVirtualCoordinate performs K-means clustering based on virtual coordinates.
func KMeansBasedOnVirtualCoordinate(K int) (map[enode.ID]int, map[int][]enode.ID, error) {
	log.Info("Started clustering nodes on the generated virtual coordinate")
	n := len(EnodeAddresses) // The number of nodes
	clusterCnt := make([]int, K)
	clusterResult := make(map[enode.ID]int, n)
	center := make([]EuclideanVector, K)
	avg := make([]EuclideanVector, K)
	clusterList := make(map[int][]enode.ID, K)

	// 初期クラスタ中心をランダムに選択
	tmpList := make(map[enode.ID]struct{})
	for i := 0; i < K; i++ {
		for {
			u_enode, _ := RandomEnodeAddresses()
			if _, exists := tmpList[u_enode]; !exists {
				center[i] = VivaldiModels[u_enode].Vector()
				tmpList[u_enode] = struct{}{}
				break
			}
		}
	}

	// K-means アルゴリズム
	for iter := 0; iter < maxIter; iter++ {
		// 最も近い中心を見つける
		for i, vivaldiModel := range VivaldiModels {
			dist := 1e100
			curCluster := 0
			for j := 0; j < K; j++ {
				d := Distance(center[j], vivaldiModel.Vector())
				if d < dist {
					dist = d
					curCluster = j
				}
			}
			clusterResult[i] = curCluster
		}

		// 中心を再計算
		for i := range avg {
			avg[i] = NewEuclideanVectorFromArray([]float64{0.0, 0.0, 0.0}) // 0に初期化
			clusterCnt[i] = 0
		}
		for i, vivaldiModel := range VivaldiModels {
			avg[clusterResult[i]] = Add(avg[clusterResult[i]], vivaldiModel.Vector())
			clusterCnt[clusterResult[i]]++
		}
		for i := 0; i < K; i++ {
			if clusterCnt[i] > 0 {
				center[i] = Div(avg[i], float64(clusterCnt[i]))
			}
		}
	}

	// クラスタ結果の出力
	for i := range clusterList {
		clusterList[i] = []enode.ID{} // 初期化
	}
	for i := range VivaldiModels {
		clusterList[clusterResult[i]] = append(clusterList[clusterResult[i]], i)
	}

	// // 結果を表示
	// fmt.Println("Cluster result:")
	// for i := 0; i < K; i++ {
	// 	fmt.Printf("Cluster %d size: %d\n", i, len(clusterList[i]))
	// }

	// // 結果をファイルに書き込む
	// output_cluster, err_cluster := os.Create("node_by_cluster.txt")
	// if err_cluster != nil {
	// 	fmt.Printf("cannot open file: %s\n", "node_by_cluster.txt")
	// 	// return
	// }
	// defer func() {
	// 	if err_cluster := output_cluster.Close(); err_cluster != nil {
	// 		fmt.Printf("error closing file: %v\n", err)
	// 	}
	// }()

	// for _, cluster := range clusterResult {
	// 	// fmt.Printf("%d ", node)
	// 	if _, err_cluster := fmt.Fprintf(output_cluster, "%d ", cluster); err_cluster != nil {
	// 		fmt.Printf("error writing to file %v\n", err)
	// 		return nil, nil, err
	// 	}
	// }
	// log.Info("Clustered nodes on the generated virtual coordinate")

	// Open output file
	// output, err_cluster := os.Create("vivaldi_cluster_result.txt")
	// ファイルを追記モードで開く（存在しない場合は新規作成）
	output, err := os.OpenFile("vivaldi_cluster_result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("cannot open file: %s\n", "vivaldi_cluster_result.txt")
		return nil, nil, err
	}

	defer func() {
		if err := output.Close(); err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}()
	// clusterResult の内容を1行ずつ書き込む
	if _, err := fmt.Fprintln(output, "Cluster Result (NodeIP NodeID ClusterID):"); err != nil {
		fmt.Println("error writing header")
	}
	for nodeID, clusterID := range clusterResult {
		nodeIP := VivaldiModels[nodeID].SelfIP
		if _, err := fmt.Fprintf(output, "%s %s %d\n", nodeIP.String(), nodeID.String(), clusterID); err != nil {
			fmt.Printf("error writing node mapping: %v\n", err)
		}
	}

	return clusterResult, clusterList, nil
}
