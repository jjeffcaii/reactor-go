package mono

import (
	"github.com/jjeffcaii/reactor-go"
)

type Sink interface {
	Success(v interface{})
	Error(e error)
}

type Mono interface {
	rs.Publisher
	Map(fn rs.FnTransform) Mono
	MapInt(transform func(int) interface{}) Mono
	MapString(transform func(string) interface{}) Mono
	SubscribeOn(s rs.Scheduler) Mono
	PublishOn(s rs.Scheduler) Mono
}
