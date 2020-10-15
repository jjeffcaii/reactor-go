package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func IsSubscribeAsync(m Mono) bool {
	var publisher reactor.RawPublisher
	switch t := m.(type) {
	case wrapper:
		publisher = t.RawPublisher
	case *oneshotWrapper:
		publisher = t.RawPublisher
	default:
		return false
	}

	var sc scheduler.Scheduler
	switch pub := publisher.(type) {
	case monoScheduleOn:
		sc = pub.sc
	case *monoScheduleOn:
		sc = pub.sc
	default:
		return false
	}
	return scheduler.IsParallel(sc) || scheduler.IsElastic(sc) || scheduler.IsSingle(sc)
}

func block(ctx context.Context, publisher reactor.RawPublisher) (Any, error) {
	done := make(chan struct{})
	vchan := make(chan reactor.Any, 1)
	echan := make(chan error, 1)
	b := subscribers.NewBlockSubscriber(done, vchan, echan)
	publisher.SubscribeWith(ctx, b)
	<-done

	defer close(vchan)
	defer close(echan)

	select {
	case value := <-vchan:
		return value, nil
	case err := <-echan:
		return nil, err
	default:
		return nil, nil
	}
}

func mustProcessor(publisher reactor.RawPublisher) *processor {
	pp, ok := publisher.(*processor)
	if !ok {
		panic(errNotProcessor)
	}
	return pp
}
