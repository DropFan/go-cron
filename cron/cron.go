package cron

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

const (
	// defaultInterval cron任务执行默认间隔，可通过SetInterval()设置
	defaultInterval = 500 * time.Millisecond
)

var (
	// ErrIsRunning ...
	ErrIsRunning = errors.New("cron is running")
)

// Cron ...
type Cron struct {
	interval time.Duration // cron任务执行默认间隔，可通过SetInterval()设置
	timer    *time.Timer
	tasks    map[string]TaskFunc // 定时任务列表，可通过AddTask()添加
	stop     chan interface{}    // 可通过Stop()停止运行cron任务
	sigCh    chan os.Signal
	// wg       sync.WaitGroup
	log    LogFunc
	logErr LogFunc
	newCtx NewContextFunc
}

// TaskFunc func
type TaskFunc func(ctx context.Context) error

// New ...
func New() *Cron {
	c := &Cron{
		interval: defaultInterval,
		stop:     make(chan interface{}, 1),
		tasks:    make(map[string]TaskFunc),
		log:      defaultLogfunc,
		logErr:   defaultLogErrorfunc,
		newCtx:   defaultNewCtx,
	}
	return c
}

// SetNewCtxFunc 创建新context的方法
func (c *Cron) SetNewCtxFunc(f NewContextFunc) {
	c.newCtx = f
}

// SetLogFunc 设置输出普通日志的方法
func (c *Cron) SetLogFunc(f LogFunc) {
	c.log = f
}

// SetLogErrorFunc 设置输出错误日志的方法
func (c *Cron) SetLogErrorFunc(f LogFunc) {
	c.logErr = f
}

// SetInterval 设置cron任务执行间隔，不设置则使用默认值defaultInterval
func (c *Cron) SetInterval(i time.Duration) {
	c.interval = i
}

// AddTask ...
func (c *Cron) AddTask(name string, f TaskFunc) {
	// if _, ok := c.tasks[name]; ok {
	// 	return
	// }
	c.tasks[name] = f
	return
}

// Stop stop cron
func (c *Cron) Stop(somethings ...interface{}) {
	c.stop <- somethings
	return
}

// Run start cron
func (c *Cron) Run() {
	// sigCh不是nil表明已经有Run方法在运行
	if c.sigCh != nil {
		panic(ErrIsRunning)
	}
	// 处理系统信号，捕捉到INT和TERM时调用cron->Stop()停止
	c.sigCh = make(chan os.Signal, 1)
	defer close(c.sigCh)
	signal.Notify(c.sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(c.sigCh)

	go func() {
		select {
		case sig := <-c.sigCh:
			c.log(context.Background(), "_cron_signal_received||interval=%v||signal=%v", c.interval, sig)
			c.Stop(fmt.Errorf("signal_received:%v", sig))
		}
	}()

	for {
		startTime := time.Now()

		ctx := c.newCtx(context.Background())

		next := startTime.Add(c.interval)
		c.timer = time.NewTimer(c.interval)

		c.log(ctx, "_cron_start_run||startTime=%v||next=%v||interval=%v", startTime.Format("2006-01-02 03:04:05.000000"), next.Format("2006-01-02 03:04:05.000000"), c.interval)

		select {
		case <-c.timer.C:
			c.log(ctx, "_cron_time_over||startTime=%v||next=%v||interval=%v", startTime.Format("2006-01-02 03:04:05.000000"), next.Format("2006-01-02 03:04:05.000000"), c.interval)
		case somethings := <-c.stop:
			// c.wg.Wait()
			c.log(ctx, "_cron_stop_run||startTime=%v||next=%v||interval=%v||stop_chan_output=%v", startTime.Format("2006-01-02 03:04:05.000000"), next.Format("2006-01-02 03:04:05.000000"), c.interval, somethings)
			return
		}

		for name, task := range c.tasks {
			// c.wg.Add(1)
			go func(ctx context.Context, name string, task TaskFunc) {
				// defer c.wg.Done()
				ctxChild := c.newCtx(ctx)

				c.log(ctx, "_cron_task_run||task_name=%v||startTime=%v||next=%v||interval=%v", name, startTime.Format("2006-01-02 03:04:05.000000"), next.Format("2006-01-02 03:04:05.000000"), c.interval)

				defer func(ctx context.Context) {
					if err := recover(); err != nil {
						stack := debug.Stack()
						lines := bytes.Split(stack, []byte("\n"))
						stack = bytes.Join(lines[7:], []byte("\n"))

						c.logErr(ctx, "_cron_task_panic||task_name=%s||error=%v||stack:\n%s", name, err, string(stack))
					}
				}(ctxChild)
				err := task(ctxChild)
				c.log(ctx, "_cron_task_done||task_name=%v||startTime=%v||next=%v||interval=%v||error=%v", name, startTime.Format("2006-01-02 03:04:05.000000"), next.Format("2006-01-02 03:04:05.000000"), c.interval, err)
			}(ctx, name, task)
		}
	}
}
