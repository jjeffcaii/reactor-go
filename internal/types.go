package internal

import "github.com/jjeffcaii/reactor-go"

type InheritableRawPublisher interface {
	reactor.RawPublisher
	Parent() reactor.RawPublisher
}

type Item struct {
	V reactor.Any
	E error
}
