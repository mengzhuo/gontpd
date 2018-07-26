package gontpd

import "container/list"

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
	key      string
	lastUnix int64
}

func (u *lru) Add(key string, val int64) {
	if ee, ok := u.cache[key]; ok {
		u.ll.MoveToFront(ee)
		ee.Value.(*entry).lastUnix = val
		return
	}

	ele := u.ll.PushFront(&entry{key, val})
	u.cache[key] = ele
	if u.maxEntry < u.ll.Len() {
		u.RemoveOldest()
	}
}

func (u *lru) RemoveOldest() {
	ele := u.ll.Back()
	ee := ele.Value.(*entry)
	delete(u.cache, ee.key)
	u.ll.Remove(ele)
}

func (u *lru) Get(key string) (val int64, ok bool) {
	var ele *list.Element
	if ele, ok = u.cache[key]; ok {
		val = ele.Value.(*entry).lastUnix
	}
	return
}
