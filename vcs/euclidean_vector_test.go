package vcs

import (
	"math"
	"testing"
)

func TestNewEuclideanVector(t *testing.T) {
	v := NewEuclideanVector(3)
	if len(v.V) != 3 {
		t.Errorf("Expected vector length of 3, got %d", len(v.V))
	}
	if math.IsNaN(v.V[0]) || math.IsNaN(v.V[1]) || math.IsNaN(v.V[2]) {
		t.Errorf("expected model is number, but got NaN")
	}
}

func TestNewEuclideanVectorEdgeCases(t *testing.T) {
	v := NewEuclideanVector(0)
	if len(v.V) != 0 {
		t.Errorf("Expected vector length of 0, got %d", len(v.V))
	}
}

func TestNewEuclideanVectorFromArray(t *testing.T) {
	arr := []float64{1.0, 2.0, 3.0}
	v := NewEuclideanVectorFromArray(arr)
	if v.V[0] != 1.0 || v.V[1] != 2.0 || v.V[2] != 3.0 {
		t.Errorf("Expected vector [1.0, 2.0, 3.0], got %v", v.V)
	}
}

func TestNewEuclideanVectorFromArrayEmpty(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{})
	if len(v.V) != 0 {
		t.Errorf("Expected empty vector, got %v", v.V)
	}
}

func TestMagnitude(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{3.0, 4.0})
	expected := 5.0 // sqrt(3^2 + 4^2) = 5
	if math.Abs(v.Magnitude()-expected) > FLOAT_ZERO {
		t.Errorf("Expected magnitude %f, got %f", expected, v.Magnitude())
	}
}

func TestIsZero(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{0.0, 0.0})
	if !v.IsZero() {
		t.Errorf("Expected vector to be zero")
	}
	v = NewEuclideanVectorFromArray([]float64{1.0, 0.0})
	if v.IsZero() {
		t.Errorf("Expected vector not to be zero")
	}
}

func TestAdd(t *testing.T) {
	v1 := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	v2 := NewEuclideanVectorFromArray([]float64{3.0, 4.0})
	expected := []float64{4.0, 6.0}
	result := Add(v1, v2)
	for i := range result.V {
		if math.Abs(result.V[i]-expected[i]) > FLOAT_ZERO {
			t.Errorf("Expected %f, got %f", expected[i], result.V[i])
		}
	}
}

func TestSub(t *testing.T) {
	v1 := NewEuclideanVectorFromArray([]float64{5.0, 7.0})
	v2 := NewEuclideanVectorFromArray([]float64{3.0, 4.0})
	expected := []float64{2.0, 3.0}
	result := Sub(v1, v2)
	for i := range result.V {
		if math.Abs(result.V[i]-expected[i]) > FLOAT_ZERO {
			t.Errorf("Expected %f, got %f", expected[i], result.V[i])
		}
	}
}

func TestMul(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	scalar := 2.0
	expected := []float64{2.0, 4.0}
	result := Mul(v, scalar)
	for i := range result.V {
		if math.Abs(result.V[i]-expected[i]) > FLOAT_ZERO {
			t.Errorf("Expected %f, got %f", expected[i], result.V[i])
		}
	}
}

func TestDiv(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{4.0, 6.0})
	scalar := 2.0
	expected := []float64{2.0, 3.0}
	result := Div(v, scalar)
	for i := range result.V {
		if math.Abs(result.V[i]-expected[i]) > FLOAT_ZERO {
			t.Errorf("Expected %f, got %f", expected[i], result.V[i])
		}
	}
}

func TestDivByZero(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	result := Div(v, 0)
	for _, val := range result.V {
		if math.Abs(val) < 0.0 && 100.0 < math.Abs(val) {
			t.Errorf("Expected zero vector when dividing by zero, got %v", result.V)
		}
	}
}

