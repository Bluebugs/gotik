package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (a *appData) NewTableWithDataColumn(column []RouterOSHeader, data *MikrotikDataTable) *widget.Table {
	t := widget.NewTable(func() (int, int) {
		return data.Length(), len(column)
	}, func() fyne.CanvasObject {
		var button *Button
		button = NewButton("MAC Address", func() {
			msg := ""

			for _, router := range a.routers {
				if router.leaseBinding == nil {
					continue
				}
				lookups, err := router.leaseBinding.Search("mac-address", button.Text)
				if err != nil {
					continue
				}

				for _, lookup := range lookups {
					ipString, _ := lookup.GetValue("active-address")
					hostnameString, _ := lookup.GetValue("host-name")

					if len(hostnameString) > 0 {
						if len(ipString) > 0 {
							msg += fmt.Sprintf("%s (%s)\n", hostnameString, ipString)
						} else {
							msg += fmt.Sprintf("%s (-)\n", hostnameString)
						}
					} else {
						msg += fmt.Sprintf("%s\n", ipString)
					}
				}
			}
			if len(msg) == 0 {
				return
			}

			dialog.ShowInformation("Matching IP addresses", msg, a.win)
		})
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
		} else {
			button.Hide()
			label.Show()
			label.Bind(col)
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
