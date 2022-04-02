package main

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type (
	Status int

	NordVPN struct {
		sync.Mutex
		status     Status
		connected  bool
		killswtich bool
		server     string
	}
)

const (
	STALLED Status = iota
	FAILED
	NONETWORK
	DONE

	CONNECTED    = "Connected"
	DISCONNECTED = "Disconnected"

	ENABLED  = "enabled"
	DISABLED = "disabled"

	noNetErrStr = "Please check your internet connection and try again"
)

var (
	ErrNoNet     = errors.New("no network connection")
	reConStatus  = regexp.MustCompile("Status: (.*)")
	reServer     = regexp.MustCompile("Current server: (.*)")
	reKillSwitch = regexp.MustCompile("Kill Switch: (.*)")
)

func (n *NordVPN) Update() {
	n.Lock()
	defer n.Unlock()

	out, err := execCmd(2*time.Second, "nordvpn", "status")
	if err != nil {
		n.parseErr("update on status", err)
		return
	}
	n.status = DONE
	n.parseStatus(out)

	out, err = execCmd(2*time.Second, "nordvpn", "settings")
	if err != nil {
		n.parseErr("update", err)
		return
	}
	n.parseSettings(out)
}

func (n *NordVPN) parseStatus(data string) {
	var conStatus string
	ss := reConStatus.FindStringSubmatch(data)
	if len(ss) > 1 {
		conStatus = ss[1]
	}

	switch conStatus {
	case CONNECTED:
		n.connected = true
		n.updateServer(data)
	case DISCONNECTED:
		n.connected = false
	default:
		n.status = STALLED
		log.Warnf("unrecognized status %s: %s", data, conStatus)
	}
}

func (n *NordVPN) parseSettings(data string) {
	var ks string
	ss := reKillSwitch.FindStringSubmatch(data)
	if len(ss) > 1 {
		ks = ss[1]
	}

	switch ks {
	case ENABLED:
		n.killswtich = true
	case DISABLED:
		n.killswtich = false
	default:
		n.status = STALLED
		log.Warnf("unrecognized kill swtich %s: %s", data, ks)
	}
}

func (n *NordVPN) parseErr(cmd string, err error) {
	switch err {
	case context.DeadlineExceeded:
		log.Errorf("on %s exceeded timeout", cmd)
		n.status = STALLED
	case ErrNoNet:
		log.Debugln("on %s: no network", cmd)
		n.status = NONETWORK
	default:
		log.Errorf("on %s: %s", cmd, err)
		n.status = FAILED
	}
}

func (n *NordVPN) updateServer(data string) {
	ss := reServer.FindStringSubmatch(data)
	if len(ss) > 1 {
		n.server = ss[1]
	}
}

func (n *NordVPN) Status() Status {
	return n.status
}

func (n *NordVPN) Connect() {
	_, err := execCmd(3*time.Second, "nordvpn", "c")
	if err != nil {
		n.parseErr("connect", err)
		return
	}
	n.status = DONE
}

func (n *NordVPN) Disconnect() {
	_, err := execCmd(3*time.Second, "nordvpn", "d")
	if err != nil {
		n.parseErr("disconnect", err)
		return
	}
	n.status = DONE
}

func (n *NordVPN) Connected() bool {
	return n.connected
}

func (n *NordVPN) Server() string {
	return n.server
}

func (n *NordVPN) KillSwitch() bool {
	return n.killswtich
}

func (n *NordVPN) SetKillSwitch(v bool) {
	s := "on"
	if !v {
		s = "off"
	}

	_, err := execCmd(3*time.Second, "nordvpn", "set", "killswitch", s)
	if err != nil {
		n.parseErr("set killswitch", err)
		return
	}
	n.status = DONE
}
