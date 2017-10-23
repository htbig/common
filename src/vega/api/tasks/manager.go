package tasks

import "strconv"

type Manager struct {
	nextID chan string
	tasks  map[string]*Task
}

func (m *Manager) Clear() {
	m.nextID = make(chan string)
	m.tasks = make(map[string]*Task)
	go func() {
		for id := uint64(0); ; id++ {
			m.nextID <- strconv.FormatUint(id, 10)
		}
	}()
}

func (m *Manager) Get(id string) *Task {
	return m.tasks[id]
}

func (m *Manager) Delete(id string) {
	delete(m.tasks, id)
}

func (m *Manager) New(r func(chan Pipe, chan struct{}) error) *Task {
	t := new(Task)
	t.id = <-m.nextID
	t.run = r
	t.state = WAITING
	m.tasks[t.id] = t
	return t
}

// NewManager returns a new task manager
func NewManager() *Manager {
	m := Manager{}
	m.Clear()
	return &m
}
