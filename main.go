//go:generate sh -c "$GOPATH/bin/2goarray activeIcon main < assets/nord-active.png > active_icon.go"
//go:generate sh -c "$GOPATH/bin/2goarray inactiveIcon main < assets/nord-inactive.png > inactive_icon.go"

package main

import (
	"time"

	"github.com/getlantern/systray"
)

var (
	nord = &NordVPN{}
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("NordTray")
	systray.SetIcon(inactiveIcon)

	mConnect := systray.AddMenuItem("Connect", "Connect NordVPN")
	mDiconnect := systray.AddMenuItem("Disconnect", "Disconnect NordVPN")
	mQuit := systray.AddMenuItem("Quit", "Quit NordTray")

	update := func() {
		nord.Update()
		if nord.Status() == DONE && nord.Connected() {
			systray.SetIcon(activeIcon)
		} else {
			systray.SetIcon(inactiveIcon)
		}

		switch nord.Status() {
		case DONE:
			if nord.Connected() {
				mConnect.Hide()
				mDiconnect.Show()
			} else {
				mConnect.Show()
				mDiconnect.Hide()
			}
		case STALLED:
			mConnect.Show()
			mDiconnect.Show()
		case FAILED:
			mConnect.Show()
			mDiconnect.Hide()
		}
	}

	update()
	go func() {
		for {
			select {
			case <-mConnect.ClickedCh:
				nord.Connect()
				update()
			case <-mDiconnect.ClickedCh:
				nord.Disconnect()
				update()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case <-time.After(10 * time.Second):
				update()
			}
		}
	}()
}

func onExit() {

}
