package bse

import (
	"math"
	"strings"
)

const (
	speedOfLight        = 299792458.0
	softeningSquared    = 1e6
	defaultSubstepDepth = 4
	defaultApproachCoef = 10.0
)

type Engine struct {
	G                   float64
	Bodies              []SpaceBody
	MaxSubstepDepth     int
	CloseApproachFactor float64
	EnableRelativity    bool
}

type MergeEvent struct {
	Absorber string
	Absorbed string
}

func NewEngine(bodies []SpaceBody) *Engine {
	return &Engine{
		G:                   6.67430e-11,
		Bodies:              bodies,
		MaxSubstepDepth:     defaultSubstepDepth,
		CloseApproachFactor: defaultApproachCoef,
		EnableRelativity:    false,
	}
}

func isBlackHole(name string) bool {
	return strings.Contains(strings.ToLower(name), "black hole")
}

func (e *Engine) computeAccelerations() []Vector3D {
	n := len(e.Bodies)
	accelerations := make([]Vector3D, n)

	cSquared := speedOfLight * speedOfLight

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			b1 := e.Bodies[i]
			b2 := e.Bodies[j]

			delta := b2.Position.Sub(b1.Position)
			rSq := delta.LengthSquared() + softeningSquared
			r := math.Sqrt(rSq)
			r3 := rSq * r

			fOverR3i := e.G * b2.Mass / r3
			fOverR3j := e.G * b1.Mass / r3

			accelerations[i] = accelerations[i].Add(delta.Scale(fOverR3i))
			accelerations[j] = accelerations[j].Sub(delta.Scale(fOverR3j))

			if !e.EnableRelativity {
				continue
			}

			relVel := b2.Velocity.Sub(b1.Velocity)
			totalMass := b1.Mass + b2.Mass
			vSq := relVel.LengthSquared()
			rDotV := delta.Dot(relVel)

			factor := e.G * totalMass / (cSquared * r3)
			coeffRadial := 4*e.G*totalMass/r - vSq

			grRel := delta.Scale(coeffRadial).Add(relVel.Scale(4 * rDotV)).Scale(factor)

			accelerations[i] = accelerations[i].Sub(grRel.Scale(b2.Mass / totalMass))
			accelerations[j] = accelerations[j].Add(grRel.Scale(b1.Mass / totalMass))
		}
	}
	return accelerations
}

func (e *Engine) needsRefinement() bool {
	n := len(e.Bodies)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			a := e.Bodies[i]
			b := e.Bodies[j]

			sumRadii := a.Radius + b.Radius
			if sumRadii <= 0 {
				continue
			}

			dist := b.Position.Sub(a.Position).Length()

			if dist <= sumRadii {
				continue
			}

			if dist < e.CloseApproachFactor*sumRadii {
				return true
			}
		}
	}
	return false
}

func (e *Engine) Step(dt float64) []MergeEvent {
	return e.adaptiveStep(dt, e.MaxSubstepDepth)
}

func (e *Engine) adaptiveStep(dt float64, depth int) []MergeEvent {
	if depth <= 0 || !e.needsRefinement() {
		return e.integrate(dt)
	}

	half := dt / 2

	first := e.adaptiveStep(half, depth-1)
	second := e.adaptiveStep(half, depth-1)

	return append(first, second...)
}

func (e *Engine) integrate(dt float64) []MergeEvent {
	n := len(e.Bodies)
	if n == 0 {
		return nil
	}

	accelCurrent := e.computeAccelerations()

	velHalf := make([]Vector3D, n)

	for i := 0; i < n; i++ {
		b := &e.Bodies[i]
		a := accelCurrent[i]

		b.Position = b.Position.Add(b.Velocity.Scale(dt)).Add(a.Scale(0.5 * dt * dt))
		velHalf[i] = b.Velocity.Add(a.Scale(0.5 * dt))
	}

	accelNext := e.computeAccelerations()

	for i := 0; i < n; i++ {
		b := &e.Bodies[i]
		aNext := accelNext[i]
		b.Velocity = velHalf[i].Add(aNext.Scale(0.5 * dt))
	}

	return e.resolveCollisions()
}

func (e *Engine) resolveCollisions() []MergeEvent {
	var events []MergeEvent

	for {
		merged := false
		n := len(e.Bodies)

		for i := 0; i < n && !merged; i++ {
			for j := i + 1; j < n && !merged; j++ {
				a := e.Bodies[i]
				b := e.Bodies[j]

				dist := b.Position.Sub(a.Position).Length()
				if dist > a.Radius+b.Radius {
					continue
				}

				absorber, absorbed := i, j
				if b.Mass > a.Mass {
					absorber, absorbed = j, i
				}

				events = append(events, MergeEvent{
					Absorber: e.Bodies[absorber].Name,
					Absorbed: e.Bodies[absorbed].Name,
				})

				e.mergeBodies(absorber, absorbed)
				merged = true
			}
		}

		if !merged {
			break
		}
	}

	return events
}

func (e *Engine) mergeBodies(absorber, absorbed int) {
	a := e.Bodies[absorber]
	b := e.Bodies[absorbed]

	totalMass := a.Mass + b.Mass

	merged := SpaceBody{
		Name: a.Name,
		Mass: totalMass,
		Position: a.Position.Scale(a.Mass).
			Add(b.Position.Scale(b.Mass)).
			Scale(1 / totalMass),
		Velocity: a.Velocity.Scale(a.Mass).
			Add(b.Velocity.Scale(b.Mass)).
			Scale(1 / totalMass),
	}

	if isBlackHole(a.Name) {
		merged.Radius = 2 * e.G * totalMass / (speedOfLight * speedOfLight)
	} else {
		merged.Radius = math.Cbrt(a.Radius*a.Radius*a.Radius + b.Radius*b.Radius*b.Radius)
	}

	e.Bodies[absorber] = merged

	newBodies := make([]SpaceBody, 0, len(e.Bodies)-1)
	for i, body := range e.Bodies {
		if i == absorbed {
			continue
		}
		newBodies = append(newBodies, body)
	}
	e.Bodies = newBodies
}
