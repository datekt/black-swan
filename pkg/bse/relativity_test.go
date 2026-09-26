package bse

import (
	"math"
	"testing"
)

func TestRelativityDisabledByDefault(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Sun", Mass: 1.989e30, Radius: 6.9634e8, Position: Vector3D{0, 0, 0}},
		{Name: "Mercury", Mass: 3.301e23, Radius: 2.44e6, Position: Vector3D{5.79e10, 0, 0}, Velocity: Vector3D{0, 47400, 0}},
	})

	if engine.EnableRelativity {
		t.Fatal("relativity must be off by default")
	}
}

func TestRelativityDivergesFromNewtonian(t *testing.T) {
	makeEngine := func(relativity bool) *Engine {
		e := NewEngine([]SpaceBody{
			{Name: "Sun", Mass: 1.989e30, Radius: 6.9634e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
			{Name: "Mercury", Mass: 3.301e23, Radius: 2.44e6, Position: Vector3D{5.79e10, 0, 0}, Velocity: Vector3D{0, 47400, 0}},
		})
		e.EnableRelativity = relativity
		return e
	}

	newtonian := makeEngine(false)
	relativistic := makeEngine(true)

	for i := 0; i < 5000; i++ {
		newtonian.Step(600)
		relativistic.Step(600)
	}

	nPos := newtonian.Bodies[1].Position
	rPos := relativistic.Bodies[1].Position

	diff := nPos.Sub(rPos).Length()
	if diff < 1e2 {
		t.Fatalf("GR correction too small to detect: %v meters", diff)
	}
	if diff > 1e9 {
		t.Fatalf("GR correction unphysically large: %v meters", diff)
	}
}

func TestRelativityPreservesMomentum(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Sun", Mass: 1.989e30, Radius: 6.9634e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, -0.0971, 0}},
		{Name: "Earth", Mass: 5.972e24, Radius: 6.371e6, Position: Vector3D{1.496e11, 0, 0}, Velocity: Vector3D{0, 29780, 0}},
		{Name: "Mars", Mass: 6.39e23, Radius: 3.389e6, Position: Vector3D{2.279e11, 0, 0}, Velocity: Vector3D{0, 24077, 0}},
	})
	engine.EnableRelativity = true

	pBefore := totalMomentum(engine.Bodies)
	scale := math.Max(math.Abs(pBefore.X), math.Max(math.Abs(pBefore.Y), math.Abs(pBefore.Z)))
	if scale == 0 {
		scale = 1
	}

	for i := 0; i < 1000; i++ {
		engine.Step(3600)
	}

	pAfter := totalMomentum(engine.Bodies)
	diff := pAfter.Sub(pBefore)

	if math.Abs(diff.X)/scale > 1e-10 || math.Abs(diff.Y)/scale > 1e-10 || math.Abs(diff.Z)/scale > 1e-10 {
		t.Fatalf("momentum drift with relativity: before %v after %v", pBefore, pAfter)
	}
}

func TestRelativityKeepsOrbitStable(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Sun", Mass: 1.989e30, Radius: 6.9634e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "Earth", Mass: 5.972e24, Radius: 6.371e6, Position: Vector3D{1.496e11, 0, 0}, Velocity: Vector3D{0, 29780, 0}},
	})
	engine.EnableRelativity = true

	initialRadius := engine.Bodies[1].Position.Length()

	minRadius := initialRadius
	maxRadius := initialRadius

	for i := 0; i < 20000; i++ {
		engine.Step(600)
		r := engine.Bodies[1].Position.Length()
		if r < minRadius {
			minRadius = r
		}
		if r > maxRadius {
			maxRadius = r
		}
	}

	if len(engine.Bodies) != 2 {
		t.Fatalf("Earth fell into the Sun under GR: %d bodies remain", len(engine.Bodies))
	}

	drift := (maxRadius - minRadius) / initialRadius
	if drift > 0.05 {
		t.Fatalf("orbit decayed or blew up under GR: drift %.3f (min %v max %v)", drift, minRadius, maxRadius)
	}
}

func TestRelativitySmallForSlowBodies(t *testing.T) {
	makeEngine := func(relativity bool) *Engine {
		e := NewEngine([]SpaceBody{
			{Name: "A", Mass: 1e24, Radius: 1e6, Position: Vector3D{-1e10, 0, 0}, Velocity: Vector3D{0, 10, 0}},
			{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e10, 0, 0}, Velocity: Vector3D{0, -10, 0}},
		})
		e.EnableRelativity = relativity
		return e
	}

	newtonian := makeEngine(false)
	relativistic := makeEngine(true)

	for i := 0; i < 1000; i++ {
		newtonian.Step(60)
		relativistic.Step(60)
	}

	nPos := newtonian.Bodies[0].Position
	rPos := relativistic.Bodies[0].Position

	diff := nPos.Sub(rPos).Length()
	scale := nPos.Length()

	if diff/scale > 1e-6 {
		t.Fatalf("GR correction unexpectedly large for slow bodies: rel diff %v", diff/scale)
	}
}
