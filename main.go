//go:generate sh -c "$GOPATH/bin/2goarray activeIcon main < assets/nord-active.png > active_icon.go"
//go:generate sh -c "$GOPATH/bin/2goarray inactiveIcon main < assets/nord-inactive.png > inactive_icon.go"

package main

import log "github.com/sirupsen/logrus"

func main() {
	log.AddHook(&Hook{})
	newNordTray().run()
}
