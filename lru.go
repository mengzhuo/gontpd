package gontpd

import (
	"container/list"
	"net"
)

type lru struct {
	cache    map[string]*list.Element
	ll       *list.List
	maxEntry int
}

func newLRU(s int) *lru {
	return &lru{
		map[string]*list.Element{},
		list.New(),
		s}
}

type entry struct {
	key      net.IP
	lastUnix int64
}

func (u *lru) Add(ip net.IP, val int64) {
	if ee, ok := u.cache[string(ip)]; ok {
		u.ll.MoveToFront(ee)
		ee.Value.(*entry).lastUnix = val
		return
	}

	ele := u.ll.PushFront(&entry{ip, val})
	u.cache[string(ip)] = ele
	if u.maxEntry < u.ll.Len() {
		u.RemoveOldest()
	}
}

func (u *lru) RemoveOldest() {
	ele := u.ll.Back()
	ee := ele.Value.(*entry)
	delete(u.cache, string(ee.key))
	u.ll.Remove(ele)
}

func (u *lru) Get(ip net.IP) (val int64, ok bool) {
	var ele *list.Element
	if ele, ok = u.cache[string(ip)]; ok {
		val = ele.Value.(*entry).lastUnix
	}
	return
}
