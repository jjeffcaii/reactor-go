package reactor

type Processor interface {
	Publisher
	Subscriber
}
