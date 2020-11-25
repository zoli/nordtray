package main

import (
	"time"

	"github.com/getlantern/systray"
)

type NordTray struct {
	vpn        *NordVPN
	mConnect   *systray.MenuItem
	mDiconnect *systray.MenuItem
	mQuit      *systray.MenuItem
}

func newNordTray() *NordTray {
	nt := &NordTray{vpn: &NordVPN{}}
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
	if nt.vpn.Status() == DONE && nt.vpn.Connected() {
		systray.SetIcon(activeIcon)
	} else {
		systray.SetIcon(inactiveIcon)
	}

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
			nt.vpn.Connect()
			nt.update()
		case <-nt.mDiconnect.ClickedCh:
			nt.vpn.Disconnect()
			nt.update()
		case <-nt.mQuit.ClickedCh:
			systray.Quit()
			return
		case <-time.After(10 * time.Second):
			nt.update()
		}
	}
}
