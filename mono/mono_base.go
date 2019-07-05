package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type baseMono struct {
	child Mono
}

func (p *baseMono) SwitchIfEmpty(alternative Mono) Mono {
	return newMonoSwitchIfEmpty(p.child, alternative)
}

func (p *baseMono) Filter(f rs.Predicate) Mono {
	return newMonoFilter(p.child, f)
}

func (p *baseMono) Map(t rs.Transformer) Mono {
	return newMonoMap(p.child, t)
}

func (p *baseMono) FlatMap(mapper flatMapper) Mono {
	return newMonoFlatMap(p.child, mapper)
}

func (p *baseMono) SubscribeOn(sc scheduler.Scheduler) Mono {
	return newMonoScheduleOn(p.child, sc)
}

func (p *baseMono) Block(ctx context.Context) (interface{}, error) {
	return toBlock(ctx, p.child)
}

func (p *baseMono) DoOnNext(fn rs.FnOnNext) Mono {
	return newMonoPeek(p.child, peekNext(fn))
}

func (p *baseMono) DoFinally(fn rs.FnOnFinally) Mono {
	return newMonoDoFinally(p.child, fn)
}
