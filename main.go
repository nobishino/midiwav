package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
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
	
	dstWAV, err := os.Create(strings.TrimSuffix(srcPath, ".mid") + ".wav")
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer dstWAV.Close()
	
	if err := midiToWAVE(dstWAV, srcMIDI); err != nil {
		return fmt.Errorf("failed to convert MIDI to WAV: %w", err)
	}
	
	if discordWebhookURL != "" {
		if err := postToDiscord(discordWebhookURL, srcMIDI, dstWAV); err != nil {
			log.Println(err)
		}
	}
	return nil
}
