# Copilot Instructions for midiwav

## Repository Overview

**midiwav** is a Go command-line application that converts MIDI files to WAV audio files. It watches a directory for new or updated MIDI files, converts them to WAV format using synthesized audio (square wave synthesis), and optionally posts both files to a Discord webhook.

### Key Details
- **Language:** Go 1.24.2+ (currently using Go 1.24.11)
- **Project Size:** Small (~493 lines of Go code across 6 source files)
- **Dependencies:** Single external dependency: `gitlab.com/gomidi/midi/v2 v2.3.18`
- **Project Type:** Single-package CLI application with main package
- **Build System:** Standard Go tooling (`go build`, `go test`)
- **Release System:** GoReleaser for creating multi-platform binaries

## Build and Test Instructions

### Prerequisites
- Go 1.24.2 or higher (specified in `go.mod`)
- No system dependencies or CGO required (CGO_ENABLED=0 in builds)

### Critical Build Steps

**ALWAYS run commands in this exact order:**

1. **Install dependencies** (run this first, always):
   ```bash
   go mod download
   ```

2. **Build the application**:
   ```bash
   go build -o midiwav .
   ```
   - Build time: ~1-2 seconds
   - Output: `midiwav` binary in current directory (ignored by .gitignore)
   - No CGO or system libraries required

3. **Run tests**:
   ```bash
   go test -v ./...
   ```
   - Test duration: ~5-6 seconds (tests generate and compare WAV files)
   - Tests use golden files in `testdata/` directory
   - All tests must pass before submitting PR

4. **Update golden test files** (only when intentionally changing WAV output):
   ```bash
   go test -update
   ```
   - Use this flag ONLY when you intentionally change the MIDI-to-WAV conversion logic
   - This regenerates the expected WAV files in `testdata/`
   - Verify the changes are intentional before committing updated WAV files

### Code Quality Checks (CI Requirements)

The CI pipeline runs these checks. **ALWAYS run these before committing:**

1. **Format check** (must pass):
   ```bash
   gofmt -l .
   ```
   - Output should be empty (no unformatted files)
   - To fix formatting: `gofmt -w .`

2. **go.mod tidy check** (must pass):
   ```bash
   go mod tidy -diff
   ```
   - Output should be empty (go.mod and go.sum are clean)
   - To fix: `go mod tidy`

3. **Optional: Run go vet**:
   ```bash
   go vet ./...
   ```
   - Checks for common Go mistakes

### Running the Application

The application requires environment variables:

```bash
export MIDIWAV_DIR=/path/to/midi/files
export MIDIWAV_DISCORD_WEBHOOK=https://discord.com/api/webhooks/...  # optional

./midiwav
```

- `MIDIWAV_DIR`: Directory to watch for MIDI files (required, or it will error)
- `MIDIWAV_DISCORD_WEBHOOK`: Discord webhook URL for posting files (optional)
- The app runs in watch mode by default, checking every 10 seconds for new/updated MIDI files

## Project Layout and Architecture

### File Structure (Root Directory)
```
/home/runner/work/midiwav/midiwav/
├── .github/
│   └── workflows/
│       ├── ci.yml          # CI: tests + formatting + go mod tidy
│       └── gorelease.yml   # Release workflow (on version tags)
├── testdata/               # Test fixtures: MIDI and golden WAV files
│   ├── *.mid              # Test MIDI input files
│   └── *.wav              # Expected WAV output (golden files)
├── .gitignore             # Ignores: binaries, test outputs, .env, samples/
├── .goreleaser.yaml       # GoReleaser config for multi-platform builds
├── go.mod / go.sum        # Go module files
├── LICENSE                # License file
├── README.md              # Basic setup and test instructions
└── Source files (all in main package):
    ├── main.go            # Entry point, watch loop, file processing
    ├── midi_player.go     # MIDI parsing and WAV synthesis (core logic)
    ├── midi_player_test.go # Tests for MIDI-to-WAV conversion
    ├── wav_writer.go      # WAV file format writer
    ├── post_discord.go    # Discord webhook posting
    └── scan_directory.go  # Directory scanning for MIDI files
```

### Architectural Components

1. **main.go**: Application entry point
   - `executer` struct orchestrates the watch loop
   - Calls `findUnprocessedMIDIFiles()` to scan directory
   - For each MIDI file, calls `midiToWAVE()` then optionally `postToDiscord()`
   - Watch mode runs every 10 seconds indefinitely

