package bse

import (
	"math"
	"testing"
)

func TestNewEngineDefaultG(t *testing.T) {
	engine := NewEngine(nil)
	if math.Abs(engine.G-6.67430e-11) > 1e-25 {
		t.Fatalf("default G: got %v", engine.G)
	}
}

func TestStepEmptyEngine(t *testing.T) {
	engine := NewEngine(nil)
	engine.Step(1.0)
	if len(engine.Bodies) != 0 {
		t.Fatalf("empty engine gained bodies: %d", len(engine.Bodies))
	}
}

func TestMomentumConservation(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Sun", Mass: 1.989e30, Radius: 6.9634e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, -0.0971, 0}},
		{Name: "Earth", Mass: 5.972e24, Radius: 6.371e6, Position: Vector3D{1.496e11, 0, 0}, Velocity: Vector3D{0, 29780, 0}},
		{Name: "Mars", Mass: 6.39e23, Radius: 3.389e6, Position: Vector3D{2.279e11, 0, 0}, Velocity: Vector3D{0, 24077, 0}},
	})

	initial := totalMomentum(engine.Bodies)
	scale := math.Max(math.Abs(initial.X), math.Max(math.Abs(initial.Y), math.Abs(initial.Z)))
	if scale == 0 {
		scale = 1
	}

	for i := 0; i < 1000; i++ {
		engine.Step(3600)
	}

	final := totalMomentum(engine.Bodies)
	diff := final.Sub(initial)
	if math.Abs(diff.X)/scale > 1e-10 || math.Abs(diff.Y)/scale > 1e-10 || math.Abs(diff.Z)/scale > 1e-10 {
		t.Fatalf("momentum drift: initial %v final %v", initial, final)
	}
}

func TestEnergyConservationTwoBody(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "Sun", Mass: 1.989e30, Radius: 6.9634e8, Position: Vector3D{0, 0, 0}, Velocity: Vector3D{0, 0, 0}},
		{Name: "Earth", Mass: 5.972e24, Radius: 6.371e6, Position: Vector3D{1.496e11, 0, 0}, Velocity: Vector3D{0, 29780, 0}},
	})

	initial := totalEnergy(engine)
	for i := 0; i < 1000; i++ {
		engine.Step(3600)
	}
	final := totalEnergy(engine)

	relDrift := math.Abs((final - initial) / initial)
	if relDrift > 1e-6 {
		t.Fatalf("energy drift too large: %v (initial %v final %v)", relDrift, initial, final)
	}
}

func TestSymmetricBodiesStaySymmetric(t *testing.T) {
	engine := NewEngine([]SpaceBody{
		{Name: "A", Mass: 1e24, Radius: 1e6, Position: Vector3D{-1e10, 0, 0}, Velocity: Vector3D{0, 1000, 0}},
		{Name: "B", Mass: 1e24, Radius: 1e6, Position: Vector3D{1e10, 0, 0}, Velocity: Vector3D{0, -1000, 0}},
	})

	engine.Step(60)

	a := engine.Bodies[0]
	b := engine.Bodies[1]

	if math.Abs(a.Position.X+b.Position.X) > 1e-3 {
		t.Fatalf("X symmetry broken: %v + %v", a.Position.X, b.Position.X)
	}
	if math.Abs(a.Velocity.Y+b.Velocity.Y) > 1e-6 {
		t.Fatalf("Vy symmetry broken: %v + %v", a.Velocity.Y, b.Velocity.Y)
	}
}

func totalMomentum(bodies []SpaceBody) Vector3D {
	var p Vector3D
	for _, b := range bodies {
		p = p.Add(b.Velocity.Scale(b.Mass))
	}
	return p
}

func totalEnergy(engine *Engine) float64 {
	var ke float64
	for _, b := range engine.Bodies {
		ke += 0.5 * b.Mass * b.Velocity.LengthSquared()
	}

	var pe float64
	for i := 0; i < len(engine.Bodies); i++ {
		for j := i + 1; j < len(engine.Bodies); j++ {
			d := engine.Bodies[j].Position.Sub(engine.Bodies[i].Position)
			r := math.Sqrt(d.LengthSquared() + softeningSquared)
			pe -= engine.G * engine.Bodies[i].Mass * engine.Bodies[j].Mass / r
		}
	}

	return ke + pe
}
