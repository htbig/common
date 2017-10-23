package tasks

const (
	WAITING   State = "waiting"
	RUNNING         = "running"
	STOPPING        = "stopping"
	COMPLETED       = "completed"
	FAILED          = "failed"
	STOPPED         = "stopped"
)

type State string

type Pipe struct {
	Progress float32
	Data     interface{}
}

type Task struct {
	id          string
	run         func(chan Pipe, chan struct{}) error
	progress    float32
	data        interface{}
	state       State
	err         error
	stop        chan struct{}
	Description string
}

type Status struct {
	State    State   `json:"state"`
	Progress float32 `json:"progress"`
	Error    string  `json:"error"`
}

func (t Task) Status() Status {
	s := Status{}

	s.Progress = t.progress
	s.State = t.state
	if t.err != nil {
		s.Error = t.err.Error()
	}

	return s
}

func (t Task) Map() map[string]interface{} {
	m := make(map[string]interface{})
	m["progress"] = t.progress
	m["state"] = string(t.state)
	if t.err != nil {
		m["error"] = t.err.Error()
	}
	return m
}

func (t Task) MapWithData(dataKey string) map[string]interface{} {
	m := make(map[string]interface{})
	m["progress"] = t.progress
	m["state"] = string(t.state)
	if t.err != nil {
		m["error"] = t.err.Error()
	}
	if t.data != nil {
		m[dataKey] = t.data
	}
	return m
}

func (t Task) Data() interface{} {
	return t.data
}

func (t Task) ID() string {
	return t.id
}

func (t Task) Error() error {
	return t.err
}

func (t Task) IsCompleted() bool {
	return t.state == COMPLETED
}

func (t Task) IsDone() bool {
	switch t.state {
	case COMPLETED, STOPPED, FAILED:
		return true
	default:
		return false
	}
}

func (t *Task) SetRun(runner func(chan Pipe, chan struct{}) error) {
	t.run = runner
}

func (t *Task) Stop() {
	select {
	case t.stop <- struct{}{}:
	default:
	}
	t.state = STOPPING
}

func (t *Task) State() string {
	return string(t.state)
}

func (t *Task) Start() {
	go func() {
		t.progress = 0
		t.data = nil
		t.state = RUNNING

		pc := make(chan Pipe)
		t.stop = make(chan struct{}, 1)

		go func() {
			defer close(pc)
			t.err = t.run(pc, t.stop)
		}()

		for p, more := <-pc; more; p, more = <-pc {
			t.progress = p.Progress
			t.data = p.Data
		}

		if t.state == STOPPING {
			t.state = STOPPED
		} else {
			if t.err != nil {
				t.state = FAILED
			} else {
				t.state = COMPLETED
			}
		}

	}()
}
