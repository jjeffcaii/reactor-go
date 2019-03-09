package rs

type OpSubscriber func(*hooks)

func OnRequest(fn FnOnRequest) OpSubscriber {
	return func(h *hooks) {
		h.DoOnRequest(fn)
	}
}

func OnNext(fn FnOnNext) OpSubscriber {
	return func(h *hooks) {
		h.DoOnNext(fn)
	}
}

func OnComplete(fn FnOnComplete) OpSubscriber {
	return func(h *hooks) {
		h.DoOnComplete(fn)
	}
}

func OnSubscribe(fn FnOnSubscribe) OpSubscriber {
	return func(h *hooks) {
		h.DoOnSubscribe(fn)
	}
}

func OnError(fn FnOnError) OpSubscriber {
	return func(h *hooks) {
		h.DoOnError(fn)
	}
}
