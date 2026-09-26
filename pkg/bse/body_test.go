package bse

import (
	"encoding/json"
	"testing"
)

func TestCloneIsIndependent(t *testing.T) {
	original := SpaceBody{
		Name:     "Earth",
		Mass:     5.972e24,
		Radius:   6.371e6,
		Position: Vector3D{1.496e11, 0, 0},
		Velocity: Vector3D{0, 29780, 0},
	}
	clone := original.Clone()
	clone.Position.X = 0
	clone.Velocity.Y = 0
	clone.Name = "Mutated"

	if original.Position.X == 0 {
		t.Fatal("Clone mutated original position")
	}
	if original.Velocity.Y == 0 {
		t.Fatal("Clone mutated original velocity")
	}
	if original.Name != "Earth" {
		t.Fatalf("Clone mutated original name: %s", original.Name)
	}
}

func TestSpaceBodyMarshalShape(t *testing.T) {
	body := SpaceBody{
		Name:     "Earth",
		Mass:     5.972e24,
		Radius:   6.371e6,
		Position: Vector3D{1, 2, 3},
		Velocity: Vector3D{4, 5, 6},
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}

	pos, ok := decoded["position"].([]interface{})
	if !ok {
		t.Fatalf("position is not a JSON array: %T", decoded["position"])
	}
	if len(pos) != 3 {
		t.Fatalf("position length: got %d want 3", len(pos))
	}

	vel, ok := decoded["velocity"].([]interface{})
	if !ok {
		t.Fatalf("velocity is not a JSON array: %T", decoded["velocity"])
	}
	if len(vel) != 3 {
		t.Fatalf("velocity length: got %d want 3", len(vel))
	}
}
