package rs

type opSubType int8

const (
	opOnNext opSubType = iota
	opOnSubscribe
	opOnError
	opOnComplete
)

type subscriberBuilder struct {
	m map[opSubType][]interface{}
}

func (p *subscriberBuilder) OnSubscribe(s Subscription) {
	found, ok := p.m[opOnSubscribe]
	if !ok {
		return
	}
	for _, fn := range found {
		fn.(func(Subscription))(s)
	}
}

func (p *subscriberBuilder) OnNext(v interface{}) {
	found, ok := p.m[opOnNext]
	if !ok {
		return
	}
	for _, fn := range found {
		fn.(func(interface{}))(v)
	}
}

func (p *subscriberBuilder) OnComplete() {
	found, ok := p.m[opOnComplete]
	if !ok {
		return
	}
	for _, fn := range found {
		fn.(func())()
	}
}

func (p *subscriberBuilder) OnError(err error) {
	found, ok := p.m[opOnError]
	if !ok {
		return
	}
	for _, fn := range found {
		fn.(func(error))(err)
	}
}

type OpSubscriber func(*subscriberBuilder)

func OnNext(fn func(v interface{})) OpSubscriber {
	return func(builder *subscriberBuilder) {
		builder.m[opOnNext] = append(builder.m[opOnNext], fn)
	}
}

func OnComplete(fn func()) OpSubscriber {
	return func(builder *subscriberBuilder) {
		builder.m[opOnComplete] = append(builder.m[opOnComplete], fn)
	}
}

func OnSubscribe(fn func(Subscription)) OpSubscriber {
	return func(builder *subscriberBuilder) {
		builder.m[opOnSubscribe] = append(builder.m[opOnSubscribe], fn)
	}
}
