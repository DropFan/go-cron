package cron

import (
	"context"
	"fmt"
	"log"
)

// LogFunc ...
type LogFunc func(ctx context.Context, args ...interface{})

// NewContextFunc ...
type NewContextFunc func(ctx context.Context) context.Context

// default new context func
func defaultNewCtx(ctx context.Context) context.Context {
	return ctx
}

// default log func
func defaultLogfunc(ctx context.Context, args ...interface{}) {
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

// default error log func
func defaultLogErrorfunc(ctx context.Context, args ...interface{}) {
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
