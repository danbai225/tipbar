go build -ldflags="-s -w -H windowsgui" -o tipbar.exe
upx.exe tipbar.exe