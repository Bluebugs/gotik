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
	"github.com/pjediny/mndp/pkg/mndp"
)

type MikrotikRouter struct {
	mndp.Message

	parent *MikrotikRouterList
}

var _ binding.String = (*MikrotikRouter)(nil)

type MikrotikRouterList struct {
	ch chan *mndp.Message

	routers map[string]*MikrotikRouter
	sorted  []string

	listeners sync.Map
}

var _ binding.DataList = (*MikrotikRouterList)(nil)

func NewMikrotikRouterList() *MikrotikRouterList {
	r := &MikrotikRouterList{ch: make(chan *mndp.Message), routers: map[string]*MikrotikRouter{}}
	listener := mndp.NewListener()
	listener.Listen(r.ch)

	go func() {
		for {
			msg, ok := <-r.ch
			if !ok {
				return
			}
			if msg == nil {
				continue
			}
			_, okv4 := msg.Fields[mndp.TagIPv4Addr]
			_, okv6 := msg.Fields[mndp.TagIPv6Addr]
			if !okv4 && !okv6 {
				continue
			}
			router, ok := r.routers[msg.Src.String()]
			if !ok {
				router = &MikrotikRouter{*msg, r}
				r.routers[msg.Src.String()] = router
				r.sorted = append(r.sorted, msg.Src.String())
			} else {
				router.Message = *msg
			}
			r.listeners.Range(func(key, value interface{}) bool {
				key.(binding.DataListener).DataChanged()
				return true
			})
		}
	}()

	return r
}

func (m *MikrotikRouterList) AddListener(l binding.DataListener) {
	m.listeners.Store(l, true)
	go l.DataChanged()
}

func (m *MikrotikRouterList) RemoveListener(l binding.DataListener) {
	m.listeners.Delete(l)
}

func (m *MikrotikRouterList) GetItem(index int) (binding.DataItem, error) {
	if index < 0 || index >= len(m.sorted) {
		return nil, errors.New("index out of range")
	}
	return m.routers[m.sorted[index]], nil
}

func (m *MikrotikRouterList) Length() int {
	return len(m.routers)
}

func (m *MikrotikRouter) AddListener(l binding.DataListener) {
	m.parent.AddListener(l)
}

func (m *MikrotikRouter) RemoveListener(l binding.DataListener) {
	m.parent.RemoveListener(l)
}

func (m *MikrotikRouter) getValue(tag mndp.TLVTag) string {
	v, ok := m.Fields[tag]
	if !ok {
		return ""
	}
	r := ""
	switch v.Tag {
	case mndp.TagMACAddr:
		r = v.ValAsHardwareAddr().String()
	case mndp.TagIdentity:
		r = v.ValAsString()
	case mndp.TagVersion:
		r = v.ValAsString()
	case mndp.TagPlatform:
		r = v.ValAsString()
	case mndp.TagUptime:
		r = v.ValAsDuration().String()
	case mndp.TagSoftwareID:
		r = v.ValAsString()
	case mndp.TagBoard:
		r = v.ValAsString()
	case mndp.TagUnpack:
		r = v.ValAsHexString()
	case mndp.TagIPv6Addr:
		r = v.ValAsIP().String()
	case mndp.TagInterfaceName:
		r = v.ValAsString()
	case mndp.TagIPv4Addr:
		r = v.ValAsIP().String()
	default:
		r = v.ValAsHexString()
	}
	return r
}

func (m *MikrotikRouter) Get() (string, error) {
	identity := m.getValue(mndp.TagIdentity)
	mac := m.getValue(mndp.TagMACAddr)
	platform := m.getValue(mndp.TagPlatform)
	version := m.getValue(mndp.TagVersion)
	ip := m.getValue(mndp.TagIPv4Addr)
	if ip == "" {
		ip = m.getValue(mndp.TagIPv6Addr)
	}

	return fmt.Sprintf("%s (%s, %s) %s - %s", identity, mac, ip, platform, version), nil
}

func (m *MikrotikRouter) Set(string) error {
	return errors.New("not implemented")
}

func (m *MikrotikRouter) IP() string {
	ip := m.getValue(mndp.TagIPv4Addr)
	if ip == "" {
		ip = m.getValue(mndp.TagIPv6Addr)
	}
	return ip
}

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
