BINARY_NAME=buff_report
VERSION=1.0
all:
	rm -rf build
	mkdir -p build
	cp config.example.txt build/config.txt
	cp cookies.example.txt build/cookies_buff.txt
	cp cookies.example.txt build/cookies_steam.txt
	cp docs/网易buff利润报表软件使用教程.pdf build/
	cp docs/EditThisCookie.crx build/
	GOARCH=arm64 GOOS=darwin go build -o build/${BINARY_NAME}  -buildmode=pie .
	GOARCH=amd64 GOOS=darwin go build -o build/${BINARY_NAME}-darwin  -buildmode=pie .
	GOARCH=amd64 GOOS=linux go build -o build/${BINARY_NAME}-linux  -buildmode=pie .
	GOARCH=amd64 GOOS=windows go build -a -ldflags="-w -s" -trimpath -o build/${BINARY_NAME}.exe  -buildmode=pie .
	cd build && rar a ${BINARY_NAME}_v${VERSION}.rar ${BINARY_NAME}.exe EditThisCookie.crx config.txt cookies_buff.txt cookies_steam.txt 网易buff利润报表软件使用教程.pdf
	cd build && rar a ${BINARY_NAME}_v${VERSION}-mac-arm.rar ${BINARY_NAME} EditThisCookie.crx config.txt cookies_buff.txt cookies_steam.txt 网易buff利润报表软件使用教程.pdf
