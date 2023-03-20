package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *appData) createUI() {
	tabs := container.NewAppTabs()

	tree := widget.NewTreeWithStrings(map[string][]string{
		"":       {"CAPsMAN", "Wireless", "Interfaces", "IP", "System"},
		"IP":     {"ARP", "DHCP Server"},
		"System": {"Certificates", "Health"},
	})
	tree.OnSelected = func(id string) {
		fmt.Println("Tree node selected:", id)
		a.buildView(tabs, id)
		a.saveCurrentView()
	}

	header := widget.NewLabel("Not Connected")
	header.Alignment = fyne.TextAlignCenter
	footer := widget.NewLabel("")
	footer.Alignment = fyne.TextAlignCenter
	sel := widget.NewSelect([]string{}, func(s string) {
		for _, b := range a.bindings {
			b.Close()
		}
		a.bindings = []*MikrotikDataTable{}

		r, ok := a.routers[s]
		if !ok {
			header.Unbind()
			header.SetText("Not Connected")
			return
		}

		identity, err := a.routerIdentity(r)
		if err != nil {
			footer.SetText(fmt.Sprintf("%v", err))
			return
		}

		a.current = r

		if a.currentView != "" {
			a.buildView(tabs, a.currentView)
		}

		header.Bind(identity)
		footer.SetText("")
	})

	a.win.SetContent(NewSplit("Gotik", container.NewBorder(container.NewBorder(nil, nil,
		widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() { a.removeHost(sel) }),
		widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() { a.newHost(sel) }),
		sel),
		nil, nil, nil, tree),
		container.NewBorder(header, footer, nil, nil, tabs)))
	a.win.Resize(fyne.NewSize(800, 600))

	if a.salt() != nil {
		a.getPassword(sel)
	}
}

func (a *appData) newHost(sel *widget.Select) {
	host := widget.NewEntry()
	host.PlaceHolder = "127.0.0.1:8728"
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	dialog.ShowForm("New router", "Connect", "Cancel",
		[]*widget.FormItem{
			{Text: "Host", Widget: host},
			{Text: "User", Widget: user},
			{Text: "Password", Widget: pass},
		}, func(confirm bool) {
			if confirm {
				r, err := routerView(host.Text, user.Text, pass.Text)
				if err != nil {
					dialog.ShowError(err, a.win)
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

func (a *appData) getPassword(sel *widget.Select) {
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

				if err := a.loadRouters(sel); err != nil {
					dialog.ShowError(err, a.win)
					return
				}
				if len(sel.Options) > 0 {
					sel.SetSelectedIndex(0)
				}
			}
		}, a.win)
	a.win.Canvas().Focus(password)
}

func (a *appData) buildView(tabs *container.AppTabs, view string) {
	a.currentView = view

	if a.current == nil {
		log.Println("no current router")
		return
	}

	tabs.Items = []*container.TabItem{}

	lookup, ok := routerOSCommands[view]
	if !ok {
		log.Println("no view found for", view)
		return
	}

	for _, cmd := range lookup {
		log.Println("loading", cmd.path)
		b, err := NewMikrotikData(a.current.host, a.current.user, a.current.password, cmd.path)
		if err != nil {
			log.Println("failed to load", cmd.path, err)
			continue
		}
		tabs.Append(container.NewTabItem(cmd.title, NewTableWithDataColumn(cmd.headers, b)))
	}
	tabs.Refresh()
	log.Println("loaded", len(tabs.Items), "tabs for", view)
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

	sel.ClearSelected()
	for i, v := range sel.Options {
		if v == sel.Selected {
			sel.Options = append(sel.Options[:i], sel.Options[i+1:]...)
			if len(sel.Options) > 0 {
				sel.SetSelectedIndex(i)
			}
			break
		}
	}
	sel.Refresh()
}

func routerView(host, user, pass string) (*router, error) {
	var err error
	r := &router{host: host, user: user, password: pass}

	r.leaseBinding, err = NewMikrotikData(host, user, pass, "/ip/dhcp-server/lease")
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (a *appData) routerIdentity(r *router) (sprintf binding.String, err error) {
	var b *MikrotikDataTable
	b, err = NewMikrotikData(r.host, r.user, r.password, "/system/routerboard")
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
