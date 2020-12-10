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
		status    Status
		connected bool
	}
)

const (
	STALLED Status = iota
	FAILED
	NONETWORK
	DONE

	CONNECTED    = "Connected"
	DISCONNECTED = "Disconnected"

	noNetErrStr = "Please check your internet connection and try again"
)

var (
	ErrNoNet = errors.New("no network connection")
	statusRe = regexp.MustCompile("Status: (.*)")
)

func (n *NordVPN) Update() {
	n.Lock()
	defer n.Unlock()

	out, err := execCmd(2*time.Second, "nordvpn", "status")
	if err != nil {
		n.parseErr("update", err)
		return
	}

	n.status = DONE
	n.parse(out)
}

func (n *NordVPN) parse(data string) {
	var status string
	ss := statusRe.FindStringSubmatch(data)
	if len(ss) > 1 {
		status = ss[1]
	}

	switch status {
	case CONNECTED:
		n.connected = true
	case DISCONNECTED:
		n.connected = false
	default:
		n.status = STALLED
		log.Warnf("unrecognized status %s", status)
	}
}

func (n *NordVPN) parseErr(cmd string, err error) {
	if err == context.DeadlineExceeded {
		log.Errorf("on %s exceeded timeout", cmd)
		n.status = STALLED
	} else if err == ErrNoNet {
		log.Debugln("on %s: no network", cmd)
		n.status = NONETWORK
	} else {
		log.Errorf("on %s: %s", cmd, err)
		n.status = FAILED
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
