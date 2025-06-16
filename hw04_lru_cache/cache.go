package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheItem struct {
	Key   Key
	Value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := c.items[key]; ok {
		item.Value = cacheItem{key, value}
		c.queue.MoveToFront(item)
		return true
	}
	newItem := c.queue.PushFront(cacheItem{key, value})
	c.items[key] = newItem
	if len(c.items) > c.capacity {
		oldest := c.queue.Back()
		if oldest != nil {
			delete(c.items, oldest.Value.(cacheItem).Key)
			c.queue.Remove(oldest)
		}
	}
	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.queue.MoveToFront(item)
	return item.Value.(cacheItem).Value, ok
}

func (c *lruCache) Clear() {
	c.items = make(map[Key]*ListItem, c.capacity)
	c.queue = NewList()
}
