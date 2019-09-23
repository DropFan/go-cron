package cron_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	cron "github.com/DropFan/go-cron/cron"
)

func init() {
	initLog()
}

func initLog() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	// log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)

}

func uninit() {
	// log.Close()
}

func TestCron_Run(t *testing.T) {
	// taskOne
	defer uninit()

	tests := []struct {
		name       string
		interval   time.Duration
		tasks      map[string]cron.TaskFunc
		timeout    time.Duration
		ignore     bool
		logFunc    cron.LogFunc
		logErrFunc cron.LogFunc
		ctxFunc    cron.NewContextFunc
	}{
		// TODO: Add test cases.
		{
			name: "1_task",
			tasks: map[string]cron.TaskFunc{
				"test_task": func(ctx context.Context) error {
					t.Logf("test_task_doing...")
					err := errors.New("test task done")
					return err
				},
			},
			interval: time.Second,
			timeout:  time.Second * 2,
			ignore:   true,
		},
		{
			name: "2_tasks",
			tasks: map[string]cron.TaskFunc{
				"task_21": func(ctx context.Context) error {
					t.Logf("task_21...")
					err := errors.New("task_21 done")
					return err
				},
				"task_22": func(ctx context.Context) error {
					t.Logf("task_22...")
					err := errors.New("task_22 done")
					return err
				},
			},
			interval: time.Second,
			timeout:  time.Second * 2,
			ignore:   true,
		},
		{
			name: "3_tasks",
			tasks: map[string]cron.TaskFunc{
				"task_31": func(ctx context.Context) error {
					t.Logf("task_31...")
					err := errors.New("task_31 done")
					return err
				},
				"task_32": func(ctx context.Context) error {
					t.Logf("task_32...")
					// var c *cron.Cron
					// c.Stop("1234567890")
					err := errors.New("task_32 done")
					return err
				},
				"task_33": func(ctx context.Context) error {
					t.Logf("task_33...")
					time.Sleep(time.Second * 3)
					err := errors.New("task_33 done")
					return err
				},
			},
			interval: time.Second,
			timeout:  time.Second * 2,
			// ignore:   true,
			logFunc:    logInfo,
			logErrFunc: logError,
			ctxFunc:    newCtx,
		},
	}

	for _, tt := range tests {
		if tt.ignore {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {

			t.Logf("run...%s", tt.name)

			startTime := time.Now()
			c := cron.New()

			if tt.ctxFunc != nil {
				c.SetNewCtxFunc(tt.ctxFunc)
			}

			if tt.logFunc != nil {
				c.SetLogFunc(tt.logFunc)
			}
			if tt.logErrFunc != nil {
				c.SetLogErrorFunc(tt.logErrFunc)
			}
			c.SetInterval(tt.interval)
			for name, task := range tt.tasks {
				c.AddTask(name, task)
			}
			go c.Run()
			count := 0
			for {
				count++
				t.Logf("wait stop...%s,%v", tt.name, count)
				time.Sleep(time.Second + time.Millisecond*10)
				if time.Now().Sub(startTime).Nanoseconds() >= int64(tt.timeout) {
					// c.Stop(tt.name + "_test_done")
					c.Stop("stop by time over")
					t.Logf("c.Stop(%v)", tt.name+"_test_done")
					break
				}
			}
		})
	}
	time.Sleep(time.Microsecond * 100)
}

// default new context func
func newCtx(ctx context.Context) context.Context {
	return ctx
}

// default log func
func logInfo(ctx context.Context, args ...interface{}) {
	var s string
	if len(args) > 1 {
		if format, ok := args[0].(string); ok {
			s = fmt.Sprintf(format, args[1:]...)
		}
	} else {
		s = fmt.Sprint(args...)
	}
	s = "[Info]" + s
	log.Output(2, s)
}

// default log func
func logError(ctx context.Context, args ...interface{}) {
	var s string
	if len(args) > 1 {
		if format, ok := args[0].(string); ok {
			s = fmt.Sprintf(format, args[1:]...)
		}
	} else {
		s = fmt.Sprint(args...)
	}
	s = "[ERROR]" + s
	log.Output(2, s)
}
