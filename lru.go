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
	key, val string
}

func (u *lru) Add(key, value string) {
	if ee, ok := u.cache[key]; ok {
		u.ll.MoveToFront(ee)
		ee.Value.(*entry).val = value
		return
	}

	ele := u.ll.PushFront(&entry{key, value})
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

func (u *lru) Get(key string) (value string, ok bool) {
	var ele *list.Element
	if ele, ok = u.cache[key]; ok {
		value = ele.Value.(*entry).val
		return
	}
	return
}
