package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fynelabs/fynetailscale"
	"tailscale.com/tsnet"
)

func (a *appData) createUI(lastHost string) {
	tabs := container.NewAppTabs()
	tabs.OnSelected = func(ti *container.TabItem) {
		a.currentTab = ti.Text
		a.saveCurrentView()
	}

	header := widget.NewLabel("Not Connected")
	header.Alignment = fyne.TextAlignCenter
	footer := widget.NewLabel("")
	footer.Alignment = fyne.TextAlignCenter

	updateStatus := func(identity binding.String, ssl bool, err error) {
		if err != nil {
			tabs.Items = []*container.TabItem{}
			tabs.Refresh()

			footer.SetText(fmt.Sprintf("%v", err))
			if a.identity != nil {
				header.Bind(a.identity)
			} else {
				header.Unbind()
				header.SetText("Not Connected")
			}
			return
		}

		if ssl {
			header.TextStyle = fyne.TextStyle{Bold: true}
		} else {
			header.TextStyle = fyne.TextStyle{}
		}

		header.Bind(identity)
		footer.SetText("")
	}

	tree := widget.NewTreeWithStrings(routerOStree)
	tree.OnSelected = func(id string) {
		err := a.buildView(tabs, id)
		if err != nil {
			updateStatus(nil, false, err)
			return
		}

		a.saveCurrentView()
		updateStatus(a.identity, a.current.ssl, nil)
	}

	sel := widget.NewSelect([]string{}, a.selectHost(tabs, updateStatus))

	var useTailScale *widget.Check
	updateTailScale := func(b bool) {
		a.useTailScale = b
		if b {
			a.ts = new(tsnet.Server)

			if err := a.ts.Start(); err != nil {
				useTailScale.Checked = false
				useTailScale.Refresh()
				return
			}
		} else {
			a.tailScaleDisconnect()
		}
	}

	useTailScale = widget.NewCheck("Use tailscale", func(b bool) {
		updateTailScale(b)
		if a.ts != nil {
			a.tailScaleLogin()
		}
		a.saveCurrentView()
	})
	useTailScale.Checked = a.useTailScale
	updateTailScale(a.useTailScale)

	a.win.SetContent(NewSplit("Gotik", container.NewBorder(container.NewVBox(container.NewBorder(nil, nil,
		widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() { a.removeHost(sel) }),
		container.NewHBox(
			widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() { a.newHost(sel, "") }),
			widget.NewButtonWithIcon("", theme.MediaReplayIcon(), func() { a.reconnectHost(updateStatus, sel) }),
			widget.NewButtonWithIcon("", theme.SearchIcon(), func() { a.displayNeighbor(sel) }),
		),
		sel), useTailScale),
		nil, nil, nil, tree),
		container.NewBorder(header, footer, nil, nil, tabs)))
	a.win.Resize(fyne.NewSize(800, 600))

	if a.salt() != nil {
		a.getPassword(lastHost, sel)
	} else {
		if a.ts != nil {
			err := a.tailScaleLogin()
			if err != nil {
				dialog.ShowError(err, a.win)
			}
		}
	}
}

func (a *appData) tailScaleLogin() error {
	lc, err := a.ts.LocalClient()
	if err != nil {
		return err
	}

	var ctx context.Context
	ctx, a.cancel = context.WithCancel(context.Background())

	fynetailscale.NewLogin(ctx, a.win, lc, func(succeeded bool) {
		if succeeded {
			a.dial = a.ts.Dial
			fmt.Println("Connected")
		} else {
			a.dial = tcpDialer.DialContext
			fmt.Println("Failed to connect")
		}
	})

	return nil
}

func (a *appData) tailScaleDisconnect() {
	if a.ts == nil {
		return
	}
	a.ts.Close()
	a.ts = nil
	a.cancel()
	a.cancel = func() {}
}

