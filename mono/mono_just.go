package mono

func Just(v interface{}) Mono {
	return New(func(sink Sink) {
		sink.Success(v)
	})
}
