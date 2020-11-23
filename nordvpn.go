package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
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
	DONE

	CONNECTED    = "Connected"
	DISCONNECTED = "Disconnected"

	template = "\r-\r  \r\r-\r  \rStatus: %s\nCurrent server: is57.nordvpn.com\nCountry: Iceland\nCity: Reykjavik\nYour new IP: 45.133.192.139\nCurrent technology: NordLynx\nTransfer: 127.29 MiB received, 414.97 MiB sent\nUptime: 36 minutes 7 seconds\n"
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
	fmt.Sscanf(data, template, &status)

	switch status {
	case CONNECTED:
		n.connected = true
	case DISCONNECTED:
		n.connected = false
	default:
		n.status = STALLED
		log.Printf("WARN unrecognized status %s", status)
	}
}

func (n *NordVPN) parseErr(cmd string, err error) {
	if err == context.DeadlineExceeded {
		log.Printf("ERR on %s exceeded timeout", cmd)
		n.status = STALLED
	} else {
		log.Printf("ERR on %s: %s", cmd, err)
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
