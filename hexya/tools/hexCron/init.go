package hexCron

import (
	"fmt"
	"time"
)

var (
	CoreWorker Worker
)

func BootStrap() {
	CoreWorker = RegisterWorker(NewInternWorker(&InternWorker{
		name:       "CoreWorker",
		pauseTime:  1 * time.Second,
		maxThreads: 8,
	}))
	scheduler := CoreWorker.NewScheduler("CoreWorkerMainScheduler")
	scheduler.CreateAndScheduleEntry(EntryDesc{
		Name:  "MonitorDatabasWorkerQueue",
		Every: 1 * time.Minute,
		Do:    func() { fmt.Println("CoreWorker") },
	})
}
