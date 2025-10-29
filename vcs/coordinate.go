package vcs

import "fmt"

// Coordinate struct
type Coordinate struct {
	V EuclideanVector
	H float64
	E float64
}

// Constructor for Coordinate with default values
func NewCoordinate(D int) Coordinate {
	return Coordinate{
		V: NewEuclideanVector(D), // Assuming 2D by default, change as necessary
		H: 0,
		E: 0,
	}
}

// Constructor for Coordinate with specified values
func NewCoordinateWithValues(vec EuclideanVector, height, e float64) Coordinate {
	return Coordinate{
		V: vec,
		H: height,
		E: e,
	}
}

// Vector returns the Euclidean vector
func (c Coordinate) Vector() EuclideanVector {
	return c.V
}

// Height returns the height
func (c Coordinate) Height() float64 {
	return c.H
}

// Error returns the error
func (c Coordinate) Error() float64 {
	return c.E
}

// Show prints the vector, height, and error
func (c Coordinate) Show() {
	fmt.Printf("vector = (")
	for i, val := range c.V.V {
		fmt.Printf("%.2f", val)
		if i < len(c.V.V)-1 {
			fmt.Printf(", ")
		}
	}
	fmt.Printf("), height = %.2f, error = %.2f\n", c.H, c.E)
}

// EstimateRTT calculates the estimated round-trip time between two coordinates
func EstimateRTT(a, b Coordinate) float64 {
	return Distance(a.Vector(), b.Vector()) + a.Height() + b.Height()
}

// Norm calculates the norm of a Euclidean vector
func Norm(a EuclideanVector) float64 {
	return a.Magnitude()
}

// Norm for a double value
func NormDouble(a float64) float64 {
	return a
}
