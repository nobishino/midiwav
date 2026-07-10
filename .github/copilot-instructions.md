# Copilot Instructions for midiwav

## Repository Overview

**midiwav** is a Go command-line application that converts MIDI files to WAV audio files. It watches directories for new or updated MIDI files, converts them to WAV format using synthesized audio (square wave synthesis), runs a four-part harmony rule check (芸大和声ベース) on MIDI files that look like SATB chorales, and optionally posts the files and the harmony report to a Discord webhook.

### Key Details
- **Language:** Go 1.24.2+ (see `go.mod`)
- **Dependencies:** `gitlab.com/gomidi/midi/v2` (MIDI parsing) and `github.com/BurntSushi/toml` (config)
- **Project Type:** CLI application (`main` package at repo root) with two library packages: `harmony` and `synth`
- **Build System:** Standard Go tooling (`go build`, `go test`)
- **Release System:** GoReleaser for creating multi-platform binaries (builds from the repo root)

## Build and Test Instructions

### Critical Build Steps

1. **Install dependencies** (run this first, always):
   ```bash
   go mod download
   ```

2. **Build the application**:
   ```bash
   go build -o midiwav .
   ```
   - Output: `midiwav` binary in current directory (ignored by .gitignore)
   - No CGO or system libraries required

3. **Run tests**:
   ```bash
   go test -v ./...
   ```
   - Test duration: ~5-10 seconds (tests generate and compare WAV files)
   - Golden files: `synth/testdata/*.wav` (WAV output) and `harmony/testdata/*.golden.txt` (harmony reports)
   - All tests must pass before submitting PR

4. **Update golden test files** (only when intentionally changing output):
   ```bash
   go test ./... -update
   ```
   - Use this flag ONLY when you intentionally change the MIDI-to-WAV conversion or the harmony report output
   - Verify the changes are intentional before committing updated golden files

### Code Quality Checks (CI Requirements)

The CI pipeline runs these checks. **ALWAYS run these before committing:**

1. **Format check** (must pass):
   ```bash
   gofmt -l .
   ```
   - Output should be empty; to fix: `gofmt -w .`

2. **go.mod tidy check** (must pass):
   ```bash
   go mod tidy -diff
   ```

3. **Optional: Run go vet**:
   ```bash
   go vet ./...
   ```

### Running the Application

The application takes a TOML configuration file (see `config.toml.example` and `README.md`):

```bash
./midiwav -config /path/to/config.toml
```

- Each `[[target]]` entry has `dir` (required), `recursive`, and `discord_webhook_url` (optional)
- The app runs in watch mode, checking every 10 seconds for new/updated MIDI files

There is also a side-effect-free subcommand that runs only the harmony analysis
(no WAV/score/Discord output) and exits with 0 (clean), 1 (violations), or 2 (error):

```bash
./midiwav check [-key <german-key>] [-format text|json] file.mid...
```

## Project Layout and Architecture

### Package Structure
```
.
├── .github/workflows/      # ci.yml (tests+fmt+tidy), gorelease.yml, tagpr.yml
├── main.go                 # CLI entry point, watch loop, per-file orchestration
├── check.go                # `check` subcommand: analysis only, text/json output
├── config.go               # TOML config loading ([[target]] entries)
├── scan_directory.go       # Finds MIDI files whose WAV is missing or stale
├── post_discord.go         # Discord webhook posting (multipart upload)
├── harmony/                # 4声体和声の分析・添削 (library package)
│   ├── harmony_check.go    #   Chorale extraction, rule checks, Analyze/Format API
│   ├── harmony_chord.go    #   Chord symbols (和音記号) and chord names
│   └── testdata/           #   *.mid inputs + *.golden.txt expected reports
└── synth/                  # MIDIからのWAV合成 (library package)
    ├── midi_player.go      #   WriteWAV, tempo map, square wave synthesis
    ├── wav_writer.go       #   WAV header/PCM writer (22050 Hz, mono, 16-bit)
    └── testdata/           #   *.mid inputs + *.wav golden outputs
```

### Architectural Components

1. **main.go**: parses the SMF once per MIDI file, then passes the parsed
   `*smf.SMF` to both `synth.WriteWAV` and `harmony.Analyze`. The harmony key
   is read from the filename via `harmony.ParseKeyFromFilename` (e.g. `es-moll.mid`).

2. **harmony package**: `Analyze(s *smf.SMF, key *Key) (*Report, bool)` returns a
   structured report (chords with symbols/names, issues with warn/info level);
   `(*Report).Format()` renders the Japanese text report. Analysis only runs on
   MIDI files with exactly 4 note tracks (SATB, top to bottom).

3. **synth package**: `WriteWAV(w io.Writer, s *smf.SMF) error` synthesizes square
   wave audio (tempo map built from all tracks, default 120 BPM) and writes a WAV.
   Only supports MetricTicks time format.

### Testing Strategy
- Golden file testing in both library packages; use `-update` to regenerate
- Harmony bug reproduction: drop a `.mid` file into `harmony/testdata/`, run
  `go test ./... -update`, review/correct the generated `.golden.txt` (see README)
- Tests verify exact byte-for-byte / string match

## Pre-commit Checklist
```bash
go test -v ./...      # Must pass
gofmt -l .            # Must output nothing
go mod tidy -diff     # Must output nothing
```

## Key Conventions and Patterns

- Standard Go formatting (`gofmt`); no linting beyond `gofmt` and `go vet`
- Commit / PR titles in Japanese, matching `git log` style
- Errors in the watch loop are logged and skipped (one bad file must not stop the loop); Discord posting errors are logged but non-fatal
- Important constants: `sampleRate = 44100 / 2` (22050 Hz) in `synth/wav_writer.go`; watch interval 10 seconds in `main.go`

## Common Pitfalls

- Golden WAV files are binary; only update them via `go test ./... -update` and verify the change is intentional
- Built binary `midiwav` and `samples/` are in .gitignore — don't commit them

## Trust These Instructions

Trust this information and only explore further if instructions are incomplete
for your task or you encounter undocumented errors. When in doubt, refer to
`README.md`, `.github/workflows/ci.yml`, and `go.mod`.
