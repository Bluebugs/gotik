package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"fyne.io/fyne/v2/data/binding"
	"github.com/go-routeros/routeros"
	"github.com/go-routeros/routeros/proto"
)

type MikrotikDataItem struct {
	id string

	properties map[string]binding.String

	listeners sync.Map
}

type MikrotikDataTable struct {
	listeners sync.Map

	cancel context.CancelFunc

	items     map[string]*MikrotikDataItem
	itemsList []*MikrotikDataItem
}

func NewMikrotikData(dial func(ctx context.Context, network, address string) (net.Conn, error),
	host string, ssl bool, user, password, path string) (*MikrotikDataTable, error) {
	port := 8728
	if ssl {
		port = 8729
	}

	rawConn, err := dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	if ssl {
		rawConn = tls.Client(rawConn, &tls.Config{InsecureSkipVerify: true})
	}

	client, err := routeros.NewClient(rawConn)
	if err != nil {
		rawConn.Close()
		return nil, err
	}
	err = client.Login(user, password)
	if err != nil {
		client.Close()
		return nil, err
	}

	r, err := client.RunArgs([]string{path + "/print"})
	if err != nil {
		return nil, err
	}

	m := &MikrotikDataTable{items: map[string]*MikrotikDataItem{}}

	listen := false

	for _, s := range r.Re {
		item := newMikrotikDataItem(s)
		m.items[item.id] = item
		m.itemsList = append(m.itemsList, item)
		if item.id != "" {
			listen = true
		}
		log.Println(item)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	if listen {
		go func() {
			l, err := client.ListenArgs([]string{path + "/listen"})
			if err != nil {
				return
			}

			for {
				select {
				case s := <-l.Chan():
					id := getID(s)
					item, ok := m.items[id]
					if !ok {
						if id != "" {
							item = newMikrotikDataItem(s)
							m.items[id] = item
							m.itemsList = append(m.itemsList, item)
							m.listeners.Range(func(key, value interface{}) bool {
								key.(binding.DataListener).DataChanged()
								return true
							})
						}
						continue
					}

					if s != nil && s.List != nil {
						for _, p := range s.List {
							if p.Key == ".id" {
								continue
							}
							_, ok := item.properties[p.Key]
							if !ok {
								item.properties[p.Key] = binding.NewString()
							}
							item.properties[p.Key].Set(p.Value)
							m.listeners.Range(func(key, value interface{}) bool {
								key.(binding.DataListener).DataChanged()
								return true
							})
						}
					}
				case <-ctx.Done():
					client.Close()
					return
				}
			}
		}()
	}

	return m, nil
}

func (m *MikrotikDataTable) Close() {
	m.listeners = sync.Map{}
	m.cancel()
}

func (m *MikrotikDataTable) Search(property, value string) ([]*MikrotikDataItem, error) {
	var items []*MikrotikDataItem

	for _, item := range m.items {
		if v, ok := item.properties[property]; ok {
			s, err := v.Get()
			if err != nil {
				continue
			}

			if s == value {
				items = append(items, item)
			}
		}
	}
	if len(items) == 0 {
		return nil, errors.New("not found")
	}
	return items, nil
}

func (m *MikrotikDataTable) Exist(property, value string) binding.Bool {
	return &MikrotikExist{property: property, value: value, m: m}
}

func (m *MikrotikDataTable) Get(key string) (*MikrotikDataItem, error) {
	if _, ok := m.items[key]; !ok {
		return nil, errors.New("key not found")
	}
	return m.items[key], nil
}

func (m *MikrotikDataTable) GetItem(index int) (*MikrotikDataItem, error) {
	if index < 0 || index >= len(m.items) {
		return nil, errors.New("index out of bounds")
	}
	return m.itemsList[index], nil
}

func (m *MikrotikDataTable) Length() int {
	return len(m.items)
}

func (m *MikrotikDataTable) AddListener(l binding.DataListener) {
	m.listeners.Store(l, true)
	go l.DataChanged()
}

func (m *MikrotikDataTable) RemoveListener(l binding.DataListener) {
	m.listeners.Delete(l)
}

func (m *MikrotikDataItem) Get(key string) (binding.String, error) {
	if b, ok := m.properties[key]; ok {
		return b, nil
	}
	return nil, errors.New("key not found")
}

func (m *MikrotikDataItem) GetValue(key string) (string, error) {
	if p, ok := m.properties[key]; ok {
		return p.Get()
	}
	return "", errors.New("key not found")
}

func (m *MikrotikDataItem) AddListener(l binding.DataListener) {
	m.listeners.Store(l, true)
	go l.DataChanged()
}

func (m *MikrotikDataItem) RemoveListener(l binding.DataListener) {
	m.listeners.Delete(l)
}

func getID(r *proto.Sentence) string {
	if r == nil || r.List == nil {
		return ""
	}

	for _, p := range r.List {
		if p.Key == ".id" {
			return p.Value
		}
	}
	return ""
}

func newMikrotikDataItem(r *proto.Sentence) *MikrotikDataItem {
	item := &MikrotikDataItem{properties: map[string]binding.String{}}
	for _, p := range r.List {
		if p.Key == ".id" {
			item.id = p.Value
		} else {
			item.properties[p.Key] = binding.NewString()
			item.properties[p.Key].Set(p.Value)
		}
	}
	return item
}

type MikrotikExist struct {
	property, value string

	m *MikrotikDataTable
}

var _ binding.Bool = (*MikrotikExist)(nil)

func (b *MikrotikExist) Get() (bool, error) {
	for _, item := range b.m.items {
		if v, ok := item.properties[b.property]; ok {
			s, err := v.Get()
			if err != nil {
				continue
			}

			if s == b.value {
				return true, nil
			}
		}
	}
	return false, nil
}

func (b *MikrotikExist) Set(v bool) error {
	return errors.New("not supported")
}

func (b *MikrotikExist) AddListener(l binding.DataListener) {
	b.m.AddListener(l)
}

func (b *MikrotikExist) RemoveListener(l binding.DataListener) {
	b.m.RemoveListener(l)
}
