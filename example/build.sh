if [[ $OSTYPE =~ "linux" ]]
then
## Linux
sudo apt-get install xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev
GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -ldflags '-H windowsgui -w -s' -o tipbar
fi
if [[ $OSTYPE =~ "darwin" ]]
then
## Mac
GOARCH=amd64 CGO_ENABLED=1 GOOS=darwin go build -ldflags '-w -s' -o tipbar
cp -r ./build/TipBar.app ./
mv tipbar ./TipBar.app/Contents/MacOS
fi
