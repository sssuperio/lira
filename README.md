# lira — a sound sketchbook

> **Lira is not a DAW. It is a sound sketchbook.**
>
> Lira turns simple variables into tiny exportable sounds.

Lira lets you generate tiny sounds, jingles, sonic logos, UI sounds, and "sound illustrations" without knowing music theory.

## What Lira is

- A fast sound sketchbook
- A parameter-based sound generator
- A tiny composition playground
- An export-first tool
- A way to make "graphics that sound"

## What Lira is not

- A DAW (Digital Audio Workstation)
- Ableton, GarageBand, or a live coding environment
- A professional music production tool
- A piano roll or tracker
- A mixer-heavy audio workstation

## Quick start

```bash
# Clone and run
git clone https://github.com/sssuperio/lira.git
cd lira
task run

# Open http://localhost:8090
```

## How it works

1. Choose a sound sketch type
2. Set a few human-friendly variables
3. Generate sound
4. Listen
5. Make small adjustments
6. Export

That's it. No notes, no chords, no scales, no audio engineering.

## Parameter language

Lira uses approachable language:

- **dot** instead of note
- **pulse** instead of beat
- **gesture** instead of melody
- **cluster** instead of chord
- **scene** instead of composition
- **material** instead of instrument
- **sketch** instead of song

## Parameters

| Parameter  | Values                                                           |
| ---------- | ---------------------------------------------------------------- |
| Mood       | calm, happy, curious, tense, magic, mechanical, aquatic, warm    |
| Material   | glass, wood, bell, bubble, pluck, breath, metal, soft-noise, toy |
| Shape      | rise, fall, bounce, pulse, wave, sparkle, swell, pop, orbit      |
| Density    | 1–10                                                             |
| Brightness | 1–10                                                             |
| Softness   | 1–10                                                             |
| Movement   | 1–10                                                             |
| Duration   | 0.2s–10s                                                         |
| Seed       | any string                                                       |

## Architecture

Lira follows the same architecture as [Chirone](https://github.com/sssuperio/chirone):

- **Go backend** — single binary serving an embedded Svelte frontend
- **Svelte frontend** — static site with Web Audio API for sound generation
- **Local project files** — JSON-based persistence under `data/`
- **Export** — WAV, JSON, and ZIP downloads from the browser

```
lira/
├── main.go              # Go server + embedded UI
├── src/                 # SvelteKit frontend
│   ├── lib/
│   │   ├── audio/       # Sound engine + export
│   │   └── types.ts     # Lira data types
│   └── routes/          # Pages
├── data/                # Local project data (gitignored)
├── Taskfile.yml
├── Dockerfile
└── .goreleaser.yaml
```

## Project file layout

```
data/
  default/
    project.json                          # Project metadata
    sketches/
      tiny-victory.json                   # Sketch definition
    variants/
      tiny-victory/
        variant-001.json                  # Variant metadata
    exports/
      tiny-victory/
        tiny-victory-variant-001.wav      # Exported audio
        tiny-victory-variant-001.json     # Exported sketch JSON
        tiny-victory-variant-001.zip      # Full ZIP export
```

## CLI

```bash
lira                    # Start server
lira serve --addr :8090 --data-dir ./data
lira version            # Print version
lira export --sketch tiny-victory --format json  # Export JSON (stub)
```

Note: Audio export (.wav, .zip) happens in the browser UI for MVP. The CLI export command writes JSON metadata.

## Build

```bash
# Install frontend dependencies
task install

# Build everything (web + Go binary)
task build

# The binary is at ./bin/lira
./bin/lira
```

## Docker

```bash
# Build and run
task docker:run

# Or manually
docker build -t lira .
docker run -p 8080:8080 -v lira-data:/data lira

# Coolify uses docker-compose.yml without host port mappings.
# Assign a domain to the lira service on container port 8080.
docker compose up --build
```

## Export formats

- **WAV** — 16-bit PCM audio file
- **JSON** — sketch definition (lira.sketch.v1 schema)
- **ZIP** — bundle with sketch.json, audio.wav, README.md

## Determinism

Given the same sketch parameters and seed, Lira generates the same sound every time.

Uses a seeded PRNG (mulberry32) for deterministic randomness.

## Current limitations

- Audio generation is browser-side only (Web Audio API)
- No server-side audio rendering (CLI export is metadata-only)
- No MIDI support
- No VST/plugin support
- No multi-track timeline
- No cloud sync
- No user accounts

## Development

```bash
# Run Go tests
task test

# Frontend dev server (with hot reload)
task dev:web

# Go server only (for API development)
task dev:server
```

## Related

- [Chirone](https://github.com/sssuperio/chirone) — a font design playground (spiritual predecessor)

## License

MIT
