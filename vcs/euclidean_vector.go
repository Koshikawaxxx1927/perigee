package vcs

import (
	"fmt"
	"math"
	"strings"
)

const FLOAT_ZERO = 1e-9

// EuclideanVector struct
type EuclideanVector struct {
	V []float64
}

// Constructor for EuclideanVector with dimension D
func NewEuclideanVector(d int) EuclideanVector {
	v := make([]float64, d)
	for i := 0; i < d; i++ {
		v[i] = randomBetweenMinus1And1() * 100
	}
	return EuclideanVector{V: v}
}

// Constructor for EuclideanVector with initialized values
func NewEuclideanVectorFromArray(arr []float64) EuclideanVector {
	return EuclideanVector{V: arr}
}

// Magnitude calculates the vector magnitude (Euclidean norm)
func (ev EuclideanVector) Magnitude() float64 {
	var sum float64
	for _, val := range ev.V {
		sum += val * val
	}
	return math.Sqrt(sum)
}

// IsZero checks if the vector is a zero vector
func (ev EuclideanVector) IsZero() bool {
	for _, val := range ev.V {
		if math.Abs(val) > FLOAT_ZERO {
			return false
		}
	}
	return true
}

// Show prints the vector components with 2 decimal precision
func (ev EuclideanVector) Show() {
	for _, val := range ev.V {
		fmt.Printf("%.2f ", val)
	}
	fmt.Println()
}

// Display returns the vector components as a formatted string
func (ev EuclideanVector) Display() string {
	var sb strings.Builder
	for _, val := range ev.V {
		// エラーチェックを追加
		if _, err := sb.WriteString(fmt.Sprintf("%.2f ", val)); err != nil {
			// エラーが発生した場合の処理を追加（例: ログ出力）
			fmt.Println("Error writing to strings.Builder:", err)
			return "" // エラー時には空文字列を返す
		}
	}
	return sb.String()
}

// Add operator (overload equivalent in Go)
func Add(a, b EuclideanVector) EuclideanVector {
	d := len(a.V)
	c := NewEuclideanVector(d)
	for i := 0; i < d; i++ {
		c.V[i] = a.V[i] + b.V[i]
	}
	return c
}

// Sub operator (overload equivalent in Go)
func Sub(a, b EuclideanVector) EuclideanVector {
	d := len(a.V)
	c := NewEuclideanVector(d)
	for i := 0; i < d; i++ {
		c.V[i] = a.V[i] - b.V[i]
	}
	return c
}

// Mul multiplies the vector by a scalar
func Mul(a EuclideanVector, b float64) EuclideanVector {
	d := len(a.V)
	c := NewEuclideanVector(d)
	for i := 0; i < d; i++ {
		c.V[i] = a.V[i] * b
	}
	return c
}

// Div divides the vector by a scalar (check for division by zero)
func Div(a EuclideanVector, b float64) EuclideanVector {
	if b == 0 {
		fmt.Println("Division by zero")
		return NewEuclideanVector(len(a.V))
	}
	return Mul(a, 1.0/b)
}

// Distance calculates the Euclidean distance between two vectors
func Distance(a, b EuclideanVector) float64 {
	c := Sub(a, b)
	return c.Magnitude()
}

// Dot calculates the dot product of two vectors
func Dot(a, b EuclideanVector) float64 {
	var ret float64
	for i := 0; i < len(a.V); i++ {
		ret += a.V[i] * b.V[i]
	}
	return ret
}

// RandomUnitVector generates a random unit vector
func RandomUnitVector(d int) EuclideanVector {
	for {
		tmp := make([]float64, d)
		for i := 0; i < d; i++ {
			tmp[i] = RandomBetween0And1()
		}
		v := NewEuclideanVectorFromArray(tmp)
		if !v.IsZero() {
			return Div(v, v.Magnitude())
		}
	}
}
