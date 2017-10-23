package locker

type Lock struct {
	channel chan int
	client  string
}

func (lock *Lock) TryLock(lockerClient string) (bool, string) {
	select {
	case lock.channel <- 0:
		lock.client = lockerClient
		return true, lockerClient
	default:
		return false, lock.client
	}
}

func (lock *Lock) Unlock() {
	select {
	case <-lock.channel:
		lock.client = ""
	default:
	}

}

func (lock *Lock) Try(lockerClient string, f func()) (bool, string) {
	if ok, client := lock.TryLock(lockerClient); ok {
		defer lock.Unlock()
		f()
		return true, client
	} else {
		return false, client
	}
}

func New() *Lock {
	lock := new(Lock)

	lock.channel = make(chan int, 1)
	lock.Unlock()

	return lock
}
