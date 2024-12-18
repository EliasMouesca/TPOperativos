package timer

/*type Timer struct {
	initInstant time.Time
	lastInstant time.Time
	duration    time.Duration
	timer       *time.Timer
}

func CreateTimer(duration time.Duration) Timer {
	return Timer{
		duration: duration,
	}
}

func (t Timer) Start() {
	logger.Trace("Starting timer")
	t.initInstant = time.Now()
	t.timer = time.NewTimer(t.duration)
}

func (t Timer) Stop() {
	t.lastInstant = time.Now()
	t.timer.Stop()
	t.timer = nil
}

func (t Timer) Unstop() {
	t.duration = t.lastInstant.Sub(t.initInstant)
	t.timer.Reset(t.duration)
}

func (t Timer) GetChannel() <-chan time.Time {
	if t.timer == nil {
		logger.Error("Timer no inicializado")
		return nil
	}
	return t.timer.C
}


*/
