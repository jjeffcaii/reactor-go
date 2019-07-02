package mono

import rs "github.com/jjeffcaii/reactor-go"

type Mono interface {
	rs.Publisher
	Map(rs.Transformer) Mono
}
