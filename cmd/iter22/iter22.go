package main

// Resetter — интерфейс, требующий наличия метода Reset()
type Resetter interface {
	Reset()
}

// Pool — generic‑структура для хранения объектов с методом Reset()
type Pool[T Resetter] struct {
	items []T
}

// New — функция‑конструктор для создания нового пула
func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		items: make([]T, 0),
	}
}

// Get возвращает объект из пула или создаёт новый, если пул пуст
func (p *Pool[T]) Get() T {

	if len(p.items) > 0 {
		item := p.items[len(p.items)-1]
		p.items = p.items[:len(p.items)-1]
		return item
	}

	var zero T
	return zero
}

// Put помещает объект в пул
func (p *Pool[T]) Put(item T) {
	p.items = append(p.items, item)
}
