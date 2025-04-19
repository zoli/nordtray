package main

import (
	"sort"
	"time"

	"github.com/getlantern/systray"
)

type NordTray struct {
	vpn         *NordVPN
	loopTimeout time.Duration
	loading     bool
	mCountries  *systray.MenuItem
	mConnect    *systray.MenuItem
	mDiconnect  *systray.MenuItem
	mQuit       *systray.MenuItem
	mKillSwitch *systray.MenuItem
	mMeshnet    *systray.MenuItem
	countryMap  map[string]*systray.MenuItem
}

func newNordTray() *NordTray {
	nt := &NordTray{vpn: &NordVPN{}, loopTimeout: 10 * time.Second}

	countryCodeMap := nt.vpn.GetCountryCodeMap()
	countryNames := make([]string, 0, len(countryCodeMap))
	for countryName := range countryCodeMap {
		countryNames = append(countryNames, countryName)
	}
	sort.Strings(countryNames)

	nt.countryMap = map[string]*systray.MenuItem{}
	nt.mCountries = systray.AddMenuItem("Countries", "Select a country to connect NordVPN server to")
	for _, name := range countryNames {
		countryMenuItem := nt.mCountries.AddSubMenuItemCheckbox(name, "", false)
		// as the countries are dynamic, we cannot use a single goroutine with for-select that handles all countries
		// hence this one goroutine per country approach. worry not, it's still CPU efficient
		go func() {
			for {
				<-countryMenuItem.ClickedCh
				for _, m := range nt.countryMap {
					if m != countryMenuItem {
						m.Uncheck()
					}
				}
				// UI's checked state is not automatically propagated, we must handle (un)checking ourselves
				countryMenuItem.Check()
			}
		}()
		code := countryCodeMap[name]
		nt.countryMap[code] = countryMenuItem
	}

	nt.mConnect = systray.AddMenuItem("Connect", "Connect NordVPN")
	nt.mDiconnect = systray.AddMenuItem("Disconnect", "Disconnect NordVPN")
	nt.mKillSwitch = systray.AddMenuItemCheckbox("Kill Switch", "Toggle kill switch", false)
	nt.mMeshnet = systray.AddMenuItemCheckbox("Meshnet", "Toggle meshnet", false)
	nt.mQuit = systray.AddMenuItem("Quit", "Quit NordTray")

	return nt
}

func (nt *NordTray) getCheckedCountry() string {
	for c, m := range nt.countryMap {
		if m.Checked() {
			return c
		}
	}
	return ""
}

func (nt *NordTray) run() {
	systray.Run(nt.onReady, nt.onExit)
}

func (nt *NordTray) onReady() {
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
			systray.SetTooltip(nt.vpn.Server())
		} else {
			nt.mConnect.Show()
			nt.mDiconnect.Hide()
		}
		if nt.vpn.KillSwitch() {
			nt.mKillSwitch.Check()
		} else {
			nt.mKillSwitch.Uncheck()
		}
		if nt.vpn.Meshnet() {
			nt.mMeshnet.Check()
		} else {
			nt.mMeshnet.Uncheck()
		}
	case STALLED:
		nt.mConnect.Show()
		nt.mDiconnect.Show()
		nt.mKillSwitch.Disable()
	case FAILED:
		nt.mConnect.Show()
		nt.mDiconnect.Hide()
		nt.mKillSwitch.Disable()
	}

	if nt.vpn.Meshnet() {
		nt.mMeshnet.Enable()
	}
}

func (nt *NordTray) loop() {
	for {
		select {
		case <-nt.mConnect.ClickedCh:
			go nt.loadingIcon()
			nt.vpn.Connect(nt.getCheckedCountry())
			nt.update()
		case <-nt.mDiconnect.ClickedCh:
			go nt.loadingIcon()
			nt.vpn.Disconnect()
			nt.update()
		case <-nt.mKillSwitch.ClickedCh:
			nt.update()
			nt.vpn.SetKillSwitch(!nt.vpn.KillSwitch())
			nt.update()
		case <-nt.mQuit.ClickedCh:
			systray.Quit()
			return
		case <-nt.mMeshnet.ClickedCh:
			nt.vpn.SetMeshnet(!nt.vpn.Meshnet())
			nt.update()
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
