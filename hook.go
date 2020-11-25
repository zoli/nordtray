package main

import (
	"fmt"

	"github.com/gen2brain/beeep"
	log "github.com/sirupsen/logrus"
)

type Hook struct{}

func (h *Hook) Levels() []log.Level {
	return []log.Level{log.ErrorLevel}
}

func (h *Hook) Fire(entry *log.Entry) error {
	return beeep.Notify("NordTray", fmt.Sprintf("ERR %s", entry.Message), "")
}