func (a *appData) displayNeighbor(sel *widget.Select) {
	var d *dialog.CustomDialog

	neighbors := widget.NewListWithData(a.neighbors,
		func() fyne.CanvasObject {
			return widget.NewLabel("Mikrotik router somewhere (cc:2d:e0:e1:09:2a, 255.255.255.255) Mikrotik - 6.49.2 (stable)")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})
	neighbors.OnSelected = func(id int) {
		b, err := a.neighbors.GetItem(id)
		if err != nil {
			return
		}
		neighbor := b.(*MikrotikRouter)
		ip := neighbor.IP()
		if ip == "" {
			return
		}
		d.Hide()
		a.newHost(sel, ip)
	}
	content := container.New(&moreSpace{a.win}, neighbors)

	d = dialog.NewCustom("Neighbors", "Close", content, a.win)
	d.Show()
}

func (a *appData) newHost(sel *widget.Select, ip string) {
	host := widget.NewEntry()
	host.PlaceHolder = "127.0.0.1"
	host.Text = ip
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	ssl := widget.NewCheck("", nil)
	dialog.ShowForm("New router", "Connect", "Cancel",
		[]*widget.FormItem{
			{Text: "Host", Widget: host},
			{Text: "User", Widget: user},
			{Text: "Password", Widget: pass},
			{Text: "SSL", Widget: ssl},
		}, func(confirm bool) {
			if confirm {
				r := a.routerView(host.Text, ssl.Checked, user.Text, pass.Text)
				if r.err != nil {
					dialog.ShowError(r.err, a.win)
					return
				}
				a.routers[r.host] = r
				sel.Options = append(sel.Options, r.host)
				sel.SetSelected(r.host)
				sel.Refresh()

				if a.key == nil {
					a.createPassword(func() {
						if err := a.saveRouter(r, pass.Text); err != nil {
							dialog.ShowError(err, a.win)
							return
						}
					})
				} else {
					if err := a.saveRouter(r, pass.Text); err != nil {
						dialog.ShowError(err, a.win)
						return
					}
				}
			}
		}, a.win)
	a.win.Canvas().Focus(host)
}

func (a *appData) selectHost(tabs *container.AppTabs, updateStatus func(identity binding.String, ssl bool, err error)) func(s string) {
	return func(s string) {
		for _, b := range a.bindings {
			b.Close()
		}
		a.bindings = []*MikrotikDataTable{}
		a.identity = nil

		r, ok := a.routers[s]
		if !ok {
			updateStatus(nil, false, errors.New("router not found"))
			return
		}

		if r.err != nil {
			updateStatus(nil, false, r.err)
			return
		}

		identity, err := a.routerIdentity(r)
		if err != nil {
			updateStatus(nil, false, err)
			return
		}

		a.current = r
		a.identity = identity

		if a.currentView != "" {
			err := a.buildView(tabs, a.currentView)
			if err != nil {
				updateStatus(nil, false, err)
				return
			}
		} else {
			tabs.Items = []*container.TabItem{}
			tabs.Refresh()
		}

		a.saveCurrentView()
		updateStatus(a.identity, a.current.ssl, nil)
	}
}

func (a *appData) createPassword(gotKey func()) {
	password := widget.NewPasswordEntry()
	repeat := widget.NewPasswordEntry()
	dialog.ShowForm("Password", "Save", "Cancel",
		[]*widget.FormItem{
			{Text: "Password", Widget: password},
			{Text: "Confirm", Widget: repeat},
		}, func(confirm bool) {
			if confirm {
				if password.Text != repeat.Text {
					dialog.ShowError(fmt.Errorf("passwords do not match"), a.win)
					return
				}

				if err := a.createKey(password.Text); err != nil {
					dialog.ShowError(err, a.win)
					return
				}
				gotKey()
			}
		}, a.win)
	a.win.Canvas().Focus(password)
}

func (a *appData) getPassword(lastHost string, sel *widget.Select) {
	password := widget.NewPasswordEntry()
	dialog.ShowForm("Password", "Unlock", "Cancel",
		[]*widget.FormItem{
			{Text: "Password", Widget: password},
		}, func(confirm bool) {
			if confirm {
				if err := a.unlockKey(password.Text); err != nil {
					dialog.ShowError(err, a.win)
					return
				}

				if a.ts != nil {
					err := a.tailScaleLogin()
					if err != nil {
						dialog.ShowError(err, a.win)
					}
				}

				if err := a.loadRouters(sel); err != nil {
					dialog.ShowError(err, a.win)
					return
				}
				if len(sel.Options) > 0 {
					found := false
					for index, host := range sel.Options {
						if host == lastHost {
							found = true
							sel.SetSelectedIndex(index)
							break
						}
					}
					if !found {
						sel.SetSelectedIndex(0)
					}
				}
			}
		}, a.win)
	a.win.Canvas().Focus(password)
}

