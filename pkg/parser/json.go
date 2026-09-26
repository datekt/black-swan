package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/datekt/black-swan/pkg/bse"
)

type JSONPreset struct {
	ScenarioName    string     `json:"scenario_name"`
	Description     string     `json:"description"`
	TimeStepSeconds float64    `json:"time_step_seconds"`
	JSONBodies      []JSONBody `json:"bodies"`
}

type JSONBody struct {
	Name     string    `json:"name"`
	Mass     float64   `json:"mass"`
	Radius   float64   `json:"radius"`
	Position []float64 `json:"position"`
	Velocity []float64 `json:"velocity"`
}

func LoadPreset(filePath string) (string, float64, []bse.SpaceBody, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, nil, fmt.Errorf("failed to open preset file: %w", err)
	}
	defer file.Close()

	var preset JSONPreset
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&preset); err != nil {
		return "", 0, nil, fmt.Errorf("failed to decode preset JSON: %w", err)
	}

	if preset.ScenarioName == "" {
		return "", 0, nil, fmt.Errorf("scenario_name is required")
	}

	if preset.TimeStepSeconds <= 0 || math.IsNaN(preset.TimeStepSeconds) || math.IsInf(preset.TimeStepSeconds, 0) {
		return "", 0, nil, fmt.Errorf("time_step_seconds must be a positive finite number, got %v", preset.TimeStepSeconds)
	}

	if len(preset.JSONBodies) == 0 {
		return "", 0, nil, fmt.Errorf("preset must contain at least one body")
	}

	seenNames := make(map[string]struct{}, len(preset.JSONBodies))
	bseBodies := make([]bse.SpaceBody, len(preset.JSONBodies))

	for i, jb := range preset.JSONBodies {
		if jb.Name == "" {
			return "", 0, nil, fmt.Errorf("body at index %d has empty name", i)
		}
		if _, exists := seenNames[jb.Name]; exists {
			return "", 0, nil, fmt.Errorf("duplicate body name: %s", jb.Name)
		}
		seenNames[jb.Name] = struct{}{}

		if jb.Mass <= 0 || math.IsNaN(jb.Mass) || math.IsInf(jb.Mass, 0) {
			return "", 0, nil, fmt.Errorf("body '%s' must have a positive finite mass", jb.Name)
		}
		if jb.Radius < 0 || math.IsNaN(jb.Radius) || math.IsInf(jb.Radius, 0) {
			return "", 0, nil, fmt.Errorf("body '%s' must have a non-negative finite radius", jb.Name)
		}

		if len(jb.Position) != 3 {
			return "", 0, nil, fmt.Errorf("body '%s' position must have exactly 3 coordinates", jb.Name)
		}
		if len(jb.Velocity) != 3 {
			return "", 0, nil, fmt.Errorf("body '%s' velocity must have exactly 3 coordinates", jb.Name)
		}

		for axis, coord := range jb.Position {
			if math.IsNaN(coord) || math.IsInf(coord, 0) {
				return "", 0, nil, fmt.Errorf("body '%s' position component %d is not finite", jb.Name, axis)
			}
		}
		for axis, coord := range jb.Velocity {
			if math.IsNaN(coord) || math.IsInf(coord, 0) {
				return "", 0, nil, fmt.Errorf("body '%s' velocity component %d is not finite", jb.Name, axis)
			}
		}

		bseBodies[i] = bse.SpaceBody{
			Name:   jb.Name,
			Mass:   jb.Mass,
			Radius: jb.Radius,
			Position: bse.Vector3D{
				X: jb.Position[0],
				Y: jb.Position[1],
				Z: jb.Position[2],
			},
			Velocity: bse.Vector3D{
				X: jb.Velocity[0],
				Y: jb.Velocity[1],
				Z: jb.Velocity[2],
			},
		}
	}

	return preset.ScenarioName, preset.TimeStepSeconds, bseBodies, nil
}
