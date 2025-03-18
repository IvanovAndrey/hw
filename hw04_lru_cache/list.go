package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	first *ListItem
	last  *ListItem
}

func NewList() List {
	return new(list)
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	if l.len == 0 {
		return nil
	}
	return l.first
}

func (l *list) Back() *ListItem {
	if l.len == 0 {
		return nil
	}
	return l.last
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: nil, Prev: nil}
	if l.len == 0 {
		l.first = item
		l.last = item
	} else {
		item.Next = l.first
		l.first.Prev = item
		l.first = item
	}
	l.len++
	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: nil, Prev: nil}
	if l.len == 0 {
		l.last = item
		l.first = item
	} else {
		item.Prev = l.last
		l.last.Next = item
		l.last = item
	}
	l.len++
	return item
}
func (l *list) Remove(i *ListItem) {
	if i == nil {
		return
	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.first = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.last = i.Prev
	}
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || i == l.first {
		return
	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.last = i.Prev
	}
	i.Prev = nil
	l.first.Prev = i
	i.Next = l.first
	l.first = i
}
