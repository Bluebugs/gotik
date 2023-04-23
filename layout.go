package main

import "fyne.io/fyne/v2"

type moreSpace struct {
	win fyne.Window
}

var _ fyne.Layout = (*moreSpace)(nil)

func (m *moreSpace) Layout(object []fyne.CanvasObject, sz fyne.Size) {
	for _, o := range object {
		o.Resize(sz)
	}
}

func (m *moreSpace) MinSize(_ []fyne.CanvasObject) fyne.Size {
	r := m.win.Canvas().Size()
	r.Width = r.Width * 90 / 100
	r.Height = r.Height * 70 / 100
	return r
}
