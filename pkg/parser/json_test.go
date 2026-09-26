package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePreset(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "preset.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write preset: %v", err)
	}
	return path
}

const validPreset = `{
  "scenario_name": "Test",
  "description": "Test scenario",
  "time_step_seconds": 3600,
  "bodies": [
    {"name": "Sun", "mass": 1.989e30, "radius": 6.9634e8, "position": [0,0,0], "velocity": [0,0,0]},
    {"name": "Earth", "mass": 5.972e24, "radius": 6.371e6, "position": [1.496e11,0,0], "velocity": [0,29780,0]}
  ]
}`

func TestLoadPresetValid(t *testing.T) {
	name, dt, bodies, err := LoadPreset(writePreset(t, validPreset))
	if err != nil {
		t.Fatalf("LoadPreset: %v", err)
	}
	if name != "Test" {
		t.Fatalf("name: got %q", name)
	}
	if dt != 3600 {
		t.Fatalf("dt: got %v", dt)
	}
	if len(bodies) != 2 {
		t.Fatalf("bodies count: got %d want 2", len(bodies))
	}
	if bodies[1].Position.X != 1.496e11 {
		t.Fatalf("earth X: got %v", bodies[1].Position.X)
	}
	if bodies[1].Velocity.Y != 29780 {
		t.Fatalf("earth Vy: got %v", bodies[1].Velocity.Y)
	}
}

func TestLoadPresetMissingFile(t *testing.T) {
	if _, _, _, err := LoadPreset("does-not-exist.json"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadPresetRejectsNegativeTimeStep(t *testing.T) {
	bad := strings.Replace(validPreset, `"time_step_seconds": 3600`, `"time_step_seconds": -1`, 1)
	_, _, _, err := LoadPreset(writePreset(t, bad))
	if err == nil || !strings.Contains(err.Error(), "time_step_seconds") {
		t.Fatalf("expected time_step_seconds error, got %v", err)
	}
}

func TestLoadPresetRejectsEmptyBodies(t *testing.T) {
	bad := `{"scenario_name":"X","time_step_seconds":3600,"bodies":[]}`
	if _, _, _, err := LoadPreset(writePreset(t, bad)); err == nil {
		t.Fatal("expected error for empty bodies")
	}
}

func TestLoadPresetRejectsNegativeMass(t *testing.T) {
	bad := strings.Replace(validPreset, `"mass": 5.972e24`, `"mass": -1`, 1)
	if _, _, _, err := LoadPreset(writePreset(t, bad)); err == nil {
		t.Fatal("expected error for negative mass")
	}
}

func TestLoadPresetRejectsWrongPositionLength(t *testing.T) {
	bad := strings.Replace(validPreset, `"position": [1.496e11,0,0]`, `"position": [1.496e11,0]`, 1)
	if _, _, _, err := LoadPreset(writePreset(t, bad)); err == nil {
		t.Fatal("expected error for wrong position length")
	}
}

func TestLoadPresetRejectsDuplicateNames(t *testing.T) {
	bad := strings.Replace(validPreset, `"name": "Earth"`, `"name": "Sun"`, 1)
	if _, _, _, err := LoadPreset(writePreset(t, bad)); err == nil {
		t.Fatal("expected error for duplicate names")
	}
}

func TestLoadPresetRejectsUnknownFields(t *testing.T) {
	bad := `{"scenario_name":"X","time_step_seconds":3600,"extra":1,"bodies":[{"name":"A","mass":1,"radius":1,"position":[0,0,0],"velocity":[0,0,0]}]}`
	if _, _, _, err := LoadPreset(writePreset(t, bad)); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestLoadPresetRejectsNaNAndInf(t *testing.T) {
	bad := strings.Replace(validPreset, `"time_step_seconds": 3600`, `"time_step_seconds": 1e400`, 1)
	if _, _, _, err := LoadPreset(writePreset(t, bad)); err == nil {
		t.Fatal("expected error for Inf time step")
	}
}