2. **midi_player.go**: Core MIDI-to-WAV conversion (125 lines)
   - `midiToWAVE()`: Main conversion function
   - `smfToPCMArray()`: Parses MIDI events and synthesizes audio
   - Uses **square wave synthesis** for all notes
   - Sample rate: 22050 Hz (44100/2), mono, 16-bit PCM
   - Only supports MetricTicks time format
   - TODO comments indicate future enhancements (multiple tracks, different waveforms)

3. **wav_writer.go**: WAV file format generation (96 lines)
   - `writeWAVE()`: Writes proper WAV header and PCM data
   - Constants define audio format: mono, 22050Hz, 16-bit

4. **scan_directory.go**: File discovery (32 lines)
   - `findUnprocessedMIDIFiles()`: Finds MIDI files that need processing
   - A MIDI file is "unprocessed" if:
     - Corresponding .wav file doesn't exist, OR
     - MIDI file is newer than WAV file

5. **post_discord.go**: Discord integration (75 lines)
   - `postToDiscord()`: Posts files via multipart form upload
   - Optional feature, silently skipped if webhook not configured

### Testing Strategy
- Golden file testing: compare generated WAV against known-good outputs
- Test files in `testdata/`: issue0001_ok1, issue0001_ok2, issue0001_ng1, issue0001_ng2
- Tests verify exact byte-for-byte WAV output match
- Use `-update` flag to regenerate golden files when conversion logic changes

## CI/CD Pipeline Details

### GitHub Actions Workflows

1. **ci.yml** - Continuous Integration (runs on push/PR to main)
   - Triggered on: push/PR to `main` branch
   - Steps:
     1. Checkout code
     2. Setup Go (version from go.mod, with cache)
     3. Run tests: `go test -v ./...`
     4. Check formatting: `gofmt -l .` (fails if any files unformatted)
     5. Check go.mod: `go mod tidy -diff` (fails if not tidy)
   - **To replicate CI locally:** Run all three commands above in order

2. **gorelease.yml** - Release (runs on version tags like v1.0.0)
   - Triggered on: tags matching `v*.*.*`
   - Uses GoReleaser to build multi-platform binaries
   - Builds for: darwin/amd64, darwin/arm64, windows/amd64
   - Creates tar.gz (Unix) and zip (Windows) archives
   - Publishes to GitHub Releases as prerelease

### Pre-commit Checklist
Before committing, ALWAYS run:
```bash
go test -v ./...      # Must pass
gofmt -l .            # Must output nothing
go mod tidy -diff     # Must output nothing
```

## Key Conventions and Patterns

### Code Style
- Standard Go formatting (use `gofmt`)
- No linting configuration beyond `gofmt` and `go vet`
- Main package (no internal packages)
- Simple, straightforward code structure

### Testing Conventions
- Golden file testing for WAV output verification
- Test files use descriptive names: `issue0001_ok1`, `issue0001_ng1`
- Always test against golden files in `testdata/`
- Tests print MIDI parsing info to stdout (MetricTicks, BPM)

### Error Handling
- Most errors are fatal: `log.Fatal(err)`
- Discord posting errors are logged but non-fatal: `log.Println(err)`

### TODOs in Codebase
- Multiple track support (`midi_player.go:36`)
- Configurable waveforms beyond square wave (`midi_player.go:66`)
- Configurable watch mode (`main.go:21`)

### Important Constants
- `sampleRate = 44100 / 2` (22050 Hz) in `wav_writer.go`
- Watch interval: 10 seconds in `main.go:36`

## Common Pitfalls and Workarounds

### Test Timing
- Tests take ~5-6 seconds to complete (WAV generation is computationally intensive)
- Don't assume tests are hanging if they take a few seconds

### WAV File Changes
- Golden WAV files are binary and should only be updated with `-update` flag
- Unintended WAV changes will cause test failures
- If tests fail after code changes, verify whether WAV output differences are intentional

### Dependencies
- Only one external dependency: `gitlab.com/gomidi/midi/v2`
- First test run downloads dependencies automatically
- No special setup needed

### Build Output
- Built binary `midiwav` is in .gitignore - don't commit it
- `samples/` directory in .gitignore for test output

## Trust These Instructions

These instructions have been validated by running all commands successfully. Trust this information and only explore further if:
- Instructions are incomplete for your specific task
- You encounter errors not documented here
- You need to understand implementation details beyond what's described

When in doubt, refer to:
1. `README.md` for basic usage
2. `.github/workflows/ci.yml` for CI requirements  
3. `go.mod` for Go version and dependencies
