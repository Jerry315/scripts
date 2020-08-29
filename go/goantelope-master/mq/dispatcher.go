package mq

import (
	"context"
	"errors"
	"log"
)

var (
	// ErrNoWorker 任务分配器未启动任何 worker
	ErrNoWorker = errors.New("dispatcher has no worker")
)

// Dispatcher 任务分配器
type Dispatcher struct {
	Jobs     chan *Job
	jobCount int

	workers        []*Worker
	workerPreFetch int
	producers      []*Producer

	cancelFunc context.CancelFunc
}

// HandleFunc 任务处理函数类型
type HandleFunc func(*Job)

// ProduceFunc 生成任务的函数类型
type ProduceFunc func() chan *Job

// Job worker 处理的任务类型
type Job struct {
	Payload interface{}
}

// NewDispatcher 创建一个新的任务分配器
func NewDispatcher(workerPreFetch int) *Dispatcher {
	return &Dispatcher{
		Jobs:           make(chan *Job),
		workers:        []*Worker{},
		workerPreFetch: workerPreFetch,
	}
}

// Dispatch 分配任务
func (d *Dispatcher) Dispatch(job *Job) error {
	if len(d.workers) == 0 {
		log.Println("mq: dispatcher has no workers, can not work")
		return ErrNoWorker
	}

	idx := d.jobCount % len(d.workers)
	w := d.workers[idx]
	w.jobs <- job
	d.jobCount++
	return nil
}

// Run 启动任务分配
func (d *Dispatcher) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	d.cancelFunc = cancelFunc
	go d.collect(ctx)
}

// collect 收取任务
func (d *Dispatcher) collect(ctx context.Context) {
	// bloody fatory never stop
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-d.Jobs:
			err := d.Dispatch(job)
			if err != nil {
				log.Printf("mq: dispatch job failed %v\n", err)
			}
		}
	}
}

// Stop 停止任务分配
func (d *Dispatcher) Stop() {
	if d.cancelFunc != nil {
		d.cancelFunc()
	}
	for _, w := range d.workers {
		w.Stop()
	}
	for _, p := range d.producers {
		p.Stop()
	}
}

// GoWorkers 启动多个新的 Worker
func (d *Dispatcher) GoWorkers(handle HandleFunc, num uint) {
	for i := num; i > 0; i-- {
		d.GoWorker(handle)
	}
}

// GoWorker 启动一个新的 Worker
func (d *Dispatcher) GoWorker(handle HandleFunc) {
	if d.workers == nil {
		d.workers = []*Worker{}
	}

	w := &Worker{
		jobs:     make(chan *Job, d.workerPreFetch),
		handle:   handle,
		preFetch: d.workerPreFetch,
	}
	d.workers = append(d.workers, w)
	w.Run()
}

// GoProducers 启动多个新的 Producer
func (d *Dispatcher) GoProducers(
	produceFunc ProduceFunc, produceCancel context.CancelFunc, num uint,
) {
	for i := num; i > 0; i-- {
		d.GoProducer(produceFunc, produceCancel)
	}
}

// GoProducer 启动一个 Produce
func (d *Dispatcher) GoProducer(produceFunc ProduceFunc, produceCancel context.CancelFunc) {
	if d.producers == nil {
		d.producers = []*Producer{}
	}

	p := &Producer{
		jobs:          d.Jobs,
		produceFunc:   produceFunc,
		produceCancel: produceCancel,
	}
	d.producers = append(d.producers, p)
	p.Run()
}

// Worker 任务处理 worker
type Worker struct {
	jobs     chan *Job
	handle   HandleFunc
	preFetch int

	cancelFunc context.CancelFunc
}

// Run 启动 worker
func (w *Worker) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	if w.cancelFunc != nil {
		w.cancelFunc()
	}
	w.cancelFunc = cancelFunc
	go w.do(ctx)
}

// Stop 停止 worker
func (w *Worker) Stop() {
	if w.cancelFunc == nil {
		return
	}
	w.cancelFunc()
}

// do 执行工作
func (w *Worker) do(ctx context.Context) {
	if w.jobs == nil {
		w.jobs = make(chan *Job, w.preFetch)
	}

	// dull boy never stop working
	for {
		select {
		// time to play
		case <-ctx.Done():
			return
		case job := <-w.jobs:
			if job == nil {
				continue
			}
			w.handle(job)
		}
	}
}

// Producer 任务生产者
type Producer struct {
	jobs          chan *Job
	produceFunc   ProduceFunc
	produceCancel context.CancelFunc
	cancelFunc    context.CancelFunc
}

// Run 启动 producer
func (p *Producer) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	p.cancelFunc = cancelFunc
	go p.do(ctx)
}

// Stop 停止 producer
func (p *Producer) Stop() {
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	if p.produceCancel != nil {
		p.produceCancel()
	}
}

// do 执行生产
func (p *Producer) do(ctx context.Context) {
	if p.jobs == nil {
		log.Println("mq: producer has no job chan, should get one from Dispatcher")
		return
	}

	jobChan := p.produceFunc()
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-jobChan:
			p.jobs <- job
		}
	}
}
