package router

import (
	"log"
	"github.com/xconstruct/stark"
)

type Conn interface {
	Read() (*stark.Message, error)
	Write(*stark.Message) error
}

type LoggedConn struct {
	Conn
}

func (c *LoggedConn) Write(msg *stark.Message) error {
	log.Printf(" --> %v\n", msg)
	return c.Conn.Write(msg)
}

func (c *LoggedConn) Read() (*stark.Message, error) {
	msg, err := c.Conn.Read()
	log.Printf(" <-- %v\n", msg)
	return msg, err
}

type Router struct {
	Name string
	Conns map[string]Conn
}

func NewRouter(name string) *Router {
	return &Router{
		name,
		make(map[string]Conn),
	}
}

func (r *Router) Write(msg *stark.Message) error {
	path := stark.GetPath(msg)

	log.Println(msg)

	next := path.Next()
	if r.Conns[next] != nil {
		if err := r.Conns[next].Write(msg); err != nil {
			return err
		}
		return nil
	}

	log.Fatalf("router/write: destination not found: %v\n", next)
	return nil
}

func (r *Router) Connect(name string, conn Conn) {
	r.Conns[name] = conn
	go func() {
		for {
			msg, err := conn.Read()
			if err != nil {
				log.Fatalf("router/connect: %v\n", err)
				return
			}
			r.Write(msg)
		}
	}()
}
