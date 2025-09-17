package go_pool

import "sync"

type ITask interface {
	Run() any
}
type IExecutor interface {
	Exec(task ITask)
}
type IGoPool interface {
	Start()
	Schedule(task ITask)
	WaitAndClose()
}
type gPool struct {
	executors []IExecutor
	workers   int
	tasks     chan ITask
	wg        sync.WaitGroup
}

func NewPool(workers int, executor ...IExecutor) IGoPool {
	if len(executor) != 0 {
		workers = len(executor)
	}
	if workers <= 0 {
		workers = 1
	}
	p := &gPool{
		workers:   workers,
		tasks:     make(chan ITask, workers*2),
		executors: append([]IExecutor{}, executor...),
	}
	return p
}
func (p *gPool) Start() {
	useExecutor := false
	if len(p.executors) > 0 {
		useExecutor = true
	}

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		var e IExecutor
		if useExecutor {
			e = p.executors[i]
		}
		go func(executor IExecutor) {
			defer p.wg.Done()
			for task := range p.tasks {
				if e != nil {
					e.Exec(task)
				} else {
					task.Run()
				}
			}
		}(e)
	}
}
func (p *gPool) Schedule(task ITask) {
	p.tasks <- task
}
func (p *gPool) WaitAndClose() {
	close(p.tasks)
	p.wg.Wait()
}
