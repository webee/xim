package rpcservice

var (
	works chan func()
)

// InitWorker init workers.
func InitWorker(count int) {
	works = make(chan func(), count*128)
	for i := 0; i < count; i++ {
		go worker(works)
	}
}

func worker(works <-chan func()) {
	for work := range works {
		work()
	}
}
