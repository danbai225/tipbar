## windows
GOARCH=amd64 CGO_ENABLED=0 GOOS=windows go build -ldflags '-H windowsgui -w -s' -o tipbar.exe
## Linux
GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -ldflags '-H windowsgui -w -s' -o tipbar
## Mac
GOARCH=amd64 CGO_ENABLED=1 GOOS=darwin go build -ldflags '-w -s' -o tipbar
cp -r ./build/TipBar.app ./
mv tipbar ./TipBar.app/Contents/MacOS
