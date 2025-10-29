package vcs

import (
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/sirupsen/logrus"
)

const (
	D = 3
)

func VirtualCoordinateSystem(K int) (map[enode.ID]int, map[int][]enode.ID, error) {
	// Kが1以上20以下でない場合にエラーを返す
	if K < 1 || K > 20 {
		return nil, nil, fmt.Errorf("the number of clusters must be between 1 and 20, but got %d", K)
	}

	var err error
	if err = GenerateVirtualCoordinate(D); err != nil {
		fmt.Printf("Failed to generate virtual coordinate %v\n", err)
		return nil, nil, err
	}

	var clusterResult map[enode.ID]int
	var clusterList map[int][]enode.ID

	clusterResult, clusterList, err = KMeansBasedOnVirtualCoordinate(K)
	if err != nil {
		fmt.Printf("Failed to cluster nodes %v\n", err)
		return nil, nil, err
	}

	// fmt.Printf("ClusterResult: %v Len: %d\n", clusterResult, len(clusterResult))
	// fmt.Printf("ClusterList: ")
	// for _, cluster := range clusterList {
	// 	fmt.Printf("%d ", len(cluster))
	// }
	// fmt.Println()

	// クラスターのサイズをフィールドに追加
	clusterSizes := make(map[string]int)
	for i, cluster := range clusterList {
		clusterSizes[fmt.Sprintf("Cluster%d", i)] = len(cluster)
	}

	// ログに出力
	log.WithFields(logrus.Fields{
		"clusterSizes": clusterSizes,
	}).Infoln("Cluster list sizes logged")

	return clusterResult, clusterList, nil
}
