package failsafe

import "time"

type stage struct {
	Duration time.Duration
	Func     func()
}

// Failsafe represents a dead man's switch
type Failsafe struct {
	duration time.Duration

	failures int
	deadline time.Time
	checkins chan time.Time
	confirms chan bool

	stages []stage
}

// New test
func New(d time.Duration) *Failsafe {
	f := &Failsafe{
		duration: d,
		deadline: time.Now().Add(d),
		checkins: make(chan time.Time, 3),
		confirms: make(chan bool, 3),
	}

	return f
}

// Run test
func (f *Failsafe) Run() {
	for {
		stage := f.nextStage()
		if stage == nil {
			t := <-f.checkins
			f.deadline = t.Add(f.duration)
			continue
		}

		stageTime := f.deadline.Add(stage.Duration)
		timeout := time.After(stageTime.Sub(time.Now()))

		select {
		case t := <-f.checkins:
			f.deadline = t.Add(f.duration)
			f.confirms <- true
		case <-timeout:
			go stage.Func()
		}
	}
}

// After test
func (f *Failsafe) After(d time.Duration, t func()) {
	f.stages = append(f.stages, stage{d, t})
}

func (f *Failsafe) nextStage() *stage {
	var earliest *stage
	now := time.Now()
	for _, s := range f.stages {
		if now.After(f.deadline.Add(s.Duration)) {
			continue
		}
		if earliest == nil || s.Duration < earliest.Duration {
			w := s
			earliest = &w
		}
	}

	return earliest
}

// CheckIn prevents triggering the failsafe and extends the deadline
func (f *Failsafe) CheckIn() time.Time {
	f.checkins <- time.Now()
	<-f.confirms
	return f.deadline
}

func (f *Failsafe) Deadline() time.Time {
	return f.deadline
}
