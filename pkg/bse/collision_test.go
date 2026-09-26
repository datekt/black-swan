package bse

import (
	"math"
	"testing"
)

func TestMergePreservesMass(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Big", Mass: 1e30, Radius: 1e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "Small", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e8, 0, 0}, Velocity: Vector3D{0, 0, 0}},
	})

	events := engine.resolveCollisions()

	if len(events) != 1 {
		t.Fatalf("expected 1 merge, got %d", len(events))
	}
	if events[0].Absorber != "Big" || events[0].Absorbed != "Small" {
		t.Fatalf("wrong merge direction: %+v", events[0])
	}
	if len(engine.Bodies) != 1 {
		t.Fatalf("expected 1 body after merge, got %d", len(engine.Bodies))
	}
	if math.Abs(engine.Bodies[0].Mass-1.000001e30) > 1e20 {
		t.Fatalf("mass not conserved: %v", engine.Bodies[0].Mass)
	}
}

func TestMergePreservesMomentum(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Big", Mass: 1e30, Radius: 1e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{100, 0, 0}},
		{Name: "Small", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e8, 0, 0}, Velocity: Vector3D{0, 5000, 0}},
	})

	pBefore := totalMomentum(engine.Bodies)
	engine.resolveCollisions()
	pAfter := totalMomentum(engine.Bodies)

	diff := pAfter.Sub(pBefore)
	if math.Abs(diff.X) > 1e-10 || math.Abs(diff.Y) > 1e-10 || math.Abs(diff.Z) > 1e-10 {
		t.Fatalf("momentum not conserved: before %v after %v", pBefore, pAfter)
	}
}

func TestMergeCascades(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e30, Radius: 1e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e8, Position: Vector3D{1e8, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "C", Mass: 1e22, Radius: 1e8, Position: Vector3D{2e8, 0, 0}, Velocity: Vector3D{0, 0, 0}},
	})

	events := engine.resolveCollisions()
	if len(events) != 2 {
		t.Fatalf("expected 2 cascaded merges, got %d", len(events))
	}
	if len(engine.Bodies) != 1 {
		t.Fatalf("expected 1 body after cascade, got %d", len(engine.Bodies))
	}
}

func TestNoMergeWhenApart(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e30, Radius: 1e6, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e9, 0, 0}, Velocity: Vector3D{0, 0, 0}},
	})

	events := engine.resolveCollisions()
	if len(events) != 0 {
		t.Fatalf("unexpected merges: %+v", events)
	}
	if len(engine.Bodies) != 2 {
		t.Fatalf("bodies should not merge, got %d", len(engine.Bodies))
	}
}

func TestBlackHoleKeepsSchwarzschildRadius(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Rogue Black Hole", Mass: 1.989e31, Radius: 30000, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "Sun", Mass: 1.989e30, Radius: 6.9634e8, Position: Vector3D{1000, 0, 0}, Velocity: Vector3D{0, 0, 0}},
	})

	engine.resolveCollisions()

	if len(engine.Bodies) != 1 {
		t.Fatalf("expected 1 body after merge, got %d", len(engine.Bodies))
	}

	totalMass := 1.989e31 + 1.989e30
	expected := 2 * 6.67430e-11 * totalMass / (299792458.0 * 299792458.0)

	got := engine.Bodies[0].Radius
	if math.Abs(got-expected)/expected > 1e-9 {
		t.Fatalf("Schwarzschild radius: got %v want %v", got, expected)
	}
	if got > 1e5 {
		t.Fatalf("black hole radius should stay sub-kilometer-scale, got %v", got)
	}
}
