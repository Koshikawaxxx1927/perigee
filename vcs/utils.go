package vcs

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Placeholder function for generating random values between 0 and 1
func RandomBetween0And1() float64 {
	// シードを設定（これにより毎回異なる乱数が生成される）
	// ローカル乱数生成器を作成
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 0から1の間の乱数を生成
	return rng.Float64()
}

// Placeholder function for generating random values between -1 and 1
func randomBetweenMinus1And1() float64 {
	return RandomBetween0And1()*2.0 - 1.0
}

// calculateMedian はRTT結果の中央値を計算する関数
func calculateMedian(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0 // 空の場合は0を返す
	}
	sort.Float64s(values) // RTTのスライスを昇順にソート
	if n%2 == 0 {
		// 要素数が偶数の場合は中央2つの平均を返す
		return (values[n/2-1] + values[n/2]) / 2
	}
	// 要素数が奇数の場合は中央の値を返す
	return values[n/2]
}

// フィルタ関数
func FilterNode(nodes []*enode.Node, predicate func(node *enode.Node) bool) []*enode.Node {
	result := []*enode.Node{}
	for _, node := range nodes {
		if predicate(node) {
			result = append(result, node)
		}
	}
	return result
}

func inverseLog(x float64) float64 {
	return 1.0 / math.Log10(1.0+x)
}

func RandomShuffle(peers []*peer.AddrInfo) {
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DecreaseExpoenentially(x int) float64 {
	return math.Exp(-2 * float64(x))
}
