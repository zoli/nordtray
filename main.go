package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/getlantern/systray"
)

var (
	nord = &NordVPN{}

	activeIcon, inActiveIcon []byte
)

func main() {
	var err error
	if activeIcon, err = ioutil.ReadFile("assets/nord-active.png"); err != nil {
		log.Fatalln(err)
	}
	if inActiveIcon, err = ioutil.ReadFile("assets/nord-inactive.png"); err != nil {
		log.Fatalln(err)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("NordTray")
	systray.SetIcon(inActiveIcon)

	mConnect := systray.AddMenuItem("Connect", "Connect NordVPN")
	mDiconnect := systray.AddMenuItem("Disconnect", "Disconnect NordVPN")
	mQuit := systray.AddMenuItem("Quit", "Quit NordTray")

	update := func() {
		nord.Update()
		if nord.Status() == DONE && nord.Connected() {
			systray.SetIcon(activeIcon)
		} else {
			systray.SetIcon(inActiveIcon)
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
