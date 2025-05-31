package perf

import (
	"fmt"
	"time"
)

type Step struct {
	name     string
	duration time.Duration
}

type Profiler struct {
	name     string
	lastStep time.Time
	steps    []Step
}

func Profile(name string) *Profiler {
	return &Profiler{
		name:     name,
		lastStep: time.Now(),
	}
}

func (p *Profiler) Step(name string) {
	duration := time.Since(p.lastStep)

	p.steps = append(p.steps, Step{
		name:     name,
		duration: duration,
	})

	p.lastStep = time.Now()
}

func (p *Profiler) Print() {
	for _, step := range p.steps {
		fmt.Printf("%s.%s: %.4fms\n", p.name, step.name, float64(step.duration.Nanoseconds())/1000000)
	}
}
