package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"go.etcd.io/bbolt"

	"net/http"
	_ "net/http/pprof"
)

type router struct {
	leaseBinding *MikrotikDataTable

	host     string
	user     string
	password string
}

type appData struct {
	routers map[string]*router

	app fyne.App
	win fyne.Window

	bindings []*MikrotikDataTable
	current  *router

	db *bbolt.DB

	key *secretKey

	currentView string
}

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	a := app.NewWithID("github.com.bluebugs.fytastik")

	myApp := &appData{routers: map[string]*router{}, app: a, win: a.NewWindow("Mikrotik Router"), bindings: []*MikrotikDataTable{}}
	myApp.openDB()

	myApp.createUI()
	defer myApp.Close()
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
}
