package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AdaptativeSplit struct {
	menu       *widget.Button
	separator  fyne.CanvasObject
	background fyne.CanvasObject

	leftContent  fyne.CanvasObject
	rightContent fyne.CanvasObject

	collapsed bool
}

var _ fyne.Layout = (*AdaptativeSplit)(nil)

func NewSplit(leftContent fyne.CanvasObject, rightContent fyne.CanvasObject) *fyne.Container {
	var c *fyne.Container
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
		c.Layout.Layout(c.Objects, c.Size())
	})
	l = &AdaptativeSplit{
		menu:         menu,
		separator:    widget.NewSeparator(),
		background:   canvas.NewRectangle(theme.BackgroundColor()),
		leftContent:  leftContent,
		rightContent: rightContent,
	}
	c = container.New(l, rightContent, l.background, menu, leftContent, l.separator)
	return c
}

func (s *AdaptativeSplit) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	s.menu.Move(fyne.NewPos(0, 0))
	s.menu.Resize(s.menu.MinSize())

	if s.collapsed {
		s.separator.Move(fyne.NewPos(s.menu.Position().X+s.menu.Size().Width+theme.Padding(), 0))
		s.separator.Resize(fyne.NewSize(1, size.Height))

		s.leftContent.Hide()

		s.rightContent.Move(fyne.NewPos(s.separator.Position().X+s.separator.Size().Width+theme.Padding(), 0))
		s.rightContent.Resize(fyne.NewSize(size.Width-s.rightContent.Position().X, size.Height))
	} else {
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

func (s *AdaptativeSplit) MinSize(objects []fyne.CanvasObject) fyne.Size {
	menuMinSize := s.menu.MinSize()
	leftMinSize := s.leftContent.MinSize()
	rightMinSize := s.rightContent.MinSize()
	separatorMinSize := s.separator.MinSize()

	width := fyne.Max(menuMinSize.Width+rightMinSize.Width+separatorMinSize.Width+theme.Padding()*2, leftMinSize.Width+separatorMinSize.Width+theme.Padding())
	height := fyne.Max(menuMinSize.Height+leftMinSize.Height, rightMinSize.Height)

	return fyne.NewSize(width, height)
}
