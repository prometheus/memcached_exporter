package prometheus

import (
	"errors"
	"math"
)

type IntervalCounter interface {
	Counter
}

func NewIntervalCounter(opts CounterOpts) Counter {
	desc := NewDesc(
		BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)
	result := &intervalCounter{value: value{desc: desc, valType: CounterValue, labelPairs: desc.constLabelPairs}}
	result.Init(result) // Init self-collection.
	return result
}

type intervalCounter struct {
	value
}

func (c *intervalCounter) Add(v float64) {
	if v < 0 {
		panic(errors.New("interval counter cannot decrease in value"))
	}
	c.value.Add(v)
}

func (c *intervalCounter) Collect(ch chan<- Metric) {
	ch <- c.self
	c.value.valBits = math.Float64bits(0)
}