func (a *appData) buildView(tabs *container.AppTabs, view string) error {
	a.currentView = view

	if a.current == nil {
		return errors.New("no current router")
	}

	tabs.Items = []*container.TabItem{}

	lookup, ok := routerOSCommands[view]
	if !ok {
		return errors.New("no view found for " + view)
	}

	selectIndex := 0
	for _, cmd := range lookup {
		log.Println("loading", cmd.path)
		b, err := NewMikrotikData(a.dial, a.current.host, a.current.ssl, a.current.user, a.current.password, cmd.path)
		if err != nil {
			log.Println("failed to load", cmd.path, err)
			continue
		}
		if a.currentTab == cmd.title {
			selectIndex = len(tabs.Items)
		}
		tabs.Items = append(tabs.Items, container.NewTabItem(cmd.title, a.NewTableWithDataColumn(cmd.headers, b)))
	}
	tabs.SelectIndex(selectIndex)
	tabs.Refresh()
	log.Println("loaded", len(tabs.Items), "tabs for", view)

	return nil
}

func (a *appData) removeHost(sel *widget.Select) {
	if sel.Selected == "" {
		return
	}

	for _, value := range a.bindings {
		value.Close()
	}
	a.bindings = nil

	r := a.routers[sel.Selected]
	if r.leaseBinding != nil {
		r.leaseBinding.Close()
	}
	delete(a.routers, sel.Selected)

	a.deleteRouter(r)

	sel.ClearSelected()
	for i, v := range sel.Options {
		if v == r.host {
			sel.Options = append(sel.Options[:i], sel.Options[i+1:]...)
			if len(sel.Options) > 0 {
				sel.SetSelectedIndex(i)
			}
			break
		}
	}
	sel.Refresh()
}

func (a *appData) reconnectHost(updateStatus func(identity binding.String, ssl bool, err error), sel *widget.Select) {
	if sel.Selected == "" {
		return
	}

	r, ok := a.routers[sel.Selected]
	if !ok {
		updateStatus(nil, false, fmt.Errorf("no router found for %s", sel.Selected))
		return
	}

	if r.leaseBinding != nil {
		r.leaseBinding.Close()
		r.leaseBinding = nil
	}
	r.leaseBinding, r.err = NewMikrotikData(a.dial, r.host, r.ssl, r.user, r.password, "/ip/dhcp-server/lease")
	if r.err != nil {
		updateStatus(nil, false, r.err)
	} else {
		sel.SetSelected(sel.Selected)
	}
}

func (a *appData) routerView(host string, ssl bool, user, pass string) *router {
	var err error
	r := &router{host: host, user: user, password: pass, ssl: ssl}

	r.leaseBinding, err = NewMikrotikData(a.dial, host, ssl, user, pass, "/ip/dhcp-server/lease")
	if err != nil {
		r.err = err
	}

	return r
}

func (a *appData) routerIdentity(r *router) (sprintf binding.String, err error) {
	var b *MikrotikDataTable
	b, err = NewMikrotikData(a.dial, r.host, r.ssl, r.user, r.password, "/system/routerboard")
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			b.Close()
		}
	}()

	var dataItem binding.DataItem
	dataItem, err = b.GetItem(0)
	if err != nil {
		return
	}

	var currentBoard *MikrotikDataItem
	currentBoard, ok := dataItem.(*MikrotikDataItem)
	if !ok {
		err = fmt.Errorf("invalid data item type")
		return
	}

	var model binding.String
	model, err = currentBoard.Get("model")
	if err != nil {
		return
	}

	var serialNumber binding.String
	serialNumber, err = currentBoard.Get("serial-number")
	if err != nil {
		return
	}

	var upgradeFirmware binding.String
	upgradeFirmware, err = currentBoard.Get("upgrade-firmware")
	if err != nil {
		return
	}

	a.bindings = append(a.bindings, b)

	sprintf = binding.NewSprintf("Router %s (%s) %s", model, serialNumber, upgradeFirmware)
	return
}
