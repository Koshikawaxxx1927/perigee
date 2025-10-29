package vcs

import (
	"io"
	"os"
	"testing"
)

func TestNewCoordinate(t *testing.T) {
	c := NewCoordinate(2)

	if c.Height() != 0 {
		t.Errorf("expected height to be 0, got %f", c.Height())
	}

	if c.Error() != 0 {
		t.Errorf("expected error to be 0, got %f", c.Error())
	}

	if len(c.Vector().V) != 2 { // Assuming default dimension is 2
		t.Errorf("expected vector dimension to be 2, got %d", len(c.Vector().V))
	}
}

func TestNewCoordinateWithValues(t *testing.T) {
	vec := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	height := 5.0
	errorVal := 1.5
	c := NewCoordinateWithValues(vec, height, errorVal)

	if c.Height() != height {
		t.Errorf("expected height to be %f, got %f", height, c.Height())
	}

	if c.Error() != errorVal {
		t.Errorf("expected error to be %f, got %f", errorVal, c.Error())
	}

	expectedVector := vec.V
	for i := range expectedVector {
		if c.Vector().V[i] != expectedVector[i] {
			t.Errorf("expected vector value at index %d to be %f, got %f", i, expectedVector[i], c.Vector().V[i])
		}
	}
}

func TestCoordinateShow(t *testing.T) {
	vec := NewEuclideanVectorFromArray([]float64{1.5, 2.5})
	c := NewCoordinateWithValues(vec, 3.0, 0.5)

	// Check if Show doesn't panic (printing function)
	c.Show()
}

func TestEstimateRTT(t *testing.T) {
	vecA := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	vecB := NewEuclideanVectorFromArray([]float64{3.0, 4.0})
	cA := NewCoordinateWithValues(vecA, 5.0, 0.5)
	cB := NewCoordinateWithValues(vecB, 2.0, 0.5)

	expectedRTT := Distance(cA.Vector(), cB.Vector()) + cA.Height() + cB.Height()
	actualRTT := EstimateRTT(cA, cB)

	if actualRTT != expectedRTT {
		t.Errorf("expected RTT to be %f, got %f", expectedRTT, actualRTT)
	}
}

func TestNorm(t *testing.T) {
	vec := NewEuclideanVectorFromArray([]float64{3.0, 4.0})
	expectedNorm := 5.0 // ||(3, 4)|| = 5

	actualNorm := Norm(vec)

	if actualNorm != expectedNorm {
		t.Errorf("expected norm to be %f, got %f", expectedNorm, actualNorm)
	}
}

func TestNormDouble(t *testing.T) {
	value := 7.0
	expected := 7.0

	if NormDouble(value) != expected {
		t.Errorf("expected norm of %f to be %f, got %f", value, expected, NormDouble(value))
	}
}

func TestNewCoordinateNegativeDimension(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for negative dimension, but no panic occurred")
		}
	}()
	NewCoordinate(-1)
}

func TestEstimateRTTWithZeroHeight(t *testing.T) {
	vecA := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	vecB := NewEuclideanVectorFromArray([]float64{3.0, 4.0})
	cA := NewCoordinateWithValues(vecA, 0.0, 0.5)
	cB := NewCoordinateWithValues(vecB, 0.0, 0.5)

	expectedRTT := Distance(cA.Vector(), cB.Vector())
	actualRTT := EstimateRTT(cA, cB)

	if actualRTT != expectedRTT {
		t.Errorf("expected RTT to be %f, got %f", expectedRTT, actualRTT)
	}
}

func TestShowOutput(t *testing.T) {
	vec := NewEuclideanVectorFromArray([]float64{1.0, 2.0})
	c := NewCoordinateWithValues(vec, 3.0, 0.5)

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Call Show
	c.Show()

	// Close the writer and restore stdout
	err = w.Close()
	if err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	os.Stdout = old

	// Read captured output
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}

	// Validate output
	expectedOutput := "vector = (1.00, 2.00), height = 3.00, error = 0.50\n"
	if string(output) != expectedOutput {
		t.Errorf("expected output %q, got %q", expectedOutput, string(output))
	}
}
