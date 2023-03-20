package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func NewTableWithDataColumn(column []RouterOSHeader, data *MikrotikDataTable) *widget.Table {
	return widget.NewTable(func() (int, int) {
		return data.Length() + 1, len(column)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("Not connected yet place holder")
	}, func(i widget.TableCellID, o fyne.CanvasObject) {
		o.(*widget.Label).Unbind()

		if i.Row == 0 {
			o.(*widget.Label).SetText(column[i.Col].title)
			return
		}

		row, err := data.GetItem(i.Row - 1)
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
}
