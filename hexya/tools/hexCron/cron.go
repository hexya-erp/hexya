package hexCron

import (
	"time"

	"reflect"

	"fmt"
)

type schedulerList struct {
	schedulers []*Scheduler
	names      []string
}

type Scheduler struct {
	isRunning    bool
	name         string
	entries      map[string]*entry
	schedule     []*entry
	updateChan   chan bool
	killChan     chan bool
	parentWorker Worker
}

type entry struct {
	parentScheduler *Scheduler
	name            string
	nextCall        time.Time
	interval        time.Duration
	f               interface{}
	params          []interface{}
	quitChan        chan bool
	lastCall        time.Time
	callAmmount     int
	lastReturn      interface{}
	returns         []interface{}
}

type EntryDesc struct {
	Name   string
	In     time.Duration
	At     string
	Every  time.Duration
	Do     interface{}
	Params []interface{}
}

var schedulers schedulerList

func (sch *Scheduler) schedulerLoop(next time.Duration) {
	sch.isRunning = true
	for {
		select {
		case <-time.After(next):
			if len(sch.schedule) > 0 {
				sch.schedule[0].Run()
				sch.schedule = sch.schedule[1:]
			}
			if len(sch.schedule) > 0 {
				next = sch.schedule[0].getTimeUntilNext()
			} else {
				next = time.Hour
			}
		case <-sch.updateChan:
			if len(sch.schedule) > 0 {
				next = sch.schedule[0].getTimeUntilNext()
			} else {
				next = time.Hour
			}
		case <-sch.killChan:
			sch.isRunning = false
			return
		}
	}
}

func (sch *Scheduler) Kill() {
	if sch.isRunning {
		sch.killChan <- true
	}
}

func (sch *Scheduler) CreateAndScheduleEntry(desc EntryDesc) *entry {
	out := sch.CreateEntry(desc)
	out.Schedule()
	return out
}

func (sch *Scheduler) CreateEntry(desc EntryDesc) *entry {
	typ := reflect.TypeOf(desc.Do)
	if typ.Kind() != reflect.Func {
		panic("only a function can be scheduled into the entry queue.")
	}
	var nextCall time.Time
	switch {
	case desc.In != 0:
		nextCall = time.Now().Add(desc.In)
	case desc.At != "":
		nextCall = readNextCallAt(desc.At)

	}

	out := &entry{
		parentScheduler: sch,
		name:            desc.Name,
		nextCall:        nextCall,
		interval:        desc.Every,
		f:               desc.Do,
		params:          desc.Params,
		quitChan:        make(chan bool),
	}
	sch.entries[desc.Name] = out
	return out
}

func (e *entry) Schedule() {
	if len(e.parentScheduler.schedule) == 0 {
		e.parentScheduler.schedule = []*entry{e}
		e.parentScheduler.updateChan <- true
		return
	}
	for i, en := range e.parentScheduler.schedule {
		if en.nextCall.After(e.nextCall) {
			e.parentScheduler.schedule = append(e.parentScheduler.schedule[:i], append([]*entry{e}, e.parentScheduler.schedule[i+1:]...)...)
			e.parentScheduler.updateChan <- true
			return
		}
	}
	e.parentScheduler.schedule = append(e.parentScheduler.schedule, e)
}

func (e *entry) reSchedule() {
	e.nextCall = time.Now().Add(e.interval)
	e.Schedule()
}

func (e *entry) getTimeUntilNext() time.Duration {
	return e.nextCall.Sub(time.Now())
}

func (e *entry) Run() {
	f := reflect.ValueOf(e.f)
	params := e.params
	if len(params) != f.Type().NumIn() {
		panic("the number of param is not adapted")
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	job := &Job{
		parentEntry:  e,
		parentWorker: e.parentScheduler.parentWorker,
		name:         e.name,
		f:            f,
		in:           in,
	}
	e.parentScheduler.parentWorker.PushToWorker(job)
	e.reSchedule()
}

func readNextCallAt(at string) time.Time {
	tim, err := time.Parse("15:04", at)
	if err != nil {
		fmt.Println("scheduler: 'At' format is wrong. assuming its '00:00'")
		tim, _ = time.Parse("15:04", "00:00")
	}
	now := time.Now()
	out := time.Date(now.Year(), now.Month(), now.Day(), tim.Hour(), tim.Minute(), 0, 0, now.Location())
	if out.Before(now) {
		out.Add(24 * time.Hour)
	}
	return out
}
