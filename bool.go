package main

import "fyne.io/fyne/v2/data/binding"

type Not struct {
	data binding.Bool
}

var _ binding.Bool = (*Not)(nil)

func NewNot(data binding.Bool) *Not {
	return &Not{data: data}
}

func (n *Not) Get() (bool, error) {
	v, err := n.data.Get()
	return !v, err
}

func (n *Not) Set(value bool) error {
	return n.data.Set(!value)
}

func (n *Not) AddListener(listener binding.DataListener) {
	n.data.AddListener(listener)
}

func (n *Not) RemoveListener(listener binding.DataListener) {
	n.data.RemoveListener(listener)
}

type And struct {
	data []binding.Bool
}

var _ binding.Bool = (*And)(nil)

func NewAnd(data ...binding.Bool) *And {
	return &And{data: data}
}

func (a *And) Get() (bool, error) {
	for _, d := range a.data {
		v, err := d.Get()
		if err != nil {
			return false, err
		}
		if !v {
			return false, nil
		}
	}
	return true, nil
}

func (a *And) Set(value bool) error {
	for _, d := range a.data {
		err := d.Set(value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *And) AddListener(listener binding.DataListener) {
	for _, d := range a.data {
		d.AddListener(listener)
	}
}

func (a *And) RemoveListener(listener binding.DataListener) {
	for _, d := range a.data {
		d.RemoveListener(listener)
	}
}

type Or struct {
	data []binding.Bool
}

var _ binding.Bool = (*Or)(nil)

func NewOr(data ...binding.Bool) *Or {
	return &Or{data: data}
}

func (o *Or) Get() (bool, error) {
	for _, d := range o.data {
		v, err := d.Get()
		if err != nil {
			return false, err
		}
		if v {
			return true, nil
		}
	}
	return false, nil
}

func (o *Or) Set(value bool) error {
	for _, d := range o.data {
		err := d.Set(value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Or) AddListener(listener binding.DataListener) {
	for _, d := range o.data {
		d.AddListener(listener)
	}
}

func (o *Or) RemoveListener(listener binding.DataListener) {
	for _, d := range o.data {
		d.RemoveListener(listener)
	}
}
