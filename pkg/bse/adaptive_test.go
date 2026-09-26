package bse

import (
	"math"
	"testing"
)

func TestNeedsRefinementFarBodies(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e24, Radius: 1e6, Position: Vector3D{0, 0, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e11, 0, 0}},
	})

	if engine.needsRefinement() {
		t.Fatal("distant bodies should not trigger refinement")
	}
}

func TestNeedsRefinementCloseBodies(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e24, Radius: 1e6, Position: Vector3D{0, 0, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{5e6, 0, 0}},
	})

	if !engine.needsRefinement() {
		t.Fatal("close bodies should trigger refinement")
	}
}

func TestNeedsRefinementIgnoresOverlapping(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e24, Radius: 1e6, Position: Vector3D{0, 0, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e5, 0, 0}},
	})

	if engine.needsRefinement() {
		t.Fatal("overlapping bodies will merge, refinement is not needed")
	}
}

func TestAdaptiveStepDisabledWhenDepthZero(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e24, Radius: 1e6, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{5e6, 0, 0}, Velocity: Vector3D{0, 1000, 0}},
	})
	engine.MaxSubstepDepth = 0

	engine.Step(10)

	if len(engine.Bodies) != 2 {
		t.Fatalf("without refinement close bodies should stay separate, got %d", len(engine.Bodies))
	}
}

func TestAdaptiveStepConservesMomentumOnCloseApproach(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Big", Mass: 1e30, Radius: 1e8, Position: Vector3D{-1e9, 0, 0}, Velocity: Vector3D{100, 0, 0}},
		{Name: "Small", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e9, 0, 0}, Velocity: Vector3D{-100, 0, 0}},
	})

	pBefore := totalMomentum(engine.Bodies)
	scale := math.Max(math.Abs(pBefore.X), math.Max(math.Abs(pBefore.Y), math.Abs(pBefore.Z)))
	if scale == 0 {
		scale = 1
	}

	for i := 0; i < 2000; i++ {
		engine.Step(50)
	}

	pAfter := totalMomentum(engine.Bodies)
	diff := pAfter.Sub(pBefore)

	if math.Abs(diff.X)/scale > 1e-10 || math.Abs(diff.Y)/scale > 1e-10 || math.Abs(diff.Z)/scale > 1e-10 {
		t.Fatalf("momentum drift: before %v after %v", pBefore, pAfter)
	}
}

func TestAdaptiveStepStillMergesOnCollision(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e30, Radius: 1e8, Position: Vector3D{-2e9, 0, 0}, Velocity: Vector3D{5000, 0, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{2e9, 0, 0}, Velocity: Vector3D{-5000, 0, 0}},
	})

	var events []MergeEvent
	for i := 0; i < 2000; i++ {
		events = append(events, engine.Step(100)...)
	}

	if len(events) != 1 {
		t.Fatalf("expected exactly 1 merge event, got %d", len(events))
	}
	if len(engine.Bodies) != 1 {
		t.Fatalf("expected 1 body after collision, got %d", len(engine.Bodies))
	}
}
