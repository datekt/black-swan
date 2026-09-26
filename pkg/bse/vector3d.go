package bse

import (
	"encoding/json"
	"fmt"
	"math"
)

type Vector3D struct {
	X float64
	Y float64
	Z float64
}

func (v Vector3D) MarshalJSON() ([]byte, error) {
	return json.Marshal([3]float64{v.X, v.Y, v.Z})
}

func (v *Vector3D) UnmarshalJSON(data []byte) error {
	var arr []float64
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("vector3d must be a 3-element array: %w", err)
	}
	if len(arr) != 3 {
		return fmt.Errorf("vector3d must have exactly 3 elements, got %d", len(arr))
	}
	v.X = arr[0]
	v.Y = arr[1]
	v.Z = arr[2]
	return nil
}

func (v Vector3D) Add(other Vector3D) Vector3D {
	return Vector3D{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

func (v Vector3D) Sub(other Vector3D) Vector3D {
	return Vector3D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

func (v Vector3D) Scale(scalar float64) Vector3D {
	return Vector3D{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

func (v Vector3D) Dot(other Vector3D) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v Vector3D) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vector3D) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

func (v Vector3D) Normalize() Vector3D {
	length := v.Length()
	if length == 0 {
		return Vector3D{0, 0, 0}
	}
	return Vector3D{
		X: v.X / length,
		Y: v.Y / length,
		Z: v.Z / length,
	}
}
