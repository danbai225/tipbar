package core

import (
	_ "embed"
	"fmt"
	"fyne.io/systray"
	logs "github.com/danbai225/go-logs"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/ncruces/zenity"
	ghook "github.com/robotn/gohook"
	"sync"
	"time"
)

//go:embed ico.ico
var iconBs []byte

//<editor-fold desc="APP主体结构体">

type App struct {
	title        []*title
	module       []*Module
	EventRegList []chan ghook.Event
	config       config
	tip          chan tip
	titleLock    sync.Mutex
	g            *ghttp.Server
	name         string
	version      string
	ico          []byte
	index        func(r *ghttp.Request)
}

func NewApp(index func(r *ghttp.Request), configPath, name, version string, ioc []byte) (*App, error) {
	configP := "config.json"
	if configPath != "" {
		configP = configPath
	}
	ih := func(r *ghttp.Request) {
		r.Response.Write(fmt.Sprintf("Hello %s", name))
	}
	if index != nil {
		ih = index
	}
	app := App{config: config{configName: configP},
		title: make([]*title, 0), module: make([]*Module, 0), tip: make(chan tip, 10),
		index:   ih,
		version: version,
		name:    name,
		ico:     ioc,
	}
	//加载配置
	err := app.config.load()
	if err != nil {
		return nil, err
	}
	app.g = g.Server()
	if app.config.HTTPPort == 0 {
		app.config.HTTPPort = 7989
	}
	app.g.SetPort(int(app.config.HTTPPort))
	if app.config.LogsDir != "" {
		logs.SetLogsDir(app.config.LogsDir)
	}
	return &app, nil
}
func (a *App) addTitle(module *Module, titleText string) {
	for i := range a.title {
		t := a.title[i]
		if t.module == module {
			t.content = titleText
			return
		}
	}
	a.title = append(a.title, &title{module: module, content: titleText})
}
func (a *App) removeTitle(module *Module) {
	off := 0
	for i := range a.title {
		if a.title[i].module == module {
			off++
			continue
		} else {
			if i+off != i {
				a.title[i+off] = a.title[i]
			}
		}
	}
	if off > 0 {
		a.title = a.title[:len(a.title)-off]
	}
}
func (a *App) Run() error {
	//获取模块配置
	for _, m := range a.module {
		if c, has := a.config.Module[m.name]; has {
			m.Config = c.Config
			m.Port = a.config.HTTPPort
			if m.route != nil {
				group := a.g.Group("/" + m.name)
				m.route(group)
			}
		}
	}
	go func() {
		// 优雅关闭
		NewHook().Close(
			func() {
				_ = a.g.Shutdown()
				systray.Quit()
			},
		)
	}()
	//按键监听
	go func() {
		EvChan := ghook.Start()
		for ev := range EvChan {
			for _, events := range a.EventRegList {
				events <- ev
			}
		}
	}()
	//运行主体
	systray.Run(a.onReady, a.exit)
	return nil
}
func (a *App) doTip() {
	for t := range a.tip {
		a.titleLock.Lock()
		systray.SetTitle(t.content)
		time.Sleep(t.time)
		systray.SetTitle("")
		a.titleLock.Unlock()
	}
}
func (a *App) doTitle() {
	for {
		for _, t := range a.title {
			a.titleLock.Lock()
			systray.SetTitle(t.content)
			a.titleLock.Unlock()
			time.Sleep(5 * time.Second)
			systray.SetTitle("")
		}
		time.Sleep(time.Second)
	}
}
func middlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}
func (a *App) onReady() {
	//设置程序基本图标等等。。
	if a.ico != nil {
		systray.SetTemplateIcon(a.ico, a.ico)
	} else {
		systray.SetTemplateIcon(iconBs, iconBs)
	}
	//运行http
	a.g.BindHandler("/", a.index)
	// 跨域
	a.g.Use(middlewareCORS)
	for _, module := range a.module {
		item := systray.AddMenuItem(module.itemName, module.tooltip)
		if module.onReady != nil {
			go module.onReady(item)
		}
	}
	systray.SetTooltip(fmt.Sprintf("%s\n版本号:%s", a.name, a.version))
	quit := systray.AddMenuItem("退出", "退出程序")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()
	go a.doTip()
	go a.doTitle()
	go a.g.Run()

}
func (a *App) exit() {
	ghook.StopEvent()
	for _, module := range a.module {
		if module.exit != nil {
			go module.exit()
		}
	}
	err := a.g.Shutdown()
	if err != nil {
		logs.Err(err)
	}
}
func (a *App) RegisterModule(module ...*Module) {
	for i := range module {
		m := module[i]
		if mc, has := a.config.Module[m.name]; !has || !mc.Enable {
			continue
		}
		m.app = a
		a.module = append(a.module, m)
	}
}

//</editor-fold>

// NewModule name是该模块的名字 itemName是显示在任务栏的名字 tooltip是鼠标移动到该模块子项时提示的文字 onReady是初始化完成后执行的函数 exit是退出时执行的函数 route是http录音函数
func NewModule(name, itemName, tooltip string, onReady func(item *systray.MenuItem), exit func(), route func(*ghttp.RouterGroup)) *Module {
	module := Module{name: name, itemName: itemName, tooltip: tooltip}
	module.onReady = onReady
	module.exit = exit
	module.route = route
	return &module
}

type tip struct {
	module  *Module
	content string
	time    time.Duration
}
type title struct {
	module  *Module
	content string
}

//<editor-fold desc="模块结构体">

type Module struct {
	onReady   func(item *systray.MenuItem)
	exit      func()
	app       *App
	name      string
	itemName  string
	tooltip   string
	Config    interface{}
	route     func(*ghttp.RouterGroup)
	Port      uint16
	eventChan chan ghook.Event
}

func (m *Module) UnmarshalConfig(dst interface{}) error {
	return Unmarshal(m.Config, dst)
}

func (m *Module) SetTitle(title string) {
	m.app.addTitle(m, title)
}
func (m *Module) RemoveTitle() {
	m.app.removeTitle(m)
}
func (m *Module) Tip(str string, time time.Duration) {
	go func() {
		m.app.tip <- tip{
			module:  m,
			content: str,
			time:    time,
		}
	}()
}
func (m *Module) Notify(str string) {
	err := zenity.Notify(str)
	if err != nil {
		logs.Err(err)
	}
}
func (m *Module) SaveConfig(c interface{}) {
	m.Config = c
	m.app.config.saveConfig(m, m.Config)
	err := m.app.config.save()
	if err != nil {
		logs.Err()
	}
}
func (m *Module) GetRootUrl() string {
	return fmt.Sprintf("http://localhost:%d/%s", m.Port, m.name)
}
func (m *Module) GetAPPName() string {
	return m.app.name
}
func (m *Module) GetAPPVersion() string {
	return m.app.version
}
func (m *Module) RegEvent() chan ghook.Event {
	m.eventChan = make(chan ghook.Event, 0)
	return m.eventChan
}

//</editor-fold>
