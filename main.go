package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	discordWebhook := os.Getenv("MIDIWAV_DISCORD_WEBHOOK")
	if discordWebhook == "" {
		log.Println("MIDIWAV_DISCORD_WEBHOOK is not set. skipping Discord upload.")
	}
	targetDir := os.Getenv("MIDIWAV_DIR")

	exec := &executer{
		discordWebhook: discordWebhook,
		midiWavDir:     targetDir,
		watchMode:      true, // TODO: make configurable
	}
	if err := exec.start(); err != nil {
		log.Fatal(err)
	}
}

type executer struct {
	discordWebhook string
	midiWavDir     string
	watchMode      bool
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
	unprocessed, err := findUnprocessedMIDIFiles(e.midiWavDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d unprocessed MIDI files.\n", len(unprocessed))

	for _, srcPath := range unprocessed {
		fmt.Println("Processing:", srcPath)
		srcMIDI, err := os.Open(srcPath)
		if err != nil {
			log.Fatal(err)
		}
		defer srcMIDI.Close()
		dstWAV, err := os.Create(strings.TrimSuffix(srcPath, ".mid") + ".wav")
		if err != nil {
			log.Fatal(err)
		}
		defer dstWAV.Close()
		if err := midiToWAVE(dstWAV, srcMIDI); err != nil {
			log.Fatal(err)
		}
		if e.discordWebhook != "" {
			if err := postToDiscord(e.discordWebhook, srcMIDI, dstWAV); err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}
