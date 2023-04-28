package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AdaptativeSplit struct {
	widget.BaseWidget

	menu       *widget.Button
	title      *widget.Label
	separator  fyne.CanvasObject
	background fyne.CanvasObject

	leftContent  fyne.CanvasObject
	rightContent fyne.CanvasObject

	collapsed bool
}

var _ fyne.Widget = (*AdaptativeSplit)(nil)

func NewSplit(title string, leftContent fyne.CanvasObject, rightContent fyne.CanvasObject) *AdaptativeSplit {
	var l *AdaptativeSplit
	var menu *widget.Button
	menu = widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		if l.collapsed {
			l.collapsed = false
			menu.Icon = theme.CancelIcon()
		} else {
			l.collapsed = true
			menu.Icon = theme.MenuIcon()
		}
		menu.Refresh()
		l.Refresh()
	})
	l = &AdaptativeSplit{
		menu:         menu,
		title:        widget.NewLabel(title),
		separator:    widget.NewSeparator(),
		background:   canvas.NewRectangle(theme.MenuBackgroundColor()),
		leftContent:  leftContent,
		rightContent: rightContent,
	}
	l.title.TextStyle.Bold = true
	l.title.TextStyle.Italic = true
	l.BaseWidget.ExtendBaseWidget(l)
	return l
}

func (s *AdaptativeSplit) CreateRenderer() fyne.WidgetRenderer {
	return &AdaptativeSplitRenderer{
		split: s,
	}
}

func (s *AdaptativeSplit) MinSize() fyne.Size {
	menuMinSize := s.menu.MinSize()
	leftMinSize := s.leftContent.MinSize()
	rightMinSize := s.rightContent.MinSize()
	separatorMinSize := s.separator.MinSize()

	width := fyne.Max(menuMinSize.Width+rightMinSize.Width+separatorMinSize.Width+theme.Padding()*2, leftMinSize.Width+separatorMinSize.Width+theme.Padding())
	height := fyne.Max(menuMinSize.Height+leftMinSize.Height, rightMinSize.Height)

	return fyne.NewSize(width, height)
}

type AdaptativeSplitRenderer struct {
	split *AdaptativeSplit
}

func (r *AdaptativeSplitRenderer) Layout(size fyne.Size) {
	s := r.split

	s.menu.Move(fyne.NewPos(0, 0))
	s.menu.Resize(s.menu.MinSize())

	if s.collapsed {
		s.title.Hide()

		s.separator.Move(fyne.NewPos(s.menu.Position().X+s.menu.Size().Width+theme.Padding(), 0))
		s.separator.Resize(fyne.NewSize(1, size.Height))

		s.leftContent.Hide()

		s.rightContent.Move(fyne.NewPos(s.separator.Position().X+s.separator.Size().Width+theme.Padding(), 0))
		s.rightContent.Resize(fyne.NewSize(size.Width-s.rightContent.Position().X, size.Height))
	} else {
		s.title.Move(fyne.NewPos(s.menu.Position().X+s.menu.Size().Width+theme.Padding(), 0))
		s.title.Resize(fyne.NewSize(s.leftContent.MinSize().Width-s.title.Position().X, s.menu.Size().Height))
		s.title.Show()

		s.separator.Move(fyne.NewPos(s.leftContent.Position().X+s.leftContent.Size().Width+theme.Padding(), 0))
		s.separator.Resize(fyne.NewSize(1, size.Height))

		s.leftContent.Move(fyne.NewPos(0, s.menu.Position().Y+s.menu.Size().Height+theme.Padding()))
		s.leftContent.Resize(fyne.NewSize(s.leftContent.MinSize().Width, size.Height-s.leftContent.Position().Y))
		s.leftContent.Show()

		rightContentMinSize := s.rightContent.MinSize()
		if rightContentMinSize.Width > size.Width-s.separator.Position().X-s.separator.Size().Width-theme.Padding()*2 {
			s.rightContent.Move(fyne.NewPos(s.menu.Position().X+s.menu.Size().Width+theme.Padding()+s.separator.Size().Width, 0))
			s.rightContent.Resize(fyne.NewSize(size.Width-s.rightContent.Position().X, size.Height))
		} else {
			s.rightContent.Move(fyne.NewPos(s.separator.Position().X+s.separator.Size().Width+theme.Padding(), 0))
			s.rightContent.Resize(fyne.NewSize(size.Width-s.rightContent.Position().X, size.Height))
		}
	}

	s.background.Move(fyne.NewPos(0, 0))
	s.background.Resize(fyne.NewSize(s.separator.Position().X, size.Height))
}

func (r *AdaptativeSplitRenderer) MinSize() fyne.Size {
	return r.split.MinSize()
}

func (r *AdaptativeSplitRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.split.rightContent,
		r.split.background,
		r.split.menu,
		r.split.title,
		r.split.leftContent,
		r.split.separator,
	}
}

func (r *AdaptativeSplitRenderer) Refresh() {
	r.Layout(r.split.Size())
}

func (r *AdaptativeSplitRenderer) Destroy() {
}
