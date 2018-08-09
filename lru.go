package gontpd

import (
	"net"
)

type lru struct {
	cache    map[string]*entry
	root     *entry
	maxEntry int
	len      int
}

func newLRU(s int) *lru {

	root := &entry{prev: nil, next: nil}
	root.next = root
	root.prev = root

	return &lru{
		make(map[string]*entry, s),
		root,
		s, 0,
	}
}

type entry struct {
	key        net.IP
	lastUnix   int64
	prev, next *entry
}

func (u *lru) Add(ip net.IP, val int64) {

	var (
		e  *entry
		ok bool
	)

	if e, ok = u.cache[string(ip)]; ok {
		// move to front
		e.lastUnix = val

		if u.root.next == e {
			return
		}
		// unlink target
		prev := e.prev
		next := e.next
		prev.next = next
		next.prev = prev

		u.insertHead(e)
		return
	}

	if u.len >= u.maxEntry {
		// remove tail
		e = u.root.prev
		delete(u.cache, string(e.key))

		u.root.prev = e.prev
		e.key = ip
		e.lastUnix = val
		u.len--
	} else {
		// not enough
		e = &entry{key: ip, lastUnix: val}
	}
	u.insertHead(e)
	u.cache[string(ip)] = e
	u.len++
}

func (u *lru) insertHead(e *entry) {
	originHead := u.root.next
	u.root.next = e
	originHead.prev = e
	e.next = originHead
}

func (u *lru) Get(ip net.IP) (val int64, ok bool) {
	var e *entry
	e, ok = u.cache[string(ip)]
	if ok {
		val = e.lastUnix
	}
	return
}
