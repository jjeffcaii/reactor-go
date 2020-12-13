package scheduler_test

import (
	"testing"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestIsElastic(t *testing.T) {
	assert.True(t, scheduler.IsElastic(scheduler.Elastic()))
	assert.True(t, scheduler.IsElastic(scheduler.ElasticBounded()))
	assert.False(t, scheduler.IsElastic(scheduler.Parallel()))
	assert.False(t, scheduler.IsElastic(scheduler.Immediate()))
	assert.False(t, scheduler.IsElastic(scheduler.Single()))
}

func TestIsSingle(t *testing.T) {
	assert.True(t, scheduler.IsSingle(scheduler.Single()))
	assert.True(t, scheduler.IsSingle(scheduler.NewSingle()))
	assert.False(t, scheduler.IsSingle(scheduler.Parallel()))
	assert.False(t, scheduler.IsSingle(scheduler.ElasticBounded()))
	assert.False(t, scheduler.IsSingle(scheduler.Elastic()))
	assert.False(t, scheduler.IsSingle(scheduler.Immediate()))
}

func TestIsParallel(t *testing.T) {
	assert.True(t, scheduler.IsParallel(scheduler.Parallel()))
	assert.False(t, scheduler.IsParallel(scheduler.Single()))
	assert.False(t, scheduler.IsParallel(scheduler.Immediate()))
}
