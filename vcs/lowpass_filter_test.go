package vcs

import (
    "math"
    "testing"
)

func floatsAlmostEqual(a, b float64, tol float64) bool {
    return math.Abs(a-b) <= tol
}

func TestButterworthFilter_StepResponse(t *testing.T) {
    filter := NewButterworthFilter(b, a, zi)

    // ステップ入力 (100個の 1.0)
    inputs := make([]float64, 100)
    for i := range inputs {
        inputs[i] = 1.0
    }

    outputs := make([]float64, len(inputs))
    for i, x := range inputs {
        outputs[i] = filter.LFilter(x)
    }

    // 最後の出力が 1.0 に近づいていればOK（フィルタの特性上、最終的に1に収束）
    finalOutput := outputs[len(outputs)-1]
    if !floatsAlmostEqual(finalOutput, 1.0, 0.01) { // 許容誤差 ±0.01
        t.Errorf("final output = %v, want around 1.0", finalOutput)
    }
}

func TestButterworthFilter_ZeroInput(t *testing.T) {
    filter := NewButterworthFilter(b, a, zi)

    // ゼロ入力 (すべて0.0)
    inputs := make([]float64, 200)
    outputs := make([]float64, len(inputs))
    for i, x := range inputs {
        outputs[i] = filter.LFilter(x)
    }

    // ゼロ入力なら最終的には出力もゼロに収束するはず
    finalOutput := outputs[len(outputs)-1]
    if !floatsAlmostEqual(finalOutput, 0.0, 0.01) { // 許容誤差 ±0.01
        t.Errorf("final output = %v, want around 0.0", finalOutput)
    }
}
