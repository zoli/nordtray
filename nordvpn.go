package main

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"time"

	parallelizer "github.com/shomali11/parallelizer"
	log "github.com/sirupsen/logrus"
)

type (
	Status int

	NordVPN struct {
		sync.Mutex
		status       Status
		connected    bool
		killswitchOn bool
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
	statusRe     = regexp.MustCompile("Status: (.*)")
	killswitchRe = regexp.MustCompile("Kill Switch: (.*)")
)

func (n *NordVPN) Update() {
	log.SetLevel(log.DebugLevel)
	n.Lock()
	defer n.Unlock()
	group := parallelizer.NewGroup()
	defer group.Close()

	group.Add(n.parseUpdate)

	group.Add(n.parseKillswitch)

	ctx, cancel := context.WithTimeout(context.Background(), 2100*time.Millisecond)
	defer cancel()

	err := group.Wait(parallelizer.WithContext(ctx))

	if err != nil {
		log.Errorf("Error in parallelizer: %v", err)
	}
}

func (n *NordVPN) parseUpdate() {
	out, err := execCmd(2*time.Second, "nordvpn", "status")
	if err != nil {
		n.parseErr("update", err)
		return
	}
	n.status = DONE
	match := n.getFirstMatch(out, statusRe)

	switch match {
	case CONNECTED:
		n.connected = true
	case DISCONNECTED:
		n.connected = false
	default:
		n.status = STALLED
		log.Warnf("unrecognized status %s", match)
	}
}

func (n *NordVPN) parseKillswitch() {
	out, err := execCmd(2*time.Second, "nordvpn", "settings")
	if err != nil {
		n.parseErr("killswitch status", err)
		return
	}
	n.status = DONE
	match := n.getFirstMatch(out, killswitchRe)

	switch match {
	case ENABLED:
		n.killswitchOn = true
	case DISABLED:
		n.killswitchOn = false
	default:
		n.status = STALLED
		log.Warnf("unrecognized killswitch status %s", match)
	}
}

func (n *NordVPN) getFirstMatch(data string, regexp *regexp.Regexp) string {
	ss := regexp.FindStringSubmatch(data)
	if len(ss) > 1 {
		return ss[1]
	}
	return ""
}

func (n *NordVPN) parseErr(cmd string, err error) {
	if err == context.DeadlineExceeded {
		log.Errorf("on %s exceeded timeout", cmd)
		n.status = STALLED
	} else if err == ErrNoNet {
		log.Debugf("on %s: no network\n", cmd)
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

func (n *NordVPN) EnableKillswitch() {
	out, err := execCmd(2*time.Second, "nordvpn", "set", "killswitch", ENABLED)
	if err != nil {
		n.parseErr("enabled killswitch", err)
		return
	}
	log.Debug(out)
	n.status = DONE
}

func (n *NordVPN) DisableKillswitch() {
	out, err := execCmd(2*time.Second, "nordvpn", "set", "killswitch", DISABLED)
	if err != nil {
		n.parseErr("disabled killswitch", err)
		return
	}
	log.Debug(out)
	n.status = DONE
}

func (n *NordVPN) Connected() bool {
	return n.connected
}
