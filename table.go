package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *appData) NewTableWithDataColumn(column []RouterOSHeader, data *MikrotikDataTable) *widget.Table {
	t := widget.NewTable(func() (int, int) {
		return data.Length(), len(column)
	}, func() fyne.CanvasObject {
		var button *Button
		button = NewButton("MAC Address", a.lookupIP(button))
		button.Hide()
		button.Importance = widget.LowImportance

		return container.NewStack(
			widget.NewLabel("Not connected yet place holder"),
			button,
		)
	}, func(i widget.TableCellID, o fyne.CanvasObject) {
		label := o.(*fyne.Container).Objects[0].(*widget.Label)
		button := o.(*fyne.Container).Objects[1].(*Button)

		label.Unbind()
		button.Unbind()

		row, err := data.GetItem(i.Row)
		if err != nil {
			button.Hide()
			label.Show()
			label.SetText("")
			return
		}
		col, err := row.Get(column[i.Col].path)
		if err != nil {
			button.Hide()
			label.Show()
			label.SetText("")
			return
		}

		if column[i.Col].mac {
			label.Hide()
			button.Show()
			button.Bind(col)
			var exist []binding.Bool

			for _, router := range a.routers {
				if router.leaseBinding == nil {
					continue
				}

				exist = append(exist, router.leaseBinding.Exist("mac-address", button.Text))
			}
			button.Icon = nil
			button.OnTapped = a.lookupIP(button)
			button.BindDisable(NewNot(NewOr(exist...)))
		} else if column[i.Col].copy {
			button.Icon = theme.ContentCopyIcon()
			button.OnTapped = a.copy(button)
			button.Bind(col)
			button.Enable()
			button.Show()
			label.Hide()
		} else {
			button.Hide()
			label.Show()
			label.Bind(col)
			label.Wrapping = fyne.TextTruncate
		}
	})

	t.ShowHeaderRow = true
	t.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		template.(*widget.Label).SetText(column[id.Col].title)
	}
	data.AddListener(binding.NewDataListener(func() {
		t.Refresh()
	}))

	return t
}

func (a *appData) lookupIP(button *Button) func() {
	return func() {
		dl := []binding.DataList{}

		for name, router := range a.routers {
			if router.leaseBinding == nil {
				continue
			}
			lookup := router.leaseBinding.Search("mac-address", button.Text)
			if lookup == nil {
				log.Println("no lease in", name)
				continue
			}

			dl = append(dl, lookup)
		}

		merged := NewMergeDataList(dl)

		list := widget.NewListWithData(merged, func() fyne.CanvasObject {
			return widget.NewLabel("hostname.somewhere.com (255.255.255.255)")
		}, func(di binding.DataItem, co fyne.CanvasObject) {
			lookup := di.(*MikrotikDataItem)
			label := co.(*widget.Label)

			ipString, _ := lookup.Get("active-address")
			hostnameString, _ := lookup.Get("host-name")

			if len(getString(hostnameString)) > 0 {
				if len(getString(ipString)) > 0 {
					label.Bind(binding.NewSprintf("%s (%s)\n", hostnameString, ipString))
				} else {
					label.Bind(binding.NewSprintf("%s (-)\n", hostnameString))
				}
			} else if len(getString(ipString)) > 0 {
				label.Bind(binding.NewSprintf("%s\n", ipString))
			} else {
				label.Unbind()
				label.SetText("unknown")
			}

		})

		d := dialog.NewCustom("Matching information for "+button.Text, "OK", container.New(&moreSpace{a.win}, list), a.win)
		d.SetOnClosed(func() {
			merged.Close()
			d.Hide()
		})
		d.Show()
	}
}

func (a *appData) copy(button *Button) func() {
	return func() {
		a.win.Clipboard().SetContent(button.Text)
	}
}
