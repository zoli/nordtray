all: generate build

generate:
	go generate
build:
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -ldflags \
	"-linkmode external -extldflags \"${LDFLAGS}\"" -o nordtray

install: install_bin install_desktop
uninstall:
	rm -rf /usr/lib/nordtray
	rm /usr/share/applications/NordTray.desktop
	rm /usr/share/icons/hicolor/48x48/apps/NordTray.png
install_bin:
	mkdir -p /usr/lib/nordtray
	mv ./nordtray /usr/lib/nordtray/
install_desktop:
	install -m 644 -p ./assets/nordtray.desktop /usr/share/applications/NordTray.desktop
	install -m 644 -p ./assets/nord-active.png /usr/share/icons/hicolor/48x48/apps/NordTray.png

clean:
	rm nordtray
