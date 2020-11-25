package main

import (
	"time"

	"github.com/getlantern/systray"
)

type NordTray struct {
	vpn         *NordVPN
	loopTimeout time.Duration
	loading     bool
	mConnect    *systray.MenuItem
	mDiconnect  *systray.MenuItem
	mQuit       *systray.MenuItem
}

func newNordTray() *NordTray {
	nt := &NordTray{vpn: &NordVPN{}, loopTimeout: 10 * time.Second}
	nt.mConnect = systray.AddMenuItem("Connect", "Connect NordVPN")
	nt.mDiconnect = systray.AddMenuItem("Disconnect", "Disconnect NordVPN")
	nt.mQuit = systray.AddMenuItem("Quit", "Quit NordTray")

	return nt
}

func (nt *NordTray) run() {
	systray.Run(nt.onReady, nt.onExit)
}

func (nt *NordTray) onReady() {
	systray.SetTitle("NordTray")
	systray.SetIcon(inactiveIcon)

	nt.update()
	go nt.loop()
}

func (nt *NordTray) onExit() {}

func (nt *NordTray) update() {
	nt.vpn.Update()
	nt.loading = false
	if nt.vpn.Status() == NONETWORK {
		nt.loopTimeout = 60 * time.Second
	} else {
		nt.loopTimeout = 10 * time.Second
	}

	nt.determineIcon()

	switch nt.vpn.Status() {
	case DONE:
		if nt.vpn.Connected() {
			nt.mConnect.Hide()
			nt.mDiconnect.Show()
		} else {
			nt.mConnect.Show()
			nt.mDiconnect.Hide()
		}
	case STALLED:
		nt.mConnect.Show()
		nt.mDiconnect.Show()
	case FAILED:
		nt.mConnect.Show()
		nt.mDiconnect.Hide()
	}
}

func (nt *NordTray) loop() {
	for {
		select {
		case <-nt.mConnect.ClickedCh:
			go nt.loadingIcon()
			nt.vpn.Connect()
			nt.update()
		case <-nt.mDiconnect.ClickedCh:
			go nt.loadingIcon()
			nt.vpn.Disconnect()
			nt.update()
		case <-nt.mQuit.ClickedCh:
			systray.Quit()
			return
		case <-time.After(nt.loopTimeout):
			nt.update()
		}
	}
}

func (nt *NordTray) loadingIcon() {
	for nt.loading = true; nt.loading; {
		systray.SetIcon(inactiveIcon)
		time.Sleep(300 * time.Millisecond)
		systray.SetIcon(activeIcon)
		time.Sleep(300 * time.Millisecond)
	}

	nt.determineIcon()
}

func (nt *NordTray) determineIcon() {
	if nt.vpn.Status() == DONE && nt.vpn.Connected() {
		systray.SetIcon(activeIcon)
	} else {
		systray.SetIcon(inactiveIcon)
	}
}
