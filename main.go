package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"gitlab.com/gomidi/midi/v2/smf"

	"github.com/nobishino/midiwav/harmony"
	"github.com/nobishino/midiwav/synth"
)

func main() {
	configPath := flag.String("config", "", "Path to TOML configuration file")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("Please specify a configuration file using -config flag")
	}

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	exec := &executer{
		targets:   config.Targets,
		watchMode: true, // TODO: make configurable
	}
	if err := exec.start(); err != nil {
		log.Fatal(err)
	}
}

type executer struct {
	targets   []Target
	watchMode bool
}

func (e *executer) start() error {
	if e.watchMode {
		tick := time.NewTicker(10 * time.Second)
		defer tick.Stop()
		for {
			if err := e.execute(); err != nil {
				return err
			}
			<-tick.C
		}
	} else {
		return e.execute()
	}
}

func (e *executer) execute() error {
	for _, target := range e.targets {
		if err := e.processTarget(target); err != nil {
			log.Printf("Error processing target %s: %v", target.Dir, err)
			continue
		}
	}
	return nil
}

func (e *executer) processTarget(target Target) error {
	unprocessed, err := findUnprocessedMIDIFiles(target.Dir, target.Recursive)
	if err != nil {
		return err
	}
	log.Printf("Found %d unprocessed MIDI files in %s.\n", len(unprocessed), target.Dir)

	for _, srcPath := range unprocessed {
		if err := processFile(srcPath, target.DiscordWebhookURL); err != nil {
			log.Printf("Error processing %s: %v", srcPath, err)
			continue
		}
	}
	return nil
}

func processFile(srcPath string, discordWebhookURL string) error {
	fmt.Println("Processing:", srcPath)
	srcMIDI, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open MIDI file: %w", err)
	}
	defer srcMIDI.Close()

	smfData, err := smf.ReadFrom(srcMIDI)
	if err != nil {
		return fmt.Errorf("failed to read MIDI file: %w", err)
	}

	dstWAV, err := os.Create(strings.TrimSuffix(srcPath, ".mid") + ".wav")
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer dstWAV.Close()

	if err := synth.WriteWAV(dstWAV, smfData); err != nil {
		return fmt.Errorf("failed to convert MIDI to WAV: %w", err)
	}

	report := checkHarmony(smfData, srcPath)
	if report != "" {
		fmt.Print(report)
	}

	if discordWebhookURL != "" {
		content, files := buildDiscordPost(srcMIDI, dstWAV, report)
		if err := postToDiscord(discordWebhookURL, content, files...); err != nil {
			log.Println(err)
		}
	}
	return nil
}

// checkHarmony runs the harmony rule check if the MIDI looks like a
// four-part chorale. It returns an empty report otherwise. The key is
// taken from the filename (e.g. es-moll.mid).
func checkHarmony(smfData *smf.SMF, srcPath string) string {
	var key *harmony.Key
	if k, ok := harmony.ParseKeyFromFilename(srcPath); ok {
		key = &k
	}
	report, ok := harmony.Analyze(smfData, key)
	if !ok {
		return ""
	}
	return report.Format()
}

// discordContentLimit is Discord's maximum message content length in characters.
const discordContentLimit = 2000

func buildDiscordPost(srcMIDI, dstWAV *os.File, report string) (string, []discordFile) {
	content := "MIDIファイルとWAVファイルを保存しました"
	var files []discordFile
	for _, f := range []*os.File{srcMIDI, dstWAV} {
		if _, err := f.Seek(0, 0); err != nil {
			log.Println(err)
			continue
		}
		files = append(files, discordFile{name: filepath.Base(f.Name()), r: f})
	}
	if report != "" {
		withReport := content + "\n4声体和声の添削結果:\n```\n" + report + "```"
		if utf8.RuneCountInString(withReport) <= discordContentLimit {
			content = withReport
		} else {
			content += "\n4声体和声の添削結果が長いためファイルとして添付します"
			name := strings.TrimSuffix(filepath.Base(srcMIDI.Name()), ".mid") + "-check.txt"
			files = append(files, discordFile{name: name, r: strings.NewReader(report)})
		}
	}
	return content, files
}
