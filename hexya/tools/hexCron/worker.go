package hexCron

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/hexya-erp/hexya/hexya/models"
)

type Worker interface {
	Name() string
	Start()
	NewScheduler(string) *Scheduler
	AddThreads(int)
	GetQueue() *[]*Job
	GetQueueSize() int64
	PushToWorker(*Job)
	Run(string, interface{}, ...interface{})
}

type workerList struct {
	workers []*Worker
	names   []string
}

var workers workerList

type JobPreArgs struct {
	WorkerName string
	RecCol     *models.RecordCollection
	Params     []interface{}
}

func (j JobPreArgs) WithParams(params ...interface{}) JobPreArgs {
	j.Params = params
	return j
}

func (j JobPreArgs) Run(method *models.Method) {
	(*Get(j.WorkerName)).Run(j.RecCol.ModelName(), method, append([]interface{}{j.RecCol}, j.Params...)...)
}

type Job struct {
	parentEntry  *entry
	parentWorker Worker
	name         string
	modelName    string
	f            reflect.Value
	in           []reflect.Value
	inRaw        []interface{}
}

func Get(name string) *Worker {
	for i, n := range workers.names {
		if n == name {
			return workers.workers[i]
		}
	}
	return nil
}

func RegisterWorker(w Worker) Worker {
	workers.workers = append(workers.workers, &w)
	workers.names = append(workers.names, w.Name())
	w.Start()
	return w
}

func (j *Job) execute() {
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			if j.parentEntry != nil {
				j.parentEntry.lastReturn = j.f.Call(j.in)
			} else {
				j.f.Call(j.in)
			}
			wg.Done()
		}()
		wg.Wait()
		j.parentWorker.AddThreads(1)
	}()
}

func (j *Job) toName() string {
	out := j.parentWorker.Name() + "_" + filepath.Base(j.name) + fmt.Sprintf("%+v", j.in)
	return out
}
