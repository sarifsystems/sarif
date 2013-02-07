package pipe

import (
	"github.com/xconstruct/stark"
)

type Pipe struct {
	in chan *stark.Message
	out chan *stark.Message
}

func (p *Pipe) Read() (*stark.Message, error) {
	return <-p.in, nil
}

func (p *Pipe) Write(msg *stark.Message) error {
	p.out <- msg
	return nil
}

func New() (*Pipe, *Pipe) {
	left := &Pipe{
		make(chan *stark.Message, 10),
		make(chan *stark.Message, 10),
	}
	right := &Pipe{left.out, left.in}
	return left, right
}
