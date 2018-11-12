package hexCron

import (
	"time"

	"reflect"

	"runtime"

	"github.com/hexya-erp/hexya/hexya/tools/strutils"
)

type InternWorker struct {
	name        string
	queue       []*Job
	pauseTime   time.Duration
	maxThreads  int
	threadschan chan bool
}

func NewInternWorker(worker *InternWorker) Worker {
	out := &InternWorker{
		name:       strutils.MakeUnique(worker.name, workers.names),
		pauseTime:  500 * time.Millisecond,
		maxThreads: 0,
	}
	if worker.pauseTime != 0 {
		out.pauseTime = worker.pauseTime
	}
	if worker.maxThreads != 0 {
		out.maxThreads = worker.maxThreads
	}
	out.threadschan = make(chan bool, out.maxThreads)
	return out
}

func (w *InternWorker) NewScheduler(name string) *Scheduler {
	out := Scheduler{
		name:         strutils.MakeUnique(name, schedulers.names),
		entries:      make(map[string]*entry),
		updateChan:   make(chan bool),
		killChan:     make(chan bool),
		parentWorker: w,
	}
	schedulers.schedulers = append(schedulers.schedulers, &out)
	schedulers.names = append(schedulers.names, out.name)
	go out.schedulerLoop(time.Hour)
	return &out
}

func (w *InternWorker) Start() {
	go w.workerLoop(time.Hour)
	for i := 0; i < w.maxThreads; i++ {
		w.threadschan <- true
	}
}

func (w *InternWorker) Name() string {
	return w.name
}

func (w *InternWorker) AddThreads(ammount int) {
	for i := 0; i < ammount; i++ {
		w.threadschan <- true
	}
}

func (w *InternWorker) GetQueue() *[]*Job {
	return &w.queue
}

func (w *InternWorker) Run(keepMeNil string, fnc interface{}, paras ...interface{}) {
	typ := reflect.TypeOf(fnc)
	if typ.Kind() != reflect.Func {
		panic("only a function can be scheduled into the entry queue.")
	}
	f := reflect.ValueOf(fnc)
	params := paras
	if len(params) != f.Type().NumIn() {
		panic("the number of param is not adapted")
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	job := &Job{
		name:         runtime.FuncForPC(f.Pointer()).Name(),
		f:            f,
		in:           in,
		parentWorker: w,
	}
	w.PushToWorker(job)
}

func (w *InternWorker) PushToWorker(job *Job) {
	queue := w.GetQueue()
	*queue = append(*queue, job)
}

func (w *InternWorker) GetQueueSize() int64 {
	return int64(len(w.queue))
}

func (w *InternWorker) workerLoop(next time.Duration) {
	for {
		if len(w.queue) > 0 {
			<-w.threadschan
			w.queue[0].execute()
			w.queue = w.queue[1:]
		} else {
			time.Sleep(w.pauseTime)
		}
	}
}
