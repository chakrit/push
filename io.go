package push

type IO struct {
	Input  chan *Session
	Output chan *Feedback
}

func NewIO() *IO {
	return &IO{
		Input:  make(chan *Session),
		Output: make(chan *Feedback),
	}
}

func (io *IO) Close() {
	close(io.Input)
	close(io.Output)
}
