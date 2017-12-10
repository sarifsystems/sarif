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

func New(d time.Duration) *Failsafe {
	f := &Failsafe{
		duration: d,
		deadline: time.Now(),
		checkins: make(chan time.Time, 3),
		confirms: make(chan bool, 3),
	}

	return f
}

func (f *Failsafe) advanceDeadline(t time.Time) {
	if t.After(f.deadline.Add(f.duration / 2)) {
		// If check-in is already half a period late, set new deadline
		// a whole period later than the check-in.
		f.deadline = t.Add(f.duration)
	} else if t.After(f.deadline.Add(-f.duration / 2)) {
		// If check-in is not more than half a period early, extend deadline
		f.deadline = f.deadline.Add(f.duration)
	}
}

func (f *Failsafe) Run() {
	for {
		stage := f.nextStage()
		if stage == nil {
			t := <-f.checkins
			f.advanceDeadline(t)
			f.confirms <- true
			continue
		}

		stageTime := f.deadline.Add(stage.Duration)
		timeout := time.After(stageTime.Sub(time.Now()))

		select {
		case t := <-f.checkins:
			f.advanceDeadline(t)
			f.confirms <- true
		case <-timeout:
			go stage.Func()
		}
	}
}

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
