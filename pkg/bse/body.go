package bse

type SpaceBody struct {
	Name     string   `json:"name"`
	Mass     float64  `json:"mass"`
	Radius   float64  `json:"radius"`
	Position Vector3D `json:"position"`
	Velocity Vector3D `json:"velocity"`
}

func (b *SpaceBody) Clone() SpaceBody {
	return SpaceBody{
		Name:     b.Name,
		Mass:     b.Mass,
		Radius:   b.Radius,
		Position: b.Position,
		Velocity: b.Velocity,
	}
}
