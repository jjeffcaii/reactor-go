package flux

type OverflowStrategy int8

const (
	OverflowBuffer OverflowStrategy = iota
	OverflowIgnore
	OverflowError
	OverflowDrop
	OverflowLatest
)

func Just(first interface{}, others ...interface{}) Flux {
	var values []interface{}
	values = append(values, first)
	values = append(values, others...)
	return newSliceFlux(values)
}