func TestDistance(t *testing.T) {
	v1 := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	v2 := NewEuclideanVectorFromArray([]float64{4.0, 6.0})
	expected := 5.0 // sqrt((4-1)^2 + (6-2)^2) = 5
	result := Distance(v1, v2)
	if math.Abs(result-expected) > FLOAT_ZERO {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestDot(t *testing.T) {
	v1 := NewEuclideanVectorFromArray([]float64{1.0, 2.0, 3.0})
	v2 := NewEuclideanVectorFromArray([]float64{4.0, 5.0, 6.0})
	expected := 32.0 // 1*4 + 2*5 + 3*6 = 32
	result := Dot(v1, v2)
	if math.Abs(result-expected) > FLOAT_ZERO {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRandomUnitVector(t *testing.T) {
	v := RandomUnitVector(3)
	if math.Abs(v.Magnitude()-1.0) > FLOAT_ZERO {
		t.Errorf("Expected magnitude of 1, got %f", v.Magnitude())
	}
}

func TestRandomUnitVectorFixedSeed(t *testing.T) {
	v := RandomUnitVector(3)
	if math.Abs(v.Magnitude()-1.0) > FLOAT_ZERO {
		t.Errorf("Expected magnitude of 1, got %f", v.Magnitude())
	}
}

func TestDisplayErrorHandling(t *testing.T) {
	arr := []float64{1.0, 2.0, 3.0}
	v := NewEuclideanVectorFromArray(arr)
	if v.Display() != "1.00 2.00 3.00 " {
		t.Errorf("Expected formatted string '1.00 2.00 3.00 ', got %s", v.Display())
	}
}

func TestNewEuclideanVectorLargeDimension(t *testing.T) {
	d := 10000
	v := NewEuclideanVector(d)
	if len(v.V) != d {
		t.Errorf("Expected vector length of %d, got %d", d, len(v.V))
	}
}

func TestDivZeroVectorByScalar(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{0.0, 0.0, 0.0})
	scalar := 2.0
	result := Div(v, scalar)
	for _, val := range result.V {
		if math.Abs(val) > FLOAT_ZERO {
			t.Errorf("Expected zero vector, got %v", result.V)
		}
	}
}

func TestAddWithZeroVector(t *testing.T) {
	v1 := NewEuclideanVectorFromArray([]float64{1.0, 2.0, 3.0})
	v2 := NewEuclideanVectorFromArray([]float64{0.0, 0.0, 0.0})
	expected := []float64{1.0, 2.0, 3.0}
	result := Add(v1, v2)
	for i := range result.V {
		if math.Abs(result.V[i]-expected[i]) > FLOAT_ZERO {
			t.Errorf("Expected %f, got %f", expected[i], result.V[i])
		}
	}
}

func TestSubWithZeroVector(t *testing.T) {
	v1 := NewEuclideanVectorFromArray([]float64{1.0, 2.0, 3.0})
	v2 := NewEuclideanVectorFromArray([]float64{0.0, 0.0, 0.0})
	expected := []float64{1.0, 2.0, 3.0}
	result := Sub(v1, v2)
	for i := range result.V {
		if math.Abs(result.V[i]-expected[i]) > FLOAT_ZERO {
			t.Errorf("Expected %f, got %f", expected[i], result.V[i])
		}
	}
}

func TestMagnitudeOfLargeValues(t *testing.T) {
	v := NewEuclideanVectorFromArray([]float64{1e10, 1e10})
	expected := math.Sqrt(1e10*1e10 + 1e10*1e10)
	if math.Abs(v.Magnitude()-expected) > FLOAT_ZERO {
		t.Errorf("Expected magnitude %f, got %f", expected, v.Magnitude())
	}
}

func TestRandomUnitVectorHighDimension(t *testing.T) {
	v := RandomUnitVector(100)
	if math.Abs(v.Magnitude()-1.0) > FLOAT_ZERO {
		t.Errorf("Expected magnitude of 1, got %f", v.Magnitude())
	}
}

func TestRandomUnitVectorRepeated(t *testing.T) {
	for i := 0; i < 100; i++ {
		v := RandomUnitVector(3)
		if math.Abs(v.Magnitude()-1.0) > FLOAT_ZERO {
			t.Errorf("Iteration %d: Expected magnitude of 1, got %f", i, v.Magnitude())
		}
	}
}

func TestDotWithZeroVector(t *testing.T) {
	v1 := NewEuclideanVectorFromArray([]float64{1.0, 2.0, 3.0})
	v2 := NewEuclideanVectorFromArray([]float64{0.0, 0.0, 0.0})
	expected := 0.0
	result := Dot(v1, v2)
	if math.Abs(result-expected) > FLOAT_ZERO {
		t.Errorf("Expected dot product %f, got %f", expected, result)
	}
}
