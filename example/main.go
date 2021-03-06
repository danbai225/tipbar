package main

import (
	logs "github.com/danbai225/go-logs"
	"github.com/danbai225/tipbar/core"
	"github.com/danbai225/tipbar/example/module/hello"
	"os"
)

func main() {
	var a *core.App
	var err error
	//指定配置文件参数 默认读取当前文件夹下config.json
	if len(os.Args) > 1 {
		a, err = core.NewApp(nil, os.Args[1], "TipBar", "v0.0.1", nil)
	} else {
		a, err = core.NewApp(nil, "", "TipBar", "v0.0.1", nil)
	}
	if err != nil {
		logs.Err(err)
		return
	}
	//注册模块
	a.RegisterModule(
		hello.ExportModule(),
	)
	err = a.Run()
	if err != nil {
		logs.Err(err)
		return
	}
}
