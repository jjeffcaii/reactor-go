package mono

import (
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
)

func TestMonoScheduleOn_SubscribeWith(t *testing.T) {
	newMonoScheduleOn(newMonoJust(1), scheduler.Immediate())
}
