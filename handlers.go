package rs

type OpSubscriber func(*Hooks)

func OnRequest(fn FnOnRequest) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnRequest(fn)
	}
}

func OnNext(fn FnOnNext) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnNext(fn)
	}
}

func OnComplete(fn FnOnComplete) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnComplete(fn)
	}
}

func OnSubscribe(fn FnOnSubscribe) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnSubscribe(fn)
	}
}

func OnError(fn FnOnError) OpSubscriber {
	return func(h *Hooks) {
		h.DoOnError(fn)
	}
}
