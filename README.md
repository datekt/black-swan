<p align="center">
  <img src="assets/black-swan-event_banner.png" alt="Black Swan Event (BSE) Banner" width="100%">
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPLv3-blue.svg" alt="License: GPL v3"></a>
  <img src="https://img.shields.io/badge/Go-1.27.1-00ADD8.svg" alt="Go Version">
  <img src="https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg" alt="Platform">
  <img src="https://img.shields.io/badge/Dependencies-zero-success.svg" alt="Zero dependencies">
</p>

---

**Black Swan Event (BSE)** is a deterministic N-body orbital mechanics simulator written from scratch in pure Go. No external physics libraries, no API calls. Just Newton's law of universal gravitation, a Velocity Verlet integrator, and local JSON scenarios.

The engine can model both stable systems and "black swan" events — sudden catastrophes such as a massive black hole passing through the Solar System, with correct body absorption, momentum conservation, and optional post-Newtonian corrections.

## ✨ Features

- **Zero dependencies.** Standard library only.
- **Deterministic.** Identical initial conditions produce bit-for-bit identical results.
- **Symplectic Velocity Verlet.** Conserves energy across millions of timesteps.
- **Adaptive stepping.** `dt` is subdivided on close approach to preserve accuracy.
- **Body merging.** Conserves mass and momentum; black holes receive a Schwarzschild radius.
- **Post-Newtonian corrections.** First-order Einstein–Infeld–Hoffmann term, enabled with `--relativity`.
- **CLI and visualizer.** Batch computation and a local web viewer in one repository.

## 🛠️ Architecture

```text
black-swan/
├── .github/             # CI, issue and PR templates
├── assets/              # Banner and media
├── presets/             # JSON scenarios
├── pkg/
│   ├── bse/             # Vectors, bodies, engine, relativity
│   └── parser/          # Preset loading and validation
└── apps/
    ├── cli/             # CLI calculator
    └── webview/         # Local web visualizer
```

## 🧠 Physics

For every pair of bodies the engine computes gravitational acceleration:

$$\vec{a}_i = G \sum_{j \neq i} \frac{m_j (\vec{x}_j - \vec{x}_i)}{|\vec{x}_j - \vec{x}_i|^3}$$

Integration is performed with the Velocity Verlet scheme:

1. $\vec{x}(t+\Delta t) = \vec{x}(t) + \vec{v}(t)\Delta t + \tfrac{1}{2}\vec{a}(t)\Delta t^2$
2. $\vec{v}(t+\tfrac{1}{2}\Delta t) = \vec{v}(t) + \tfrac{1}{2}\vec{a}(t)\Delta t$
3. Recompute $\vec{a}(t+\Delta t)$ using the updated positions.
4. $\vec{v}(t+\Delta t) = \vec{v}(t+\tfrac{1}{2}\Delta t) + \tfrac{1}{2}\vec{a}(t+\Delta t)\Delta t$

A softening factor $\varepsilon^2 = 10^6\,\text{m}^2$ is added to the denominator to avoid division by zero.

### Post-Newtonian corrections

The `--relativity` flag adds a first-order term:

$$\vec{a}_{i,\text{GR}} = \frac{G M}{c^2 r^3} \left[ \left( \frac{4GM}{r} - v^2 \right) \vec{r} + 4(\vec{r} \cdot \vec{v}) \vec{v} \right]$$

where $M = m_i + m_j$, and $\vec{r}$, $\vec{v}$ are relative position and velocity. For Mercury this produces the classic 43 arcseconds per century.

## 🚀 Quick Start

```bash
git clone https://github.com/datekt/black-swan.git
cd black-swan
```

### CLI

```bash
go run ./apps/cli --preset presets/solar_system.json --days 365
go run ./apps/cli --preset presets/rogue_black_hole.json --days 180
go run ./apps/cli --preset presets/solar_system.json --days 36500 --relativity
```

Output is written to `output/trajectory.json`.

### Visualizer

In one terminal:

```bash
go run ./apps/cli --preset presets/rogue_black_hole.json --days 180
```

In another:

```bash
go run ./apps/webview
```

Open `http://localhost:8080`. Trajectory data is loaded from `output/` automatically — no file copying required.

### Makefile

```bash
make build      # Build bse-cli and bse-web
make run-cli    # Run the CLI on solar_system
make run-web    # Start the web server
make test       # Run all tests with race detector
make fmt vet    # Formatting and static analysis
make clean      # Remove build artifacts
```

On Windows without GNU Make, use the `go run` commands directly.

## 📦 Presets

| File | What it models |
|------|----------------|
| `presets/solar_system.json` | Stable inner Solar System, 3 bodies, zero net momentum |
| `presets/rogue_black_hole.json` | A 10 M☉ black hole enters the system and absorbs the Sun |

Custom scenarios are plain JSON. The schema is defined in `pkg/parser/json.go` and validated on load: empty names, negative masses, `NaN`, duplicates, and unknown fields are rejected.

## 🗺️ Roadmap

- [x] Velocity Verlet + softening
- [x] Adaptive substepping
- [x] Body merging with momentum conservation
- [x] Schwarzschild radius for black holes
- [x] Post-Newtonian corrections
- [ ] Barnes–Hut for $O(n \log n)$ on thousands of bodies
- [ ] CSV trajectory export
- [ ] Asteroid belt presets
- [ ] Event timeline in the web visualizer

## 🤝 Contributing

PRs and issues are welcome. Before opening a PR, run `make fmt vet test`. Issue templates live in `.github/ISSUE_TEMPLATE/`, the PR template is at `.github/PULL_REQUEST_TEMPLATE.md`.

## 🛡️ License

GNU GPL v3. Any derivative work that includes the BSE core must be distributed under the same license. See [LICENSE](LICENSE) for the full text.