package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func NewTableWithDataColumn(column []RouterOSHeader, data *MikrotikDataTable) *widget.Table {
	t := widget.NewTable(func() (int, int) {
		return data.Length(), len(column)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("Not connected yet place holder")
	}, func(i widget.TableCellID, o fyne.CanvasObject) {
		o.(*widget.Label).Unbind()

		row, err := data.GetItem(i.Row)
		if err != nil {
			o.(*widget.Label).SetText("")
			return
		}
		col, err := row.Get(column[i.Col].path)
		if err != nil {
			o.(*widget.Label).SetText("")
			return
		}
		o.(*widget.Label).Bind(col)
	})

	t.ShowHeaderRow = true
	t.StickyRowCount = 1
	t.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		template.(*widget.Label).SetText(column[id.Col].title)
	}

	return t
}
