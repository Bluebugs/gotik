package main

import (
	"context"
	"net"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"go.etcd.io/bbolt"
	"tailscale.com/tsnet"
)

type router struct {
	leaseBinding *MikrotikDataTable

	ssh *remote

	err error

	host     string
	user     string
	password string
	ssl      bool
}

type appData struct {
	routers   map[string]*router
	neighbors *MikrotikRouterList

	app fyne.App
	win fyne.Window
	m   *fyne.Menu

	bindings []*MikrotikDataTable
	current  *router
	identity binding.String

	db *bbolt.DB

	key *secretKey

	currentView, currentTab string

	ts           *tsnet.Server
	dial         func(ctx context.Context, network, address string) (net.Conn, error)
	cancel       context.CancelFunc
	useTailScale bool
}

var tcpDialer = net.Dialer{Timeout: 5 * time.Second}

func main() {
	a := app.NewWithID("github.com.bluebugs.gotik")
	a.Settings().SetTheme(&myTheme{})

	myApp := &appData{
		routers:   map[string]*router{},
		app:       a,
		win:       a.NewWindow("Mikrotik Router"),
		bindings:  []*MikrotikDataTable{},
		dial:      tcpDialer.DialContext,
		cancel:    func() {},
		neighbors: NewMikrotikRouterList(),
	}
	lastHost, _ := myApp.openDB()

	myApp.createUI(lastHost)
	defer myApp.Close()

	if desk, ok := a.(desktop.App); ok {
		myApp.m = fyne.NewMenu("Gotik", fyne.NewMenuItem("Show", func() {
			myApp.win.Show()
		}))
		desk.SetSystemTrayMenu(myApp.m)
	}

	selfManage(a, myApp.win)
	myApp.win.ShowAndRun()
}

func (a *appData) Close() {
	for _, value := range a.bindings {
		value.Close()
	}
	for _, value := range a.routers {
		if value.leaseBinding != nil {
			value.leaseBinding.Close()
		}
	}
	if a.useTailScale {
		a.tailScaleDisconnect()
	}
}
