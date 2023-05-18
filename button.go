package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type Button struct {
	widget.Button

	dataListener binding.DataListener
	data         binding.String

	disableListener binding.DataListener
	disable         binding.Bool
}

var _ fyne.Widget = (*Button)(nil)

func NewButton(text string, f func()) *Button {
	r := Button{Button: widget.Button{Text: text, OnTapped: f}}
	r.ExtendBaseWidget(&r)
	return &r
}

func (b *Button) Unbind() {
	if b.dataListener == nil {
		return
	}
	b.data.RemoveListener(b.dataListener)
	b.dataListener = nil
	b.Text = ""
}

func (b *Button) Bind(s binding.String) {
	if b.data == s {
		goto refreshText
	}
	if b.dataListener != nil {
		b.Unbind()
	}
	b.data = s
	b.dataListener = binding.NewDataListener(func() {
		b.SetText(getString(s))
	})
	s.AddListener(b.dataListener)
refreshText:
	b.SetText(getString(s))
}

func (b *Button) UnbindDisable() {
	if b.disableListener == nil {
		return
	}
	b.disable.RemoveListener(b.disableListener)
	b.disableListener = nil
}

func (b *Button) BindDisable(data binding.Bool) {
	applyDisable := func(disable bool) {
		if disable {
			b.Disable()
		} else {
			b.Enable()
		}
	}

	if b.disable == data {
		goto refreshDisable
	}
	if b.disableListener != nil {
		b.UnbindDisable()
	}

	b.disableListener = binding.NewDataListener(func() {
		isDisable := b.Disabled()
		disable := getBool(data)
		if isDisable == disable {
			return
		}
		applyDisable(disable)
	})
	b.disable = data
	b.disable.AddListener(b.disableListener)

refreshDisable:
	disable := getBool(data)
	applyDisable(disable)
}

func getString(data binding.String) string {
	if data == nil {
		return ""
	}

	v, err := data.Get()
	if err != nil {
		return ""
	}
	return v
}

func getBool(data binding.Bool) bool {
	if data == nil {
		return false
	}

	v, err := data.Get()
	if err != nil {
		return false
	}
	return v
}
