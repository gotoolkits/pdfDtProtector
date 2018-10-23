package index

import (
	"container/list"
)

//建立内存缓存库
type ContentCacheTable struct {
	Cache *list.List
}

func NewContentCacheable() ContentCacheTable {
	return ContentCacheTable{
		Cache: list.New(),
	}
}

type CharacterIndexTable struct {
	IndexTable map[string][]*list.Element
	Len        int
}

//为每一个字符建立内存索引库
func NewCharacterIndexTable() CharacterIndexTable {
	return CharacterIndexTable{
		IndexTable: make(map[string][]*list.Element),
	}
}

func (ct *CharacterIndexTable) ItemPush(k string, v *list.Element) {
	ct.IndexTable[k] = append(ct.IndexTable[k], v)
}
func (ct *CharacterIndexTable) ItemPop(k string) (elem []*list.Element) {
	return ct.IndexTable[k]
}
func (ct *CharacterIndexTable) Exsit(k string) (ok bool) {
	_, ok = ct.IndexTable[k]
	return
}
