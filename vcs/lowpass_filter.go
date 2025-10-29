package vcs

// // ===== 4次バターワース =====
// // python の from scipy.signal import butter で出力
// // カットオフ周波数: 0.1Hz、サンプリング周波数: 10.0Hz、フィルタ次数: 4
// var b = []float64{
// 	8.98486146e-07, 3.59394459e-06, 5.39091688e-06, 3.59394459e-06,
// 	8.98486146e-07,
// }
// var a = []float64{
//     1.      ,   -3.83582554,  5.52081914, -3.53353522,  0.848556,
// }

// // 初期zi（Pythonでlfilter_ziで取得した値をそのまま使う）
// var zi = []float64{
//     0.9999991, -2.83583003, 2.68498371, -0.8485551,
// }
// // ========================

// ===== 4次バターワース =====
// python の from scipy.signal import butter で出力
// カットオフ周波数: 0.03Hz、サンプリング周波数: 10.0Hz、フィルタ次数: 4
var b = []float64{
    7.69909891e-09, 3.07963957e-08, 4.61945935e-08, 3.07963957e-08, 7.69909891e-09,
}
var a = []float64{
    1.        , -3.95074409,  5.85344172, -3.85463384, 0.95193634,
}
// 初期zi（Pythonでlfilter_ziで取得した値をそのまま使う）
var zi = []float64{
    0.99999999, -2.95074413, 2.90269755, -0.95193633,
}
// ========================


type ButterworthFilter struct {
    b  []float64
    a  []float64
    zi []float64
}

func NewButterworthFilter(b, a, zi []float64) *ButterworthFilter {
    return &ButterworthFilter{
        b:  b,
        a:  a,
        zi: append([]float64{}, zi...), // deep copy
    }
}

func (f *ButterworthFilter) LFilter(xn float64) float64 {
    yn := f.b[0]*xn + f.zi[0]

    for i := 1; i < len(f.zi); i++ {
        f.zi[i-1] = f.b[i]*xn + f.zi[i] - f.a[i]*yn
    }
    f.zi[len(f.zi)-1] = f.b[len(f.b)-1]*xn - f.a[len(f.a)-1]*yn

    return yn
}