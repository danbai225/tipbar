package hello

import (
	"fmt"
	"fyne.io/systray"
	logs "github.com/danbai225/go-logs"
	"github.com/danbai225/tipbar/core"
	"github.com/gogf/gf/net/ghttp"
)

var hello *core.Module
var count = 0

func ExportModule() *core.Module {
	hello = core.NewModule("hello", "helloModule", "这是一个测试模块", onReady, exit, router)
	return hello
}
func onReady(item *systray.MenuItem) {
	for {
		select {
		case <-item.ClickedCh:
			item.SetTitle(fmt.Sprintf("hell:%d", count))
			count++
		}
	}
}
func exit() {
	logs.Info("exit")
}

func router(group *ghttp.RouterGroup) {
	group.GET("/", func(r *ghttp.Request) {
		r.Response.Write("hello!!!", count)
	})
}
