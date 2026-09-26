package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/datekt/black-swan/pkg/bse"
	"github.com/datekt/black-swan/pkg/parser"
)

const engineVersion = "1.27.1"

type TrajectorySnapshot struct {
	Step       int             `json:"step"`
	SimTimeSec float64         `json:"sim_time_seconds"`
	Bodies     []bse.SpaceBody `json:"bodies"`
	Events     []Event         `json:"events,omitempty"`
}

type Event struct {
	Absorber string `json:"absorber"`
	Absorbed string `json:"absorbed"`
}

func main() {
	presetPath := flag.String("preset", "presets/solar_system.json", "Path to the cosmic scenario JSON file")
	durationDays := flag.Int("days", 365, "Simulation duration in Earth days")
	outputDir := flag.String("out", "output", "Directory to save trajectory simulation data")
	relativity := flag.Bool("relativity", false, "Enable post-Newtonian relativistic corrections")
	flag.Parse()

	if *durationDays < 0 {
		fmt.Println("[❌] Simulation duration must be non-negative")
		os.Exit(1)
	}

	fmt.Printf("🦅 BLACK SWAN EVENT (BSE) ENGINE v%s\n", engineVersion)
	fmt.Println("--------------------------------------------------")

	scenarioName, dt, bodies, err := parser.LoadPreset(*presetPath)
	if err != nil {
		fmt.Printf("[❌] Failed to load preset: %v\n", err)
		os.Exit(1)
	}

	if dt <= 0 {
		fmt.Printf("[❌] Invalid time step in preset: %v\n", dt)
		os.Exit(1)
	}

	if len(bodies) == 0 {
		fmt.Println("[❌] Preset contains no bodies")
		os.Exit(1)
	}

	fmt.Printf("[⚙️] Loaded Scenario : %s\n", scenarioName)
	fmt.Printf("[⏳] Time Step (dt)  : %.1f seconds\n", dt)
	fmt.Printf("[🪐] Active Bodies   : %d\n", len(bodies))
	fmt.Printf("[🧮] Relativity      : %v\n", *relativity)

	engine := bse.NewEngine(bodies)
	engine.EnableRelativity = *relativity

	totalDurationSec := float64(*durationDays) * 24.0 * 60.0 * 60.0
	totalSteps := int(math.Ceil(totalDurationSec / dt))

	fmt.Printf("[🚀] Simulating %d orbital steps for %d days...\n", totalSteps, *durationDays)

	history := make([]TrajectorySnapshot, 0, totalSteps/10+1)
	pendingEvents := make([]Event, 0)

	initialBodies := make([]bse.SpaceBody, len(engine.Bodies))
	for i, b := range engine.Bodies {
		initialBodies[i] = b.Clone()
	}
	history = append(history, TrajectorySnapshot{
		Step:       0,
		SimTimeSec: 0,
		Bodies:     initialBodies,
	})

	simStart := time.Now()

	for step := 1; step <= totalSteps; step++ {
		merges := engine.Step(dt)

		for _, m := range merges {
			fmt.Printf("[💥] Step %d: %s absorbed %s\n", step, m.Absorber, m.Absorbed)
			pendingEvents = append(pendingEvents, Event{
				Absorber: m.Absorber,
				Absorbed: m.Absorbed,
			})
		}

		if step%10 == 0 || step == totalSteps || len(merges) > 0 {
			currentBodies := make([]bse.SpaceBody, len(engine.Bodies))
			for i, b := range engine.Bodies {
				currentBodies[i] = b.Clone()
			}

			snapshot := TrajectorySnapshot{
				Step:       step,
				SimTimeSec: float64(step) * dt,
				Bodies:     currentBodies,
			}

			if len(pendingEvents) > 0 {
				snapshot.Events = append([]Event(nil), pendingEvents...)
				pendingEvents = pendingEvents[:0]
			}

			history = append(history, snapshot)
		}
	}
	simDuration := time.Since(simStart)

	fmt.Println("[💾] Serializing orbital telemetry...")
	jsonData, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		fmt.Printf("[❌] Error marshaling simulation dataset: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("[❌] Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	outputFile := filepath.Join(*outputDir, "trajectory.json")
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		fmt.Printf("[❌] Failed to write trajectory data to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[✅] Calculation successfully finished!")
	fmt.Printf("[📊] Data exported to: %s\n", outputFile)
	fmt.Printf("[⏱️] Pure physics computation took: %v\n", simDuration)
	fmt.Println("--------------------------------------------------")
}
